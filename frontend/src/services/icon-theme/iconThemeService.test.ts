import { describe, it, expect, beforeEach } from "vitest";
import { iconThemeService } from "./iconThemeService";
import { workbenchSemanticIcons } from "./workbenchIcons";

describe("IconThemeService & Seti Fallback Engine", () => {
  beforeEach(() => {
    iconThemeService.switchTheme("seti");
  });

  it("should resolve exact technology names via Seti UI", () => {
    const goIcon = iconThemeService.resolveIcon("go");
    expect(goIcon).toBeDefined();
    expect(goIcon.type).toBe("font");
    expect(goIcon.fontFamily).toBe("seti");
    expect(goIcon.definitionKey).toBe("_go");
  });

  it("should resolve semantic workbench keys with Seti fallback", () => {
    const dashboardIcon = iconThemeService.resolveIcon("workbench:dashboard");
    expect(dashboardIcon).toBeDefined();
    expect(dashboardIcon.definitionKey).toBe(workbenchSemanticIcons["workbench:dashboard"].fallbackSetiKey);

    const doctorIcon = iconThemeService.resolveIcon("workbench:doctor");
    expect(doctorIcon).toBeDefined();
    expect(doctorIcon.definitionKey).toBe(workbenchSemanticIcons["workbench:doctor"].fallbackSetiKey);

    const tasksIcon = iconThemeService.resolveIcon("workbench:tasks");
    expect(tasksIcon).toBeDefined();
    expect(tasksIcon.definitionKey).toBe(workbenchSemanticIcons["workbench:tasks"].fallbackSetiKey);
  });

  it("should gracefully fallback to Seti UI _default on completely unknown identifiers", () => {
    const unknownIcon = iconThemeService.resolveIcon("totally-nonexistent-unknown-runtime-12345");
    expect(unknownIcon).toBeDefined();
    expect(unknownIcon.definitionKey).toBe("_default");
    expect(unknownIcon.fontFamily).toBe("seti");
    expect(unknownIcon.character).toBe(iconThemeService.parseCharacter("\\E023"));
  });

  it("should fallback gracefully when third-party theme has gaps", () => {
    // Register a mock partial theme that only knows python
    iconThemeService.registerCustomTheme("partial-theme", {
      iconDefinitions: {
        _py: { fontCharacter: "\\E07C", fontColor: "#3776ab" },
      },
      fileExtensions: {
        py: "_py",
      },
    });

    iconThemeService.switchTheme("partial-theme");

    // Match in custom theme
    const pyIcon = iconThemeService.resolveIcon("py");
    expect(pyIcon.definitionKey).toBe("_py");
    expect(pyIcon.isFallback).toBe(false);

    // Gaps fallback to Seti UI exact
    const rustIcon = iconThemeService.resolveIcon("rust");
    expect(rustIcon.definitionKey).toBe("_rust");
    expect(rustIcon.isFallback).toBe(true);

    // Unknown falls back to Seti UI default
    const strangeIcon = iconThemeService.resolveIcon("strange.unknownext");
    expect(strangeIcon.definitionKey).toBe("_default");
    expect(strangeIcon.isFallback).toBe(true);
  });

  it("should resolve workbench:delete to high-fidelity SVG", () => {
    const deleteIcon = iconThemeService.resolveIcon("workbench:delete");
    expect(deleteIcon).toBeDefined();
    expect(deleteIcon.type).toBe("svg");
    expect(deleteIcon.definitionKey).toBe("builtin:workbench:delete");
  });
});
