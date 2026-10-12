/**
 * Format log timestamp to HH:MM:SS.mmm format
 */
export function formatLogTime(ts: string | Date | number): string {
  if (!ts) return "";
  const date = new Date(ts);
  if (isNaN(date.getTime())) return String(ts);

  const hours = date.getHours().toString().padStart(2, "0");
  const minutes = date.getMinutes().toString().padStart(2, "0");
  const seconds = date.getSeconds().toString().padStart(2, "0");
  const ms = date.getMilliseconds().toString().padStart(3, "0");

  return `${hours}:${minutes}:${seconds}.${ms}`;
}

/**
 * Format logs array to plain text format
 */
export interface LogItem {
  timestamp: string | Date;
  level: string;
  message: string;
  details?: string;
}

export function formatLogsToText(logs: LogItem[]): string {
  return logs
    .map((l) => `[${formatLogTime(l.timestamp)}] [${l.level}] ${l.message}${l.details ? " - " + l.details : ""}`)
    .join("\n");
}
