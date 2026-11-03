# Data Model: Open VSX Theme Marketplace

## Backend Types (`internal/theme/marketplace.go`)

```go
type MarketplaceExtensionFiles struct {
	Download  string `json:"download"`
	Icon      string `json:"icon,omitempty"`
}

type MarketplaceExtension struct {
	Namespace     string                   `json:"namespace"`
	Name          string                   `json:"name"`
	DisplayName   string                   `json:"displayName"`
	Version       string                   `json:"version"`
	Description   string                   `json:"description"`
	DownloadCount int64                    `json:"downloadCount"`
	AverageRating float64                  `json:"averageRating"`
	Files         MarketplaceExtensionFiles `json:"files"`
	URL           string                   `json:"url"`
	IsInstalled   bool                     `json:"is_installed"`
}

type MarketplaceSearchResponse struct {
	Offset     int                    `json:"offset"`
	TotalSize  int                    `json:"totalSize"`
	Extensions []MarketplaceExtension `json:"extensions"`
}
```

## Frontend Types (`frontend/src/types/theme-manager.ts`)

```typescript
export interface MarketplaceExtension {
  namespace: string;
  name: string;
  displayName: string;
  version: string;
  description: string;
  downloadCount: number;
  averageRating: number;
  files: {
    download: string;
    icon?: string;
  };
  url?: string;
  is_installed?: boolean;
}

export interface MarketplaceSearchResponse {
  offset: number;
  totalSize: number;
  extensions: MarketplaceExtension[];
}
```
