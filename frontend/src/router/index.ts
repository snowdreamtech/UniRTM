// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import DashboardView from "../views/DashboardView.vue";
import ToolsView from "../views/ToolsView.vue";
import TasksView from "../views/TasksView.vue";
import MarketplaceView from "../views/MarketplaceView.vue";
import DoctorView from "../views/DoctorView.vue";
import SettingsView from "../views/SettingsView.vue";
import AboutView from "../views/AboutView.vue";

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
    path: "/tools",
    name: "tools",
    component: ToolsView,
  },
  {
    path: "/tasks",
    name: "tasks",
    component: TasksView,
  },
  {
    path: "/environments",
    name: "environments",
    component: () => import("../views/EnvironmentsView.vue"),
  },
  {
    path: "/marketplace/:namespace?/:name?",
    name: "marketplace",
    component: MarketplaceView,
  },
  {
    path: "/doctor",
    name: "doctor",
    component: DoctorView,
  },
  {
    path: "/settings/:category?",
    name: "settings",
    component: SettingsView,
  },
  {
    path: "/about",
    name: "about",
    component: AboutView,
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
