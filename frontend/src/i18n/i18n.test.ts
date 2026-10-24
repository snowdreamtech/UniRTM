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

  it("ensures all t() usages across codebase match existing keys in en-US", async () => {
    const fs = await import("fs");
    const path = await import("path");
    const enModule = await import("./locales/en-US");
    const validKeys = new Set(Object.keys(enModule.enUs));

    const scanDir = (dir: string, fileList: string[] = []): string[] => {
      const files = fs.readdirSync(dir);
      for (const f of files) {
        const fullPath = path.join(dir, f);
        const stat = fs.statSync(fullPath);
        if (stat.isDirectory()) {
          if (f !== "node_modules" && f !== "dist" && f !== "locales") {
            scanDir(fullPath, fileList);
          }
        } else if (f.endsWith(".vue") || (f.endsWith(".ts") && !f.endsWith(".test.ts"))) {
          fileList.push(fullPath);
        }
      }
      return fileList;
    };

    const srcDir = path.resolve(__dirname, "..");
    const files = scanDir(srcDir);
    const missing: { file: string; key: string }[] = [];
    const tCallRegex = /\bt\(\s*["']([^"']+)["']/g;

    for (const file of files) {
      const code = fs.readFileSync(file, "utf-8");
      let match: RegExpExecArray | null;
      while ((match = tCallRegex.exec(code)) !== null) {
        const key = match[1].trim();
        if (!key.includes(" ") && !key.includes("${")) {
          if (!validKeys.has(key)) {
            missing.push({ file: path.relative(srcDir, file), key });
          }
        }
      }
    }

    expect(missing, `Found unregistered i18n keys: ${JSON.stringify(missing)}`).toEqual([]);
  });
});
