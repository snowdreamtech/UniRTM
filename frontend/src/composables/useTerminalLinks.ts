// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import type { Terminal, ILinkProvider, ILink, IBufferRange } from "@xterm/xterm";
import { useWorkspaceBus } from "./useWorkspaceBus";

export interface ParsedTerminalLink {
  type: "url" | "file";
  text: string;
  startIndex: number; // 0-indexed in line
  endIndex: number; // 0-indexed exclusive
  url?: string;
  filePath?: string;
  line?: number;
  column?: number;
}

// 匹配 Web URL
const URL_REGEX = /https?:\/\/[^\s"'`()<>[\]{}]+/gi;

// 匹配 Python 堆栈: File "xxx", line 123
const PYTHON_STACK_REGEX = /File\s+"([^"]+)",\s+line\s+(\d+)/gi;

// 匹配标准文件路径以及附带的行号/列号 (如 src/App.vue:42:15, ./main.go:88, /home/user/project/file.ts:10)
// 支持常见扩展名
const FILE_PATH_EXT_REGEX =
  /(?:^|[\s([`'"<>=])((?:[a-zA-Z]:[\\/]|(?:\.{1,2}[\\/])|(?:[\w.-]+[\\/]))*[\w.-]+\.(?:ts|js|vue|jsx|tsx|json|md|html|css|scss|sh|go|py|rs|yaml|yml|toml|sql|txt|c|cpp|h|java|php|rb|xml|env))(?::(\d+))?(?::(\d+))?/gi;

/**
 * 解析终端单行字符中的所有可点击超链接（URL 与 文件路径/堆栈）
 */
export function parseTerminalLine(lineText: string): ParsedTerminalLink[] {
  if (!lineText) return [];

  const results: ParsedTerminalLink[] = [];
  const occupiedRanges: [number, number][] = [];

  function isOverlapping(start: number, end: number): boolean {
    return occupiedRanges.some(([s, e]) => Math.max(s, start) < Math.min(e, end));
  }

  // 1. 匹配 Web URL
  let urlMatch: RegExpExecArray | null;
  URL_REGEX.lastIndex = 0;
  while ((urlMatch = URL_REGEX.exec(lineText)) !== null) {
    const rawUrl = urlMatch[0];
    const startIndex = urlMatch.index;

    // 清理可能误吸入的末尾符号（如句点、逗号、括号）
    const cleanedUrl = rawUrl.replace(/[.,;:!?)]+$/, "");
    const finalEndIndex = startIndex + cleanedUrl.length;

    if (!isOverlapping(startIndex, finalEndIndex)) {
      occupiedRanges.push([startIndex, finalEndIndex]);
      results.push({
        type: "url",
        text: cleanedUrl,
        startIndex,
        endIndex: finalEndIndex,
        url: cleanedUrl,
      });
    }
  }

  // 2. 匹配 Python 堆栈
  let pyMatch: RegExpExecArray | null;
  PYTHON_STACK_REGEX.lastIndex = 0;
  while ((pyMatch = PYTHON_STACK_REGEX.exec(lineText)) !== null) {
    const filePath = pyMatch[1];
    const lineNum = parseInt(pyMatch[2], 10);
    const startIndex = pyMatch.index;
    const endIndex = startIndex + pyMatch[0].length;

    if (!isOverlapping(startIndex, endIndex)) {
      occupiedRanges.push([startIndex, endIndex]);
      results.push({
        type: "file",
        text: pyMatch[0],
        startIndex,
        endIndex,
        filePath,
        line: lineNum,
        column: 1,
      });
    }
  }

  // 3. 匹配常规文件路径加行列号
  let fileMatch: RegExpExecArray | null;
  FILE_PATH_EXT_REGEX.lastIndex = 0;
  while ((fileMatch = FILE_PATH_EXT_REGEX.exec(lineText)) !== null) {
    const fullMatchedStr = fileMatch[0];
    const rawFilePath = fileMatch[1];
    const lineStr = fileMatch[2];
    const colStr = fileMatch[3];

    // 计算实际匹配文本在 lineText 中的起始索引（因为前导可能有空白或括号字符）
    const offset = fullMatchedStr.indexOf(rawFilePath);
    const startIndex = fileMatch.index + offset;
    const matchedTokenLen =
      rawFilePath.length + (lineStr ? `:${lineStr}`.length : 0) + (colStr ? `:${colStr}`.length : 0);
    const endIndex = startIndex + matchedTokenLen;

    if (!isOverlapping(startIndex, endIndex)) {
      occupiedRanges.push([startIndex, endIndex]);
      results.push({
        type: "file",
        text: lineText.slice(startIndex, endIndex),
        startIndex,
        endIndex,
        filePath: rawFilePath.replace(/^\.\//, ""), // 清洗开头的 ./
        line: lineStr ? parseInt(lineStr, 10) : undefined,
        column: colStr ? parseInt(colStr, 10) : undefined,
      });
    }
  }

  // 按起始索引升序排列
  return results.sort((a, b) => a.startIndex - b.startIndex);
}

export interface TerminalLinkProviderOptions {
  onFileClick?: (path: string, line?: number, col?: number) => void;
  onUrlClick?: (url: string) => void;
}

/**
 * 为 xterm 实例构建并注册自定义超链接提供者
 */
export function createTerminalLinkProvider(term: Terminal, options?: TerminalLinkProviderOptions): ILinkProvider {
  const { openInEditor } = useWorkspaceBus();

  return {
    provideLinks(bufferLineNumber: number, callback: (links: ILink[] | undefined) => void): void {
      try {
        const line = term.buffer.active.getLine(bufferLineNumber - 1);
        if (!line) {
          callback(undefined);
          return;
        }

        const lineText = line.translateToString(true);
        const parsedLinks = parseTerminalLine(lineText);

        if (parsedLinks.length === 0) {
          callback(undefined);
          return;
        }

        const links: ILink[] = parsedLinks.map((item) => {
          const range: IBufferRange = {
            start: {
              x: item.startIndex + 1, // xterm 是 1-indexed
              y: bufferLineNumber,
            },
            end: {
              x: item.endIndex,
              y: bufferLineNumber,
            },
          };

          return {
            range,
            text: item.text,
            decorations: {
              pointerCursor: true,
              underline: true,
            },
            activate(_evt: MouseEvent, _text: string) {
              if (item.type === "url" && item.url) {
                if (options?.onUrlClick) {
                  options.onUrlClick(item.url);
                } else if (typeof window !== "undefined") {
                  window.open(item.url, "_blank", "noopener,noreferrer");
                }
              } else if (item.type === "file" && item.filePath) {
                if (options?.onFileClick) {
                  options.onFileClick(item.filePath, item.line, item.column);
                } else {
                  openInEditor({
                    path: item.filePath,
                    line: item.line,
                    column: item.column,
                  });
                }
              }
            },
          };
        });

        callback(links);
      } catch (err) {
        console.error("[useTerminalLinks] provideLinks error:", err);
        callback(undefined);
      }
    },
  };
}

export function useTerminalLinks() {
  return {
    parseTerminalLine,
    createTerminalLinkProvider,
  };
}
