// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

const MAX_BUFFER_CHARS = 300_000; // 约 5000 行
const STORAGE_PREFIX = "unirtm_term_buf_";

const memoryBuffers = new Map<string, string>();

/**
 * 清除 ANSI 颜色与光标控制字符，输出干净易读的文本
 */
export function stripAnsiCodes(text: string): string {
  if (!text) return "";
  // 匹配标准 ANSI escape sequences
  const ansiRegex = /[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g;
  return text.replace(ansiRegex, "");
}

/**
 * useTerminalPersistence: 终端历史会话持久化与日志导出管理
 */
export function useTerminalPersistence() {
  /**
   * 将终端输出追加到指定 pane 的持久化缓冲区
   */
  function appendChunk(paneId: string, chunk: string): void {
    if (!paneId || !chunk) return;

    let current = memoryBuffers.get(paneId) || "";
    // 如果内存中没有，尝试从 sessionStorage 恢复一次
    if (!current && typeof sessionStorage !== "undefined") {
      try {
        current = sessionStorage.getItem(`${STORAGE_PREFIX}${paneId}`) || "";
      } catch {}
    }

    current += chunk;

    // 滑动窗口控制最大字符长度
    if (current.length > MAX_BUFFER_CHARS) {
      current = current.slice(current.length - MAX_BUFFER_CHARS);
    }

    memoryBuffers.set(paneId, current);

    if (typeof sessionStorage !== "undefined") {
      try {
        sessionStorage.setItem(`${STORAGE_PREFIX}${paneId}`, current);
      } catch {
        // 当达到存储配额时静默降级为仅内存持有
      }
    }
  }

  /**
   * 获取指定终端 pane 的全部历史内容
   */
  function getBuffer(paneId: string): string {
    if (!paneId) return "";
    let buf = memoryBuffers.get(paneId);
    if (buf !== undefined) return buf;

    if (typeof sessionStorage !== "undefined") {
      try {
        buf = sessionStorage.getItem(`${STORAGE_PREFIX}${paneId}`) || "";
        memoryBuffers.set(paneId, buf);
        return buf;
      } catch {}
    }
    return "";
  }

  /**
   * 判断指定终端是否有持久化内容
   */
  function hasBuffer(paneId: string): boolean {
    return !!getBuffer(paneId);
  }

  /**
   * 清空指定终端的持久化缓冲区
   */
  function clearBuffer(paneId: string): void {
    if (!paneId) return;
    memoryBuffers.delete(paneId);
    if (typeof sessionStorage !== "undefined") {
      try {
        sessionStorage.removeItem(`${STORAGE_PREFIX}${paneId}`);
      } catch {}
    }
  }

  /**
   * 导出终端日志为 .log 文件
   */
  function exportLog(paneId: string, title = "terminal", stripAnsi = true): boolean {
    const rawContent = getBuffer(paneId);
    if (!rawContent) return false;

    const outputContent = stripAnsi ? stripAnsiCodes(rawContent) : rawContent;
    const now = new Date();
    const timestamp = `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, "0")}${String(now.getDate()).padStart(2, "0")}-${String(now.getHours()).padStart(2, "0")}${String(now.getMinutes()).padStart(2, "0")}${String(now.getSeconds()).padStart(2, "0")}`;
    const filename = `unirtm-${title.toLowerCase().replace(/[^a-z0-9_-]/g, "_")}-${timestamp}.log`;

    try {
      if (typeof window !== "undefined" && typeof document !== "undefined") {
        const blob = new Blob([outputContent], { type: "text/plain;charset=utf-8" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        return true;
      }
    } catch (err) {
      console.error("[useTerminalPersistence] Export log failed:", err);
    }
    return false;
  }

  function _clearAllForTest(): void {
    memoryBuffers.clear();
    if (typeof sessionStorage !== "undefined") {
      try {
        const keysToRemove: string[] = [];
        for (let i = 0; i < sessionStorage.length; i++) {
          const k = sessionStorage.key(i);
          if (k && k.startsWith(STORAGE_PREFIX)) {
            keysToRemove.push(k);
          }
        }
        keysToRemove.forEach((k) => sessionStorage.removeItem(k));
      } catch {}
    }
  }

  return {
    appendChunk,
    getBuffer,
    hasBuffer,
    clearBuffer,
    exportLog,
    _clearAllForTest,
  };
}
