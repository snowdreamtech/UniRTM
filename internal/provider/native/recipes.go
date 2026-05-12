package native

// GetBuiltinRecipes returns the hardcoded list of tool recipes.
// Keeping this as a function to allow future dynamic loading or overrides if needed.
func GetBuiltinRecipes() map[string]Recipe {
	return map[string]Recipe{
		"go": {
			ID:      "go",
			Handler: &GolangHandler{},
			BaseURL: "https://go.dev/dl",
			Aliases: map[string]string{
				"latest": "1.26.2",
			},
			GPGKeys: []string{
				"EB4C1BF148A2C440F831EE8752D905A67D74191C", // Google Go Release Team
			},
		},
		"golang": {
			ID:      "go",
			Handler: &GolangHandler{},
			BaseURL: "https://go.dev/dl",
			GPGKeys: []string{
				"EB4C1BF148A2C440F831EE8752D905A67D74191C",
			},
		},
		"node": {
			ID:      "node",
			Handler: &NodeJSHandler{},
			BaseURL: "https://nodejs.org/dist",
			Aliases: map[string]string{
				"lts": "22.14.0",
			},
			GPGKeys: []string{
				"C4F0DFFF4E8C1A8236409D08E73BC641CC11F4C8", // Node.js Release Team
			},
		},
		"nodejs": {
			ID:      "node",
			Handler: &NodeJSHandler{},
			BaseURL: "https://nodejs.org/dist",
			GPGKeys: []string{
				"C4F0DFFF4E8C1A8236409D08E73BC641CC11F4C8",
			},
		},
		"python": {
			ID: "python",
			Handler: &GithubHandler{
				Owner: "astral-sh",
				Repo:  "python-build-standalone",
			},
			BaseURL: "https://github.com/astral-sh/python-build-standalone/releases",
			Aliases: map[string]string{
				"latest": "3.12.2",
			},
		},
		"zig": {
			ID:      "zig",
			Handler: &ZigHandler{},
			BaseURL: "https://ziglang.org/download/index.json",
		},
		"rust": {
			ID:      "rust",
			Handler: &RustHandler{},
			BaseURL: "https://static.rust-lang.org/dist",
			Aliases: map[string]string{
				"latest": "1.76.0",
			},
		},
		"bun": {
			ID: "bun",
			Handler: &GithubHandler{
				Owner: "oven-sh",
				Repo:  "bun",
			},
			BaseURL: "https://github.com/oven-sh/bun/releases",
		},
		"deno": {
			ID: "deno",
			Handler: &GithubHandler{
				Owner: "denoland",
				Repo:  "deno",
			},
			BaseURL: "https://github.com/denoland/deno/releases",
		},
		"terraform": {
			ID: "terraform",
			Handler: &GithubHandler{
				Owner: "hashicorp",
				Repo:  "terraform",
			},
			BaseURL: "https://github.com/hashicorp/terraform/releases",
		},
		"opentofu": {
			ID: "opentofu",
			Handler: &GithubHandler{
				Owner: "opentofu",
				Repo:  "opentofu",
			},
			BaseURL: "https://github.com/opentofu/opentofu/releases",
			Aliases: map[string]string{
				"latest": "1.6.2",
			},
		},
		"java": {
			ID:      "java",
			Handler: &JavaHandler{ImageType: "jdk"},
			BaseURL: "https://api.adoptium.net",
		},
		"jdk": {
			ID:      "java",
			Handler: &JavaHandler{ImageType: "jdk"},
			BaseURL: "https://api.adoptium.net",
		},
		"jre": {
			ID:      "java",
			Handler: &JavaHandler{ImageType: "jre"},
			BaseURL: "https://api.adoptium.net",
		},
		"kubectl": {
			ID:      "kubectl",
			Handler: &KubectlHandler{},
			BaseURL: "https://dl.k8s.io",
		},
		"ruby": {
			ID:      "ruby",
			Handler: &RubyHandler{},
			BaseURL: "https://github.com/ruby/ruby-builder/releases",
		},
		"maven": {
			ID:      "maven",
			Handler: &MavenHandler{},
			BaseURL: "https://archive.apache.org/dist/maven/maven-3",
		},
		"gradle": {
			ID:      "gradle",
			Handler: &GradleHandler{},
			BaseURL: "https://services.gradle.org/versions/all",
		},
	}
}

// IsNativeTool checks if a tool has a built-in native recipe.
func IsNativeTool(name string) bool {
	_, ok := GetBuiltinRecipes()[name]
	return ok
}
