// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import "./styles/theme.css";

const STARTUP_LOADER_MIN_DURATION_MS = 300;
const startupStartedAt = performance.now();

function setStartupProgress(progress: number) {
  document.documentElement.style.setProperty("--startup-progress", `${progress}%`);
}

function dismissStartupLoader() {
  const elapsed = performance.now() - startupStartedAt;
  const remaining = Math.max(0, STARTUP_LOADER_MIN_DURATION_MS - elapsed);
  setStartupProgress(92);

  window.setTimeout(() => {
    setStartupProgress(100);
    document.documentElement.classList.add("app-ready");
    window.setTimeout(() => {
      document.getElementById("startup-loader")?.remove();
    }, 260);
  }, remaining);
}

window.addEventListener("unirtm:config-ready", dismissStartupLoader, { once: true });

// Fallback safety timeout: ensure startup loader is dismissed if event doesn't fire
window.setTimeout(() => {
  if (!document.documentElement.classList.contains("app-ready")) {
    dismissStartupLoader();
  }
}, 1200);

setStartupProgress(25);
createApp(App).use(router).mount("#app");
setStartupProgress(75);
