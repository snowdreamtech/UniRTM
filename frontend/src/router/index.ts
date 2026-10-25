// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import DashboardView from "../views/DashboardView.vue";

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    redirect: "/dashboard",
  },
  {
    path: "/dashboard",
    name: "dashboard",
    component: DashboardView,
  },
  {
    path: "/explorer",
    name: "explorer",
    component: () => import("../views/ExplorerView.vue"),
  },
  {
    path: "/scm",
    name: "scm",
    component: () => import("../views/SourceControlView.vue"),
  },
  {
    path: "/guard",
    name: "guard",
    component: () => import("../views/GuardView.vue"),
  },
  {
    path: "/tools",
    name: "tools",
    component: () => import("../views/ToolsView.vue"),
  },
  {
    path: "/tasks",
    name: "tasks",
    component: () => import("../views/TasksView.vue"),
  },
  {
    path: "/environments",
    name: "environments",
    component: () => import("../views/EnvironmentsView.vue"),
  },
  {
    path: "/marketplace/:namespace?/:name?",
    name: "marketplace",
    component: () => import("../views/MarketplaceView.vue"),
  },
  {
    path: "/doctor",
    name: "doctor",
    component: () => import("../views/DoctorView.vue"),
  },
  {
    path: "/settings/:category?",
    name: "settings",
    component: () => import("../views/SettingsView.vue"),
  },
  {
    path: "/about",
    name: "about",
    component: () => import("../views/AboutView.vue"),
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/dashboard",
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

import { useAppScope } from "../composables/useAppScope";

router.beforeEach((to, _from, next) => {
  const { activeScope } = useAppScope();
  const toScope = to.query.scope;
  if (!toScope || (toScope !== "user" && toScope !== "workspace")) {
    next({
      name: to.name || undefined,
      path: to.name ? undefined : to.path,
      params: to.params,
      query: {
        ...to.query,
        scope: activeScope.value || "workspace",
      },
      hash: to.hash,
      replace: true,
    });
    return;
  }
  if (activeScope.value !== toScope) {
    activeScope.value = toScope;
  }
  next();
});

export default router;
