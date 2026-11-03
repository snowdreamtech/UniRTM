# Phase 1 Data Model: VS Code Theme Package Management System

**Branch**: `031-vscode-theme-manager` | **Date**: 2026-10-15

## 1. Backend Go Models

```go
package theme

type ThemeType string

const (
    ThemeTypeColor ThemeType = "color"
    ThemeTypeIcon  ThemeType = "icon"
)

type ThemeItem struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Type        ThemeType `json:"type"`
    IsBuiltin   bool      `json:"is_builtin"`
    UITheme     string    `json:"ui_theme,omitempty"` // "vs-dark" or "vs"
    Path        string    `json:"path"`
    Description string    `json:"description,omitempty"`
    Author      string    `json:"author,omitempty"`
    Version     string    `json:"version,omitempty"`
}

type ThemesListResponse struct {
    Colors []ThemeItem `json:"colors"`
    Icons  []ThemeItem `json:"icons"`
    Active struct {
        Color string `json:"color"`
        Icon  string `json:"icon"`
    } `json:"active"`
}
```

## 2. Frontend TypeScript Models

```typescript
export interface ThemeEntry {
  id: string;
  name: string;
  type: "color" | "icon";
  is_builtin: boolean;
  ui_theme?: "vs-dark" | "vs" | string;
  path: string;
  description?: string;
  version?: string;
}

export interface ThemesResponse {
  colors: ThemeEntry[];
  icons: ThemeEntry[];
  active: {
    color: string;
    icon: string;
  };
}
```
