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
		},
		"golang": {
			ID:      "go",
			Handler: &GolangHandler{},
			BaseURL: "https://go.dev/dl",
		},
		"node": {
			ID:      "node",
			Handler: &NodeJSHandler{},
			BaseURL: "https://nodejs.org/dist",
			Aliases: map[string]string{
				"lts": "22.14.0",
			},
		},
		"nodejs": {
			ID:      "node",
			Handler: &NodeJSHandler{},
			BaseURL: "https://nodejs.org/dist",
		},
	}
}

// IsNativeTool checks if a tool has a built-in native recipe.
func IsNativeTool(name string) bool {
	_, ok := GetBuiltinRecipes()[name]
	return ok
}
