// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package theme

type ThemeType string

const (
	ThemeTypeColor   ThemeType = "color"
	ThemeTypeIcon    ThemeType = "icon"
	ThemeTypeProduct ThemeType = "product"
)

// VSCodeThemeContribute matches the contributes.themes entry in package.json
type VSCodeThemeContribute struct {
	ID      string `json:"id,omitempty"`
	Label   string `json:"label"`
	UITheme string `json:"uiTheme,omitempty"` // "vs-dark", "vs", "hc-black"
	Path    string `json:"path"`
}

// VSCodeIconThemeContribute matches the contributes.iconThemes entry in package.json
type VSCodeIconThemeContribute struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

// VSCodeProductIconThemeContribute matches the contributes.productIconThemes entry in package.json
type VSCodeProductIconThemeContribute struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

// VSCodePackageJSON represents the manifest of an unpacked theme package
type VSCodePackageJSON struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Publisher   string `json:"publisher,omitempty"`
	Contributes struct {
		Themes            []VSCodeThemeContribute            `json:"themes,omitempty"`
		IconThemes        []VSCodeIconThemeContribute        `json:"iconThemes,omitempty"`
		ProductIconThemes []VSCodeProductIconThemeContribute `json:"productIconThemes,omitempty"`
	} `json:"contributes"`
}

// ThemeItem represents a registered theme item exposed to the API and UI
type ThemeItem struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         ThemeType `json:"type"`
	IsBuiltin    bool      `json:"is_builtin"`
	UITheme      string    `json:"ui_theme,omitempty"`
	PackageID    string    `json:"package_id,omitempty"`
	RelativePath string    `json:"relative_path,omitempty"`
	Description  string    `json:"description,omitempty"`
	Version      string    `json:"version,omitempty"`
}

// ThemesListResponse represents the payload of GET /api/themes
type ThemesListResponse struct {
	Colors   []ThemeItem `json:"colors"`
	Icons    []ThemeItem `json:"icons"`
	Products []ThemeItem `json:"products"`
	Active   struct {
		Color      string `json:"color"`
		Icon       string `json:"icon"`
		Product    string `json:"product"`
		Appearance string `json:"appearance,omitempty"`
	} `json:"active"`
}
