// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, beforeEach } from "vitest";
import { translateMonacoLabel } from "./monacoLocalization";
import { currentLocale } from "../../i18n";

describe("monacoLocalization", () => {
  beforeEach(() => {
    currentLocale.value = "zh-CN";
  });

  it("translates Monaco context menu labels to standard Chinese", () => {
    expect(translateMonacoLabel("Change All Occurrences")).toBe("更改所有匹配项");
    expect(translateMonacoLabel("Cut")).toBe("剪切");
    expect(translateMonacoLabel("Copy")).toBe("复制");
    expect(translateMonacoLabel("Paste")).toBe("粘贴");
    expect(translateMonacoLabel("Command Palette")).toBe("命令面板...");
    expect(translateMonacoLabel("Format Document")).toBe("格式化文档");
    expect(translateMonacoLabel("Refactor...")).toBe("重构...");
    expect(translateMonacoLabel("Rename Symbol")).toBe("重命名符号");
  });

  it("retains original label for English locale", () => {
    currentLocale.value = "en-US";
    expect(translateMonacoLabel("Cut")).toBe("Cut");
    expect(translateMonacoLabel("Command Palette")).toBe("Command Palette");
  });

  it("supports Traditional Chinese (zh-TW)", () => {
    currentLocale.value = "zh-TW";
    expect(translateMonacoLabel("Cut")).toBe("剪下");
    expect(translateMonacoLabel("Copy")).toBe("複製");
    expect(translateMonacoLabel("Paste")).toBe("貼上");
  });
});
