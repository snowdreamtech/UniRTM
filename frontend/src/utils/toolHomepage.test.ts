import { describe, it, expect } from "vitest";
import { getToolHomepage } from "./toolHomepage";

describe("getToolHomepage", () => {
  it("should respect explicit homepage if present", () => {
    expect(
      getToolHomepage({ id: "mytool", homepage: "https://mytool.org" })
    ).toBe("https://mytool.org");
  });

  it("should resolve standard core runtimes", () => {
    expect(getToolHomepage({ id: "node" })).toBe("https://nodejs.org");
    expect(getToolHomepage({ id: "go" })).toBe("https://go.dev");
    expect(getToolHomepage({ id: "python" })).toBe("https://www.python.org");
    expect(getToolHomepage({ id: "rust" })).toBe("https://www.rust-lang.org");
    expect(getToolHomepage({ id: "bun" })).toBe("https://bun.sh");
  });

  it("should resolve GitHub tools", () => {
    expect(getToolHomepage({ id: "github:anchore/syft" })).toBe(
      "https://github.com/anchore/syft"
    );
    expect(getToolHomepage({ id: "github:astral-sh/ruff" })).toBe(
      "https://github.com/astral-sh/ruff"
    );
    expect(getToolHomepage({ id: "github:cli/cli" })).toBe(
      "https://github.com/cli/cli"
    );
  });

  it("should resolve npm packages", () => {
    expect(getToolHomepage({ id: "npm:@commitlint/cli" })).toBe(
      "https://www.npmjs.com/package/@commitlint/cli"
    );
    expect(getToolHomepage({ id: "npm:prettier" })).toBe(
      "https://www.npmjs.com/package/prettier"
    );
  });

  it("should resolve Go packages", () => {
    expect(getToolHomepage({ id: "go:mvdan.cc/sh/v3/cmd/shfmt" })).toBe(
      "https://pkg.go.dev/mvdan.cc/sh/v3/cmd/shfmt"
    );
    expect(
      getToolHomepage({
        id: "go:github.com/golangci/golangci-lint/v2/cmd/golangci-lint",
      })
    ).toBe("https://github.com/golangci/golangci-lint/v2/cmd/golangci-lint");
  });

  it("should resolve pipx packages", () => {
    expect(getToolHomepage({ id: "pipx:pre-commit" })).toBe(
      "https://pypi.org/project/pre-commit/"
    );
  });

  it("should handle string argument", () => {
    expect(getToolHomepage("github:koalaman/shellcheck")).toBe(
      "https://github.com/koalaman/shellcheck"
    );
  });
});
