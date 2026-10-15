// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package theme

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// Manager manages installed VS Code themes and icon packages.
type Manager struct {
	baseDir      string
	iconsDir     string
	colorsDir    string
}

// NewManager creates and initializes a theme manager using standard user paths.
func NewManager() *Manager {
	baseDir := env.GetThemesDir()
	iconsDir := env.GetIconThemesDir()
	colorsDir := env.GetColorThemesDir()

	_ = os.MkdirAll(iconsDir, 0755)
	_ = os.MkdirAll(colorsDir, 0755)

	return &Manager{
		baseDir:   baseDir,
		iconsDir:  iconsDir,
		colorsDir: colorsDir,
	}
}

// EnsureDirectories ensures that storage directories exist on disk.
func (m *Manager) EnsureDirectories() error {
	if err := os.MkdirAll(m.iconsDir, 0755); err != nil {
		return err
	}
	return os.MkdirAll(m.colorsDir, 0755)
}

// ListThemes scans user directory and returns both builtin and installed themes.
func (m *Manager) ListThemes() (*ThemesListResponse, error) {
	resp := &ThemesListResponse{
		Colors: []ThemeItem{
			{
				ID:        "default",
				Name:      "UniRTM (Default)",
				Type:      ThemeTypeColor,
				IsBuiltin: true,
				UITheme:   "vs-dark",
			},
			{
				ID:        "dark",
				Name:      "Dark (UniRTM Default)",
				Type:      ThemeTypeColor,
				IsBuiltin: true,
				UITheme:   "vs-dark",
			},
			{
				ID:        "light",
				Name:      "Light",
				Type:      ThemeTypeColor,
				IsBuiltin: true,
				UITheme:   "vs",
			},
			{
				ID:        "system",
				Name:      "Follow System",
				Type:      ThemeTypeColor,
				IsBuiltin: true,
				UITheme:   "vs-dark",
			},
		},
		Icons: []ThemeItem{
			{
				ID:          "seti",
				Name:        "Seti UI",
				Type:        ThemeTypeIcon,
				IsBuiltin:   true,
				Description: "Official Visual Studio Code default file icon theme",
			},
		},
	}

	// Scan installed color themes
	if colorEntries, err := os.ReadDir(m.colorsDir); err == nil {
		for _, entry := range colorEntries {
			if !entry.IsDir() {
				continue
			}
			pkgDir := filepath.Join(m.colorsDir, entry.Name())
			pkg, err := m.readPackageJSON(pkgDir)
			if err != nil {
				continue
			}
			for _, th := range pkg.Contributes.Themes {
				thID := th.ID
				if thID == "" {
					thID = th.Label
				}
				resp.Colors = append(resp.Colors, ThemeItem{
					ID:           thID,
					Name:         th.Label,
					Type:         ThemeTypeColor,
					IsBuiltin:    false,
					UITheme:      th.UITheme,
					PackageID:    entry.Name(),
					RelativePath: th.Path,
					Description:  pkg.Description,
					Version:      pkg.Version,
				})
			}
		}
	}

	// Scan installed icon themes
	if iconEntries, err := os.ReadDir(m.iconsDir); err == nil {
		for _, entry := range iconEntries {
			if !entry.IsDir() {
				continue
			}
			pkgDir := filepath.Join(m.iconsDir, entry.Name())
			pkg, err := m.readPackageJSON(pkgDir)
			if err != nil {
				continue
			}
			for _, ic := range pkg.Contributes.IconThemes {
				icID := ic.ID
				if icID == "" {
					icID = ic.Label
				}
				resp.Icons = append(resp.Icons, ThemeItem{
					ID:           icID,
					Name:         ic.Label,
					Type:         ThemeTypeIcon,
					IsBuiltin:    false,
					PackageID:    entry.Name(),
					RelativePath: ic.Path,
					Description:  pkg.Description,
					Version:      pkg.Version,
				})
			}
		}
	}

	return resp, nil
}

