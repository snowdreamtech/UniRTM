# Quickstart: Using the VS Code Icon Theme System in UniRTM

## 1. Using `<TechIcon>` in Components

Replace any hardcoded emojis or switch statements with `<TechIcon>`:

```vue
<template>
  <div class="runtime-card">
    <!-- Render tool or language icon -->
    <TechIcon :id="runtime.id" :size="24" :colored="true" />
    <span class="runtime-name">{{ runtime.name }}</span>

    <!-- Render lockfile / config file icon -->
    <TechIcon id="unirtm.lock" :size="16" />
    <span class="file-name">unirtm.lock</span>
  </div>
</template>

<script setup lang="ts">
import TechIcon from "@/components/common/TechIcon.vue";
</script>
```

## 2. Resolving Icons Programmatically

You can also use the composable directly in logic:

```typescript
import { useIconTheme } from "@/composables/useIconTheme";

const { resolveIcon } = useIconTheme();

const iconInfo = resolveIcon("go");
// => { character: "\uE024", color: "#00ADD8", type: "font", fontFamily: "seti" }
```

## 3. Adding a Custom Tool or Alias

To add a new tool alias (e.g. `shfmt` -> `shell`):
Simply add the alias to `languageIds` in `seti.json` or custom theme JSON:

```json
"languageIds": {
  "shfmt": "_shell"
}
```

No frontend code recompilation or component changes required.
