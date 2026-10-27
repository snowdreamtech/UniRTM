// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed, watch } from "vue";
import { useGit } from "./useGit";

export interface ProjectItem {
  id: string;
  name: string;
  path: string; // relative to workspace root or empty for top-level
  fullPath?: string;
  isGit?: boolean;
  hasToml?: boolean;
}

const STORAGE_ACTIVE_PROJECT_KEY = "unirtm_active_project";
const STORAGE_PROJECT_PINNED_KEY = "unirtm_project_pinned";

const activeProject = ref<ProjectItem | null>(null);
const isPinned = ref<boolean>(
  typeof localStorage !== "undefined" && localStorage.getItem(STORAGE_PROJECT_PINNED_KEY) === "true"
);
const availableProjects = ref<ProjectItem[]>([]);
const isInitialized = ref(false);

export function useActiveProject() {
  const { repositories, activeRepo, setActiveRepo } = useGit();

  async function loadProjects(): Promise<ProjectItem[]> {
    try {
      const res = await fetch("/api/workspace/roots");
      if (res.ok) {
        const data = await res.json();
        const roots: Array<{ id: string; name: string; path: string }> = data.roots || [];
        const items: ProjectItem[] = roots.map((r) => {
          const isGit = repositories.value.some(
            (repo) => repo.name === r.name || repo.rootPath === r.path || repo.id === r.id
          );
          return {
            id: r.id || `root-${r.name}`,
            name: r.name,
            path: r.path,
            isGit,
          };
        });

        // Also merge any git repositories not in roots
        for (const repo of repositories.value) {
          if (!items.some((i) => i.name === repo.name || i.path === repo.rootPath)) {
            items.push({
              id: repo.id || `repo-${repo.name}`,
              name: repo.name,
              path: repo.rootPath,
              isGit: true,
            });
          }
        }

        availableProjects.value = items;

        // Initialize active project if not set
        if (!activeProject.value && items.length > 0) {
          const saved = typeof localStorage !== "undefined" ? localStorage.getItem(STORAGE_ACTIVE_PROJECT_KEY) : null;
          const matched = saved ? items.find((i) => i.id === saved || i.name === saved || i.path === saved) : null;
          if (matched) {
            activeProject.value = matched;
          } else {
            // Default to first non-empty subproject if exists
            const nonRoot = items.find((i) => i.path !== "" && i.path !== ".");
            activeProject.value = nonRoot || items[0];
          }
        }
        isInitialized.value = true;
        return items;
      }
    } catch (e) {
      console.error("[useActiveProject] Failed to load workspace projects:", e);
    }
    return availableProjects.value;
  }

  function setActiveProject(projectOrNameOrId: ProjectItem | string, manualPin: boolean = false) {
    let target: ProjectItem | null = null;
    if (typeof projectOrNameOrId === "string") {
      target =
        availableProjects.value.find(
          (p) => p.id === projectOrNameOrId || p.name === projectOrNameOrId || p.path === projectOrNameOrId
        ) || null;
      if (!target && projectOrNameOrId) {
        target = {
          id: `proj-${projectOrNameOrId}`,
          name: projectOrNameOrId,
          path: projectOrNameOrId,
        };
      }
    } else {
      target = projectOrNameOrId;
    }

    if (!target) return;

    activeProject.value = target;

    if (manualPin) {
      isPinned.value = true;
      if (typeof localStorage !== "undefined") {
        localStorage.setItem(STORAGE_PROJECT_PINNED_KEY, "true");
      }
    }

    if (typeof localStorage !== "undefined") {
      localStorage.setItem(STORAGE_ACTIVE_PROJECT_KEY, target.id || target.name || target.path);
    }

    // Synchronize with Git repository if matching
    const matchingRepo = repositories.value.find(
      (r) => r.name === target?.name || r.rootPath === target?.path || r.id === target?.id
    );
    if (matchingRepo && activeRepo.value?.id !== matchingRepo.id) {
      setActiveRepo(matchingRepo);
    }
  }

  function togglePin() {
    isPinned.value = !isPinned.value;
    if (typeof localStorage !== "undefined") {
      localStorage.setItem(STORAGE_PROJECT_PINNED_KEY, String(isPinned.value));
    }
  }

  function resolveProjectFromPath(filePath: string): ProjectItem | null {
    if (!filePath || availableProjects.value.length === 0) return null;
    const cleanPath = filePath.replace(/\\/g, "/").trim();

    // Match longest matching prefix among available projects
    let bestMatch: ProjectItem | null = null;
    let maxLen = -1;

    for (const proj of availableProjects.value) {
      if (!proj.path || proj.path === "." || proj.path === "/") continue;
      const projPath = proj.path.replace(/\\/g, "/").trim();
      const projName = proj.name.trim();

      // Check if path contains or starts with projPath or projName
      if (
        cleanPath === projPath ||
        cleanPath.startsWith(projPath + "/") ||
        cleanPath.includes("/" + projPath + "/") ||
        cleanPath.includes("/" + projName + "/") ||
        cleanPath.startsWith(projName + "/")
      ) {
        if (projPath.length > maxLen) {
          maxLen = projPath.length;
          bestMatch = proj;
        }
      }
    }

    return bestMatch;
  }

  function onFileFocused(filePath: string) {
    if (isPinned.value) return; // Do not auto-switch if user manually pinned
    const matched = resolveProjectFromPath(filePath);
    if (matched && matched.id !== activeProject.value?.id) {
      setActiveProject(matched, false);
    }
  }

  // When Git activeRepo changes, synchronize activeProject if not pinned
  watch(activeRepo, (newRepo) => {
    if (!newRepo) return;
    if (isPinned.value) return;
    const matched = availableProjects.value.find(
      (p) => p.name === newRepo.name || p.path === newRepo.rootPath || p.id === newRepo.id
    );
    if (matched && matched.id !== activeProject.value?.id) {
      activeProject.value = matched;
      if (typeof localStorage !== "undefined") {
        localStorage.setItem(STORAGE_ACTIVE_PROJECT_KEY, matched.id || matched.name);
      }
    }
  });

  return {
    activeProject,
    isPinned,
    availableProjects,
    isInitialized,
    isMultiProject: computed(() => availableProjects.value.length > 1),
    loadProjects,
    setActiveProject,
    togglePin,
    resolveProjectFromPath,
    onFileFocused,
  };
}
