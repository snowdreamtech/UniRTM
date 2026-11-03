// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/**
 * Lightweight, zero-dependency Markdown renderer designed for extension READMEs and changelogs.
 * Enforces XSS sanitization, expands collapsible details/screenshots by default,
 * and ensures responsive image sizing without container overflow.
 */

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

function sanitizeUrl(rawUrl: string): string {
  const trimmed = rawUrl.trim();
  if (/^(https?:\/\/|data:image\/|\/|\.\/|\.\.\/)/i.test(trimmed)) {
    return trimmed;
  }
  return "#";
}

function sanitizeHtmlAttributes(attrStr: string): string {
  // Strip on* event handlers (e.g. onclick, onerror)
  let clean = attrStr.replace(/\s*on\w+\s*=\s*(["'][^"']*["']|[^\s>]+)/gi, "");
  // Strip javascript: uris
  clean = clean.replace(/javascript\s*:/gi, "forbidden:");
  return clean;
}

function parseInline(inlineText: string): string {
  let res = inlineText;

  // Inline images: ![alt](url)
  res = res.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_m, alt, url) => {
    const safeUrl = sanitizeUrl(url);
    const safeAlt = escapeHtml(alt);
    return `<img class="md-img" src="${safeUrl}" alt="${safeAlt}" loading="lazy" />`;
  });

  // Links: [text](url)
  res = res.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_m, txt, url) => {
    const safeUrl = sanitizeUrl(url);
    return `<a class="md-link" href="${safeUrl}" target="_blank" rel="noopener noreferrer">${parseInline(txt)}</a>`;
  });

  // Inline code: `code`
  res = res.replace(/`([^`]+)`/g, (_m, code) => {
    return `<code class="md-inline-code">${escapeHtml(code)}</code>`;
  });

  // Bold & Italic: ***text*** or ___text___
  res = res.replace(/(\*\*\*|___)(.*?)\1/g, "<strong><em>$2</em></strong>");

  // Bold: **text** or __text__
  res = res.replace(/(\*\*|__)(.*?)\1/g, "<strong>$2</strong>");

  // Italic: *text* or _text_
  res = res.replace(/(\*|_)(.*?)\1/g, "<em>$2</em>");

  // Strikethrough: ~~text~~
  res = res.replace(/~~(.*?)~~/g, "<del>$1</del>");

  return res;
}

/**
 * Preprocesses raw HTML tags in Markdown:
 * 1. Automatically injects 'open' into <details> so images & previews are NOT collapsed by default.
 * 2. Rewrites raw <img> tags to add class="md-img" and loading="lazy", ensuring container fit.
 * 3. Removes malicious scripts or unsafe objects.
 */
function preprocessHtmlTags(content: string): string {
  let res = content;

  // Remove dangerous tags completely
  res = res.replace(/<\s*(script|iframe|object|embed|style)\b[^>]*>[\s\S]*?<\s*\/\s*\1\s*>/gi, "");
  res = res.replace(/<\s*(script|iframe|object|embed|style)\b[^>]*>/gi, "");

  // Ensure <details> is open by default so screenshots and icon grids are immediately visible
  res = res.replace(/<details\b(?![^>]*\bopen\b)([^>]*)>/gi, '<details open class="md-details"$1>');
  res = res.replace(/<details\b([^>]*)>/gi, (m) => {
    if (!m.includes("class=")) {
      return m.replace("<details", '<details class="md-details"');
    }
    return m;
  });

  // Summary tag styling
  res = res.replace(/<summary\b([^>]*)>/gi, '<summary class="md-summary"$1>');

  // Raw <img> tag enhancement: ensure loading="lazy", responsive class, sanitized src
  res = res.replace(/<img\s+([^>]*?)src=["']([^"']+)["']([^>]*?)>/gi, (_m, pre, src, post) => {
    const safeSrc = sanitizeUrl(src);
    const cleanPre = sanitizeHtmlAttributes(pre);
    const cleanPost = sanitizeHtmlAttributes(post);
    return `<img class="md-img" src="${safeSrc}" loading="lazy" ${cleanPre} ${cleanPost} />`;
  });

  return res;
}

export function renderMarkdown(markdown: string): string {
  if (!markdown) return "";

  // Preprocess HTML tags first (details open, image class, xss cleanup)
  const preprocessed = preprocessHtmlTags(markdown);
  const lines = preprocessed.replace(/\r\n/g, "\n").replace(/\r/g, "\n").split("\n");
  const out: string[] = [];

  let inCodeBlock = false;
  let codeLang = "";
  let codeBuffer: string[] = [];

  let inList = false;
  let listType: "ul" | "ol" = "ul";

  let inTable = false;
  let tableBuffer: string[] = [];

  const flushTable = () => {
    if (tableBuffer.length === 0) return;
    let html = '<div class="md-table-wrap"><table class="md-table">';
    let isHeader = true;

    for (const row of tableBuffer) {
      const trimmed = row.trim();
      if (/^\|?(\s*:?-+:?\s*\|?)+$/.test(trimmed)) {
        // Separator row
        isHeader = false;
        continue;
      }
      const cells = trimmed
        .replace(/^\|/, "")
        .replace(/\|$/, "")
        .split("|")
        .map((c) => c.trim());

      const tag = isHeader ? "th" : "td";
      html += "<tr>";
      for (const c of cells) {
        html += `<${tag}>${parseInline(c)}</${tag}>`;
      }
      html += "</tr>";
    }

    html += "</table></div>";
    out.push(html);
    tableBuffer = [];
    inTable = false;
  };

  const flushList = () => {
    if (inList) {
      out.push(listType === "ul" ? "</ul>" : "</ol>");
      inList = false;
    }
  };

  // Block-level HTML tags that should NOT be wrapped in <p>
  const blockHtmlRegex =
    /^<\/?(details|summary|p|div|section|article|h[1-6]|ul|ol|li|table|thead|tbody|tr|th|td|blockquote|pre|br|hr)\b/i;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    // Fenced Code Blocks
    if (line.trim().startsWith("```")) {
      if (!inCodeBlock) {
        flushList();
        flushTable();
        inCodeBlock = true;
        codeLang = line.trim().slice(3).trim();
        codeBuffer = [];
      } else {
        inCodeBlock = false;
        const codeContent = escapeHtml(codeBuffer.join("\n"));
        out.push(`<pre class="md-code-block" data-lang="${escapeHtml(codeLang)}"><code>${codeContent}</code></pre>`);
        codeBuffer = [];
      }
      continue;
    }

    if (inCodeBlock) {
      codeBuffer.push(line);
      continue;
    }

    const trimmed = line.trim();

    // Table row detection
    if (trimmed.startsWith("|") && trimmed.includes("|", 1)) {
      flushList();
      inTable = true;
      tableBuffer.push(trimmed);
      continue;
    } else if (inTable) {
      flushTable();
    }

    // Blank line
    if (!trimmed) {
      flushList();
      continue;
    }

    // Horizontal Rule
    if (/^(-{3,}|\*{3,}|_{3,})$/.test(trimmed)) {
      flushList();
      out.push('<hr class="md-hr" />');
      continue;
    }

    // Headings (# Heading)
    const headingMatch = line.match(/^(#{1,6})\s+(.*)$/);
    if (headingMatch) {
      flushList();
      const level = headingMatch[1].length;
      const text = parseInline(headingMatch[2]);
      out.push(`<h${level} class="md-heading md-h${level}">${text}</h${level}>`);
      continue;
    }

    // Blockquote (> Quote)
    if (line.startsWith(">")) {
      flushList();
      const quoteText = parseInline(line.replace(/^>\s?/, ""));
      out.push(`<blockquote class="md-blockquote">${quoteText}</blockquote>`);
      continue;
    }

    // Unordered List (- or *)
    const ulMatch = line.match(/^(\s*)[-*+]\s+(.*)$/);
    if (ulMatch) {
      if (!inList || listType !== "ul") {
        flushList();
        out.push('<ul class="md-ul">');
        inList = true;
        listType = "ul";
      }
      out.push(`<li class="md-li">${parseInline(ulMatch[2])}</li>`);
      continue;
    }

    // Ordered List (1. )
    const olMatch = line.match(/^(\s*)\d+\.\s+(.*)$/);
    if (olMatch) {
      if (!inList || listType !== "ol") {
        flushList();
        out.push('<ol class="md-ol">');
        inList = true;
        listType = "ol";
      }
      out.push(`<li class="md-li">${parseInline(olMatch[2])}</li>`);
      continue;
    }

    // Block-level HTML passthrough (don't wrap in <p>)
    if (blockHtmlRegex.test(trimmed)) {
      flushList();
      out.push(parseInline(line));
      continue;
    }

    // Regular paragraph
    flushList();
    out.push(`<p class="md-p">${parseInline(line)}</p>`);
  }

  flushList();
  flushTable();

  return out.join("\n");
}
