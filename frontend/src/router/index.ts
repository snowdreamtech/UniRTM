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

export default router;
