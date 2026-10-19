// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, beforeEach } from "vitest";
import {
  translateMonacoLabel,
  localizeEditorActions,
} from "./monacoLocalization";
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
    expect(translateMonacoLabel("Select All")).toBe("全选");
    expect(translateMonacoLabel("Undo")).toBe("撤销");
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

  it("localizes editor action objects directly", () => {
    const fakeActions = [
      { id: "editor.action.clipboardCutAction", label: "Cut" },
      { id: "editor.action.changeAll", label: "Change All Occurrences" },
    ];
    const fakeEditor = {
      getSupportedActions: () => fakeActions,
    };

    localizeEditorActions(fakeEditor);
    expect(fakeActions[0].label).toBe("剪切");
    expect(fakeActions[1].label).toBe("更改所有匹配项");
  });

  it("translates repeated context menu items without reverting to English", () => {
    // 模拟第一次弹出菜单项目
    const fakeMenu1 = [
      { classList: { contains: (c: string) => c === "action-label" }, textContent: "Cut" },
      { classList: { contains: (c: string) => c === "action-label" }, textContent: "Copy" },
    ];
    fakeMenu1.forEach((el) => {
      el.textContent = translateMonacoLabel(el.textContent);
    });
    expect(fakeMenu1[0].textContent).toBe("剪切");
    expect(fakeMenu1[1].textContent).toBe("复制");

    // 模拟第二次右键弹出（重新构造项）
    const fakeMenu2 = [
      { classList: { contains: (c: string) => c === "action-label" }, textContent: "Change All Occurrences" },
      { classList: { contains: (c: string) => c === "action-label" }, textContent: "Paste" },
    ];
    fakeMenu2.forEach((el) => {
      el.textContent = translateMonacoLabel(el.textContent);
    });
    expect(fakeMenu2[0].textContent).toBe("更改所有匹配项");
    expect(fakeMenu2[1].textContent).toBe("粘贴");
  });
});


