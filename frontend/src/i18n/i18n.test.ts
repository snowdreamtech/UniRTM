import { describe, it, expect, beforeEach } from "vitest";
import { t, setLanguage, currentLocale, isRtl, SUPPORTED_LANGUAGES } from "./index";

describe("i18n module", () => {
  beforeEach(async () => {
    await setLanguage("en-US");
  });

  it("translates pre-bundled en-US string correctly", () => {
    expect(currentLocale.value).toBe("en-US");
    const appTitle = t("app.title");
    expect(appTitle).toBe("UniRTM");
  });

  it("interpolates parameters correctly with {key}", async () => {
    await setLanguage("en-US");
    const notice = t("about.newVersionNotice", { tag: "v1.2.3" });
    expect(notice).toContain("v1.2.3");
  });

  it("switches to zh-CN and translates", async () => {
    await setLanguage("zh-CN");
    expect(currentLocale.value).toBe("zh-CN");
    expect(t("common.close")).toBe("关闭");
  });

  it("identifies RTL languages correctly", async () => {
    await setLanguage("ar-SA");
    expect(isRtl.value).toBe(true);

    await setLanguage("en-US");
    expect(isRtl.value).toBe(false);
  });

  it("contains 53 supported languages", () => {
    expect(SUPPORTED_LANGUAGES.length).toBe(53);
  });

  it("ensures all 53 locales contain all translation keys from en-US", async () => {
    const enModule = await import("./locales/en-US");
    const enDict = enModule.enUs;
    const enKeys = Object.keys(enDict);

    for (const lang of SUPPORTED_LANGUAGES) {
      if (lang.code === "en-US") continue;
      const mod = await import(`./locales/${lang.code}.ts`);
      const exportKey = Object.keys(mod).find((k) => k !== "default") || Object.keys(mod)[0];
      const dict = mod[exportKey];
      expect(dict, `Locale dictionary for ${lang.code} should exist`).toBeDefined();
      const localeKeys = new Set(Object.keys(dict));
      const missingKeys = enKeys.filter((k) => !localeKeys.has(k));
      expect(missingKeys, `Missing keys in ${lang.code}`).toEqual([]);
    }
  }, 60000);

  it("falls back to en-US when encountering unknown locale or missing fallback key", async () => {
    await setLanguage("invalid-locale");
    expect(currentLocale.value).toBe("en-US");
    expect(t("common.close")).toBe("Close");
  });
});
