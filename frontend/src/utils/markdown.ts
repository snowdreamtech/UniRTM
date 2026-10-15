// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/**
 * Lightweight, zero-dependency Markdown renderer designed for extension READMEs and changelogs.
 * Enforces XSS sanitization and handles common markdown constructs (headings, code blocks,
 * images, links, tables, lists, quotes, emphasis).
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
  res = res.replace(/(\*\*\*|___)(.*?)\1/g, '<strong><em>$2</em></strong>');

  // Bold: **text** or __text__
  res = res.replace(/(\*\*|__)(.*?)\1/g, '<strong>$2</strong>');

  // Italic: *text* or _text_
  res = res.replace(/(\*|_)(.*?)\1/g, '<em>$2</em>');

  // Strikethrough: ~~text~~
  res = res.replace(/~~(.*?)~~/g, '<del>$1</del>');

  return res;
}

export function renderMarkdown(markdown: string): string {
  if (!markdown) return "";

  const lines = markdown.replace(/\r\n/g, "\n").replace(/\r/g, "\n").split("\n");
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
        out.push(
          `<pre class="md-code-block" data-lang="${escapeHtml(codeLang)}"><code>${codeContent}</code></pre>`
        );
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

    // Headings
    const headingMatch = line.match(/^(#{1,6})\s+(.*)$/);
    if (headingMatch) {
      flushList();
      const level = headingMatch[1].length;
      const text = parseInline(headingMatch[2]);
      out.push(`<h${level} class="md-heading md-h${level}">${text}</h${level}>`);
      continue;
    }

    // Blockquote
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

    // Regular paragraph
    flushList();
    out.push(`<p class="md-p">${parseInline(line)}</p>`);
  }

  flushList();
  flushTable();

  return out.join("\n");
}