// readPackageJSON parses extension package.json from directory
func (m *Manager) readPackageJSON(pkgDir string) (*VSCodePackageJSON, error) {
	data, err := os.ReadFile(filepath.Join(pkgDir, "package.json"))
	if err != nil {
		return nil, err
	}
	var pkg VSCodePackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

// InstallVSIX extracts a .vsix or .zip file into user directory.
func (m *Manager) InstallVSIX(r io.Reader) ([]ThemeItem, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "unirtm-theme-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, r); err != nil {
		return nil, fmt.Errorf("failed to copy uploaded data: %w", err)
	}

	// Open zip archive
	zipReader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("invalid .vsix or zip archive: %w", err)
	}
	defer zipReader.Close()

	// Locate and parse extension/package.json
	var pkgJSON *VSCodePackageJSON
	for _, f := range zipReader.File {
		cleanName := filepath.Clean(f.Name)
		if cleanName == "extension/package.json" || cleanName == "package.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to read package.json: %w", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read package.json content: %w", err)
			}
			var parsed VSCodePackageJSON
			if err := json.Unmarshal(content, &parsed); err != nil {
				return nil, fmt.Errorf("failed to parse package.json: %w", err)
			}
			pkgJSON = &parsed
			break
		}
	}

	if pkgJSON == nil {
		return nil, errors.New("invalid VS Code extension: missing extension/package.json")
	}

	hasThemes := len(pkgJSON.Contributes.Themes) > 0
	hasIconThemes := len(pkgJSON.Contributes.IconThemes) > 0

	if !hasThemes && !hasIconThemes {
		return nil, errors.New("package does not contribute any color or icon themes")
	}

	pkgName := pkgJSON.Name
	if pkgName == "" {
		pkgName = "custom-theme"
	}
	// Sanitize package directory name
	pkgName = strings.ReplaceAll(pkgName, "/", "-")
	pkgName = strings.ReplaceAll(pkgName, "\\", "-")
	pkgName = strings.ReplaceAll(pkgName, "..", "")

	var installedItems []ThemeItem

	// Unpack to target folders
	if hasThemes {
		targetDir := filepath.Join(m.colorsDir, pkgName)
		if err := m.extractExtensionFiles(zipReader.File, targetDir); err != nil {
			return nil, fmt.Errorf("failed to extract color theme files: %w", err)
		}
		for _, th := range pkgJSON.Contributes.Themes {
			thID := th.ID
			if thID == "" {
				thID = th.Label
			}
			installedItems = append(installedItems, ThemeItem{
				ID:           thID,
				Name:         th.Label,
				Type:         ThemeTypeColor,
				IsBuiltin:    false,
				UITheme:      th.UITheme,
				PackageID:    pkgName,
				RelativePath: th.Path,
				Description:  pkgJSON.Description,
				Version:      pkgJSON.Version,
			})
		}
	}

	if hasIconThemes {
		targetDir := filepath.Join(m.iconsDir, pkgName)
		if err := m.extractExtensionFiles(zipReader.File, targetDir); err != nil {
			return nil, fmt.Errorf("failed to extract icon theme files: %w", err)
		}
		for _, ic := range pkgJSON.Contributes.IconThemes {
			icID := ic.ID
			if icID == "" {
				icID = ic.Label
			}
			installedItems = append(installedItems, ThemeItem{
				ID:           icID,
				Name:         ic.Label,
				Type:         ThemeTypeIcon,
				IsBuiltin:    false,
				PackageID:    pkgName,
				RelativePath: ic.Path,
				Description:  pkgJSON.Description,
				Version:      pkgJSON.Version,
			})
		}
	}

	return installedItems, nil
}

// extractExtensionFiles safely unpacks files rooted in 'extension/' into targetDir
func (m *Manager) extractExtensionFiles(files []*zip.File, targetDir string) error {
	_ = os.RemoveAll(targetDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	for _, f := range files {
		if strings.Contains(f.Name, "..") {
			return fmt.Errorf("illegal relative path traversal in zip: %s", f.Name)
		}

		cleanName := filepath.Clean(f.Name)
		var relPath string
		if strings.HasPrefix(cleanName, "extension/") || strings.HasPrefix(cleanName, "extension\\") {
			relPath = cleanName[len("extension/"):]
		} else if strings.HasPrefix(cleanName, "extension") {
			relPath = strings.TrimPrefix(cleanName, "extension")
		} else {
			continue // Skip manifest root files outside of extension/
		}

		relPath = strings.TrimPrefix(relPath, "/")
		relPath = strings.TrimPrefix(relPath, "\\")
		if relPath == "" {
			continue
		}

		destPath := filepath.Join(targetDir, relPath)
		// Zip Slip safety check
		if !strings.HasPrefix(destPath, filepath.Clean(targetDir)+string(filepath.Separator)) {
			return fmt.Errorf("illegal file path in zip: %s", cleanName)
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(destPath, 0755)
			continue
		}

		_ = os.MkdirAll(filepath.Dir(destPath), 0755)
		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteTheme removes an installed theme package from user directory.
func (m *Manager) DeleteTheme(themeType ThemeType, packageID string) error {
	if packageID == "" || strings.Contains(packageID, "..") || strings.Contains(packageID, "/") || strings.Contains(packageID, "\\") {
		return errors.New("invalid package ID")
	}

	var targetDir string
	if themeType == ThemeTypeColor {
		targetDir = filepath.Join(m.colorsDir, packageID)
	} else if themeType == ThemeTypeIcon {
		targetDir = filepath.Join(m.iconsDir, packageID)
	} else {
		return errors.New("invalid theme type")
	}

	return os.RemoveAll(targetDir)
}

// ResolveThemeFile returns the absolute filesystem path to a theme file or asset.
func (m *Manager) ResolveThemeFile(themeType ThemeType, packageID, subpath string) (string, error) {
	if packageID == "" || strings.Contains(packageID, "..") {
		return "", errors.New("invalid package ID")
	}

	var baseDir string
	if themeType == ThemeTypeColor {
		baseDir = filepath.Join(m.colorsDir, packageID)
	} else if themeType == ThemeTypeIcon {
		baseDir = filepath.Join(m.iconsDir, packageID)
	} else {
		return "", errors.New("invalid theme type")
	}

	cleanSub := filepath.Clean(strings.TrimPrefix(subpath, "/"))
	target := filepath.Join(baseDir, cleanSub)
	if !strings.HasPrefix(target, filepath.Clean(baseDir)) {
		return "", errors.New("path traversal detected")
	}

	if _, err := os.Stat(target); err != nil {
		return "", err
	}

	return target, nil
}
