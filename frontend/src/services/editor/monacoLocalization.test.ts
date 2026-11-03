// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, beforeEach } from "vitest";
import { translateMonacoLabel, localizeEditorActions, getIconForLabel } from "./monacoLocalization";
import { setLanguage } from "../../i18n";
import { zhCn } from "../../i18n/locales/zh-CN";
import { zhTw } from "../../i18n/locales/zh-TW";
import { jaJp } from "../../i18n/locales/ja-JP";
import { deDe } from "../../i18n/locales/de-DE";
import { frFr } from "../../i18n/locales/fr-FR";
import { esEs } from "../../i18n/locales/es-ES";

describe("monacoLocalization", () => {
  beforeEach(async () => {
    await setLanguage("zh-CN");
  });

  it("translates Monaco context menu labels to standard Chinese", () => {
    expect(translateMonacoLabel("Change All Occurrences")).toBe(zhCn["editor.action.change_all"]);
    expect(translateMonacoLabel("Cut")).toBe(zhCn["editor.action.cut"]);
    expect(translateMonacoLabel("Copy")).toBe(zhCn["editor.action.copy"]);
    expect(translateMonacoLabel("Paste")).toBe(zhCn["editor.action.paste"]);
    expect(translateMonacoLabel("Command Palette")).toBe(zhCn["editor.action.command_palette"]);
    expect(translateMonacoLabel("Format Document")).toBe(zhCn["editor.action.format_document"]);
    expect(translateMonacoLabel("Refactor...")).toBe(zhCn["editor.action.refactor"]);
    expect(translateMonacoLabel("Rename Symbol")).toBe(zhCn["editor.action.rename_symbol"]);
    expect(translateMonacoLabel("Select All")).toBe(zhCn["editor.action.select_all"]);
    expect(translateMonacoLabel("Undo")).toBe(zhCn["editor.action.undo"]);
  });

  it("retains original label for English locale", async () => {
    await setLanguage("en-US");
    expect(translateMonacoLabel("Cut")).toBe("Cut");
    expect(translateMonacoLabel("Command Palette...")).toBe("Command Palette...");
  });

  it("supports Traditional Chinese (zh-TW)", async () => {
    await setLanguage("zh-TW");
    expect(translateMonacoLabel("Cut")).toBe(zhTw["editor.action.cut"]);
    expect(translateMonacoLabel("Copy")).toBe(zhTw["editor.action.copy"]);
    expect(translateMonacoLabel("Paste")).toBe(zhTw["editor.action.paste"]);
  });

  it("supports Japanese (ja-JP)", async () => {
    await setLanguage("ja-JP");
    expect(translateMonacoLabel("Cut")).toBe(jaJp["editor.action.cut"]);
    expect(translateMonacoLabel("Copy")).toBe(jaJp["editor.action.copy"]);
  });

  it("supports German (de-DE)", async () => {
    await setLanguage("de-DE");
    expect(translateMonacoLabel("Cut")).toBe(deDe["editor.action.cut"]);
    expect(translateMonacoLabel("Paste")).toBe(deDe["editor.action.paste"]);
  });

  it("supports French (fr-FR)", async () => {
    await setLanguage("fr-FR");
    expect(translateMonacoLabel("Cut")).toBe(frFr["editor.action.cut"]);
    expect(translateMonacoLabel("Format Document")).toBe(frFr["editor.action.format_document"]);
  });

  it("supports Spanish (es-ES)", async () => {
    await setLanguage("es-ES");
    expect(translateMonacoLabel("Cut")).toBe(esEs["editor.action.cut"]);
    expect(translateMonacoLabel("Undo")).toBe(esEs["editor.action.undo"]);
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
    expect(fakeActions[0].label).toBe(zhCn["editor.action.cut"]);
    expect(fakeActions[1].label).toBe(zhCn["editor.action.change_all"]);
  });

  it("translates repeated context menu items without reverting to English", () => {
    const fakeMenu1 = [
      { classList: { contains: (c: string) => c === "action-label" }, textContent: "Cut" },
      { classList: { contains: (c: string) => c === "action-label" }, textContent: "Copy" },
    ];
    fakeMenu1.forEach((el) => {
      el.textContent = translateMonacoLabel(el.textContent);
    });
    expect(fakeMenu1[0].textContent).toBe(zhCn["editor.action.cut"]);
    expect(fakeMenu1[1].textContent).toBe(zhCn["editor.action.copy"]);

    const fakeMenu2 = [
      { classList: { contains: (c: string) => c === "action-label" }, textContent: "Change All Occurrences" },
      { classList: { contains: (c: string) => c === "action-label" }, textContent: "Paste" },
    ];
    fakeMenu2.forEach((el) => {
      el.textContent = translateMonacoLabel(el.textContent);
    });
    expect(fakeMenu2[0].textContent).toBe(zhCn["editor.action.change_all"]);
    expect(fakeMenu2[1].textContent).toBe(zhCn["editor.action.paste"]);
  });

  it("maps proper Codicon icons for action labels", () => {
    expect(getIconForLabel("Cut")).toBe("codicon-scissors");
    expect(getIconForLabel("Copy")).toBe("codicon-copy");
    expect(getIconForLabel("Format Document")).toBe("codicon-code");
    expect(getIconForLabel("Command Palette")).toBe("codicon-terminal");
  });

  it("translates Chinese context menu labels to Japanese when language is switched", async () => {
    await setLanguage("ja-JP");
    expect(translateMonacoLabel("在活动终端中运行")).toBe(jaJp["editor.run_in_terminal"]);
    expect(translateMonacoLabel("更改所有匹配项")).toBe(jaJp["editor.action.change_all"]);
    expect(translateMonacoLabel("剪切")).toBe(jaJp["editor.action.cut"]);
    expect(translateMonacoLabel("复制")).toBe(jaJp["editor.action.copy"]);
    expect(translateMonacoLabel("粘贴")).toBe(jaJp["editor.action.paste"]);
    expect(translateMonacoLabel("命令面板")).toBe(jaJp["editor.action.command_palette"]);
  });

  it("translates Chinese labels to English when switched to en-US", async () => {
    await setLanguage("en-US");
    expect(translateMonacoLabel("在活动终端中运行")).toBe("Run in Active Terminal");
    expect(translateMonacoLabel("剪切")).toBe("Cut");
    expect(translateMonacoLabel("复制")).toBe("Copy");
    expect(translateMonacoLabel("粘贴")).toBe("Paste");
    expect(translateMonacoLabel("更改所有匹配项")).toBe("Change All Occurrences");
    expect(translateMonacoLabel("命令面板")).toBe("Command Palette...");
  });

  it("updates custom terminal action in localizeEditorActions", async () => {
    await setLanguage("ja-JP");
    const fakeTermAct: any = {
      id: "unirtm.sendSelectionToTerminal",
      label: "在活动终端中运行",
      _descriptor: { label: "在活动终端中运行" },
    };
    const fakeEditor = {
      getAction: (id: string) => (id === "unirtm.sendSelectionToTerminal" ? fakeTermAct : null),
      getSupportedActions: () => [],
    };
    localizeEditorActions(fakeEditor);
    expect(fakeTermAct.label).toBe(jaJp["editor.run_in_terminal"]);
    expect(fakeTermAct._descriptor.label).toBe(jaJp["editor.run_in_terminal"]);
  });
});
