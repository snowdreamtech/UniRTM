# Contracts: Icon Theme Resolution & Workbench Semantic System

## 1. Composable Contract: `useIconTheme`

```typescript
export interface UseIconThemeReturn {
  activeThemeId: ComputedRef<string>;
  activeManifest: ComputedRef<IconThemeManifest>;
  resolveIcon: (rawIdentifier: string) => ResolvedIcon;
  setManifest: (manifest: IconThemeManifest) => boolean;
  switchTheme: (themeId: string) => boolean;
  loadAndApplyTheme: (themeItem: ThemeItem) => Promise<boolean>;
  ensureStylesInjected: () => void;
}
```

## 2. Component Contract: `<TechIcon>`

### Props
| Prop | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `id` | `string` | **required** | File name, language ID, tool name, or `workbench:<action>` semantic key |
| `size` | `number` | `16` | Icon size in pixels |
| `colored` | `boolean` | `true` | If true, applies theme-specified color; if false, inherits parent CSS color |
| `className` | `string` | `""` | Additional CSS class names to apply to element |

### Emits
None.

### DOM Output
- When resolved to SVG: `<img class="tech-icon tech-icon-svg" :src="svgPath" role="img" ... />`
- When resolved to Font: `<span class="tech-icon" role="img" ...>{{ character }}</span>`
