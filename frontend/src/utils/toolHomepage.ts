/**
 * Resolves the official homepage or upstream repository URL for any tool.
 * Supports explicit tool homepages, ecosystem prefixes (github:, npm:, go:, pipx:),
 * and standard language runtimes.
 */
export function getToolHomepage(tool: { id?: string; name?: string; homepage?: string } | string): string {
  if (typeof tool === "string") {
    tool = { id: tool };
  }

  // 1. If explicit homepage is provided and non-empty, use it directly
  if (tool.homepage && tool.homepage.trim()) {
    return tool.homepage.trim();
  }

  const id = (tool.id || tool.name || "").trim();
  if (!id) return "https://github.com";

  // 2. GitHub tools (e.g. github:anchore/syft, github:astral-sh/ruff, github:cli/cli)
  if (id.startsWith("github:")) {
    const repoPath = id.replace(/^github:/, "").trim();
    return `https://github.com/${repoPath}`;
  }

  // 3. NPM packages (e.g. npm:@commitlint/cli, npm:prettier, npm:eslint)
  if (id.startsWith("npm:")) {
    const pkg = id.replace(/^npm:/, "").trim();
    return `https://www.npmjs.com/package/${pkg}`;
  }

  // 4. Go modules / packages (e.g. go:mvdan.cc/sh/v3/cmd/shfmt, go:golang.org/x/tools/cmd/goimports)
  if (id.startsWith("go:")) {
    const pkg = id.replace(/^go:/, "").trim();
    if (pkg.startsWith("github.com/")) {
      return `https://${pkg}`;
    }
    return `https://pkg.go.dev/${pkg}`;
  }

  // 5. Python pipx packages (e.g. pipx:pre-commit, pipx:yamllint)
  if (id.startsWith("pipx:")) {
    const pkg = id.replace(/^pipx:/, "").trim();
    return `https://pypi.org/project/${pkg}/`;
  }

  // 6. Well-known language runtimes and CLI tools
  const knownSites: Record<string, string> = {
    node: "https://nodejs.org",
    nodejs: "https://nodejs.org",
    go: "https://go.dev",
    golang: "https://go.dev",
    python: "https://www.python.org",
    rust: "https://www.rust-lang.org",
    cargo: "https://crates.io",
    java: "https://openjdk.org",
    bun: "https://bun.sh",
    deno: "https://deno.land",
    ruby: "https://www.ruby-lang.org",
    zig: "https://ziglang.org",
    php: "https://www.php.net",
    dart: "https://dart.dev",
    flutter: "https://flutter.dev",
    elixir: "https://elixir-lang.org",
    erlang: "https://www.erlang.org",
    scala: "https://www.scala-lang.org",
    kotlin: "https://kotlinlang.org",
    dotnet: "https://dotnet.microsoft.com",
    lua: "https://www.lua.org",
    perl: "https://www.perl.org",
    julia: "https://julialang.org",
    haskell: "https://www.haskell.org",
    terraform: "https://www.terraform.io",
    kubectl: "https://kubernetes.io",
    helm: "https://helm.sh",
  };

  const cleanId = id.toLowerCase();
  if (knownSites[cleanId]) {
    return knownSites[cleanId];
  }

  // 7. If id looks like an owner/repo or host/path
  if (id.includes("/") && !id.includes(":")) {
    return `https://github.com/${id}`;
  }

  // 8. Default fallback: GitHub search or topics
  return `https://github.com/search?q=${encodeURIComponent(id)}`;
}
