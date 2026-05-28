import { defineConfig } from "vitepress";

let releases: any[] = [];
try {
  const res = await fetch("https://api.github.com/repos/snowdreamtech/UniRTM/releases");
  if (res.ok) {
    const data = await res.json();
    if (Array.isArray(data)) {
      releases = data.slice(0, 5).map((r: any) => ({
        text: r.name || r.tag_name,
        link: r.html_url,
      }));
    }
  }
} catch (e) {
  console.error("Failed to fetch GitHub releases:", e);
}

const versionNavEn = {
  text: "Versions",
  items: releases.length > 0
      ? [...releases, { text: "All Releases", link: "https://github.com/snowdreamtech/UniRTM/releases" }]
      : [{ text: "GitHub Releases", link: "https://github.com/snowdreamtech/UniRTM/releases" }],
};

const versionNavZh = {
  text: "发行版本",
  items: releases.length > 0
      ? [...releases, { text: "所有版本", link: "https://github.com/snowdreamtech/UniRTM/releases" }]
      : [{ text: "GitHub Releases", link: "https://github.com/snowdreamtech/UniRTM/releases" }],
};

const sidebarEn = [
  {
    text: "Guides",
    items: [
      { text: "Getting Started", link: "/guide/getting-started" },
      { text: "Walkthrough", link: "/guide/walkthrough" },
      { text: "Installing UniRTM", link: "/guide/installing-unirtm" },
      { text: "IDE Integration", link: "/guide/ide-integration" },
      { text: "Continuous Integration", link: "/guide/continuous-integration" },
    ],
  },
  {
    text: "Configuration",
    items: [
      { text: ".unirtm.toml", link: "/configuration/unirtm-toml" },
      { text: "Settings", link: "/configuration/settings" },
      { text: "Configuration Environments", link: "/configuration/environments" },
    ],
  },
  {
    text: "Dev Tools",
    items: [
      { text: "Dev Tools Overview", link: "/dev-tools/overview" },
      { text: "Comparison to asdf/mise", link: "/dev-tools/comparison-to-asdf" },
      { text: "Shims & PATH", link: "/dev-tools/shims" },
      { text: "Lockfile (unirtm.lock)", link: "/dev-tools/unirtm-lock" },
      { text: "Security (Trivy & Syft)", link: "/dev-tools/security" },
    ],
  },
  {
    text: "Environments",
    items: [
      { text: "Environment Variables", link: "/environments/overview" },
      { text: "Secrets Management", link: "/environments/secrets" },
    ],
  },
  {
    text: "Tasks",
    items: [
      { text: "Task Overview", link: "/tasks/overview" },
      { text: "Running Tasks", link: "/tasks/running-tasks" },
      { text: "TOML Tasks", link: "/tasks/toml-tasks" },
      { text: "File Tasks", link: "/tasks/file-tasks" },
    ],
  },
  {
    text: "Plugins",
    items: [
      { text: "Plugin Overview", link: "/plugins/overview" },
      { text: "Backend Development", link: "/plugins/backend-development" },
    ],
  },
  {
    text: "CLI Reference",
    collapsed: true,
    items: [
      { text: "CLI Overview", link: "/cli/overview" },
      { text: "unirtm install", link: "/cli/install" },
      { text: "unirtm use", link: "/cli/use" },
      { text: "unirtm run", link: "/cli/run" },
      { text: "unirtm env", link: "/cli/env" },
      { text: "unirtm doctor", link: "/cli/doctor" },
    ],
  },
  {
    text: "Advanced",
    items: [
      { text: "Architecture", link: "/advanced/architecture" },
      { text: "Cache Behavior", link: "/advanced/cache-behavior" },
    ],
  },
  {
    text: "About",
    items: [
      { text: "FAQs", link: "/about/faq" },
      { text: "Troubleshooting", link: "/about/troubleshooting" },
    ],
  },
];

const sidebarZh = [
  {
    text: "指南",
    items: [
      { text: "快速开始", link: "/zh/guide/getting-started" },
      { text: "核心特性漫游", link: "/zh/guide/walkthrough" },
      { text: "安装 UniRTM", link: "/zh/guide/installing-unirtm" },
      { text: "IDE 深度集成", link: "/zh/guide/ide-integration" },
      { text: "CI/CD 集成", link: "/zh/guide/continuous-integration" },
    ],
  },
  {
    text: "配置",
    items: [
      { text: ".unirtm.toml 详解", link: "/zh/configuration/unirtm-toml" },
      { text: "全局设置", link: "/zh/configuration/settings" },
      { text: "环境隔离配置", link: "/zh/configuration/environments" },
    ],
  },
  {
    text: "开发工具 (Dev Tools)",
    items: [
      { text: "工具管理概览", link: "/zh/dev-tools/overview" },
      { text: "与 asdf/mise 的对比", link: "/zh/dev-tools/comparison-to-asdf" },
      { text: "Shims 与 PATH 劫持", link: "/zh/dev-tools/shims" },
      { text: "锁定文件 (unirtm.lock)", link: "/zh/dev-tools/unirtm-lock" },
      { text: "安全检测 (Trivy & Syft)", link: "/zh/dev-tools/security" },
    ],
  },
  {
    text: "环境变量 (Environments)",
    items: [
      { text: "变量管理概览", link: "/zh/environments/overview" },
      { text: "密钥管理机制", link: "/zh/environments/secrets" },
    ],
  },
  {
    text: "任务运行器 (Tasks)",
    items: [
      { text: "任务运行器概览", link: "/zh/tasks/overview" },
      { text: "执行任务", link: "/zh/tasks/running-tasks" },
      { text: "配置 TOML 任务", link: "/zh/tasks/toml-tasks" },
      { text: "文件型任务", link: "/zh/tasks/file-tasks" },
    ],
  },
  {
    text: "插件系统 (Plugins)",
    items: [
      { text: "插件概览", link: "/zh/plugins/overview" },
      { text: "后端开发指南", link: "/zh/plugins/backend-development" },
    ],
  },
  {
    text: "CLI 命令参考",
    collapsed: true,
    items: [
      { text: "命令概览", link: "/zh/cli/overview" },
      { text: "unirtm install", link: "/zh/cli/install" },
      { text: "unirtm use", link: "/zh/cli/use" },
      { text: "unirtm run", link: "/zh/cli/run" },
      { text: "unirtm env", link: "/zh/cli/env" },
      { text: "unirtm doctor", link: "/zh/cli/doctor" },
    ],
  },
  {
    text: "进阶指南",
    items: [
      { text: "架构设计", link: "/zh/advanced/architecture" },
      { text: "缓存策略", link: "/zh/advanced/cache-behavior" },
    ],
  },
  {
    text: "关于",
    items: [
      { text: "常见问题 (FAQ)", link: "/zh/about/faq" },
      { text: "故障排查", link: "/zh/about/troubleshooting" },
    ],
  },
];

export default defineConfig({
  title: "UniRTM",
  description: "The fast, simple, cross-platform tool to manage your dev tools, environments, and tasks.",
  base: "/UniRTM/",
  ignoreDeadLinks: true,

  head: [
    ["link", { rel: "icon", href: "/UniRTM/favicon.ico" }],
    [
      "script",
      {},
      `
      (function() {
        if (typeof window !== 'undefined' && typeof localStorage !== 'undefined') {
          if (window.location.pathname === '/UniRTM/' || window.location.pathname === '/UniRTM') {
            const hasPref = localStorage.getItem('unirtm_lang_pref');
            if (!hasPref) {
              const lang = navigator.language || navigator.userLanguage;
              if (lang && lang.toLowerCase().startsWith('zh')) {
                window.location.href = '/UniRTM/zh/';
              }
            }
          }
          window.addEventListener('click', function(e) {
            const target = e.target.closest('a');
            if (target && target.href) {
              const url = new URL(target.href);
              if (url.pathname.startsWith('/UniRTM/zh')) {
                localStorage.setItem('unirtm_lang_pref', 'zh');
              } else if (url.pathname === '/UniRTM/' || url.pathname.startsWith('/UniRTM/guide')) {
                localStorage.setItem('unirtm_lang_pref', 'en');
              }
            }
          });
        }
      })();
      `,
    ],
  ],

  themeConfig: {
    logo: "/logo.png",
    socialLinks: [
      { icon: "github", link: "https://github.com/snowdreamtech/UniRTM" },
    ],
    search: {
      provider: "local",
    },
  },

  locales: {
    root: {
      label: "English",
      lang: "en-US",
      themeConfig: {
        nav: [
          { text: "Docs", link: "/guide/getting-started" },
          versionNavEn,
          { text: "Changelog", link: "https://github.com/snowdreamtech/UniRTM/blob/main/CHANGELOG.md" },
        ],
        sidebar: sidebarEn,
        footer: {
          message: "Released under the MIT License.",
          copyright: "Copyright © 2026-present SnowdreamTech Inc.",
        },
        editLink: {
          pattern: "https://github.com/snowdreamtech/UniRTM/edit/main/docs/:path",
          text: "Edit this page on GitHub",
        },
      },
    },
    zh: {
      label: "简体中文",
      lang: "zh-CN",
      link: "/zh/",
      description: "快速、简单、跨平台的工具，统一管理您的开发工具、环境变量和任务。",
      themeConfig: {
        nav: [
          { text: "文档", link: "/zh/guide/getting-started" },
          versionNavZh,
          { text: "更新日志", link: "https://github.com/snowdreamtech/UniRTM/blob/main/CHANGELOG.md" },
        ],
        sidebar: sidebarZh,
        footer: {
          message: "基于 MIT 许可发布。",
          copyright: "版权所有 © 2026-present SnowdreamTech Inc.",
        },
        editLink: {
          pattern: "https://github.com/snowdreamtech/UniRTM/edit/main/docs/:path",
          text: "在 GitHub 上编辑此页",
        },
        docFooter: { prev: "上一页", next: "下一页" },
        outline: { label: "页面导航" },
        lastUpdated: { text: "最后更新于" },
        returnToTopLabel: "回到顶部",
        sidebarMenuLabel: "菜单",
        darkModeSwitchLabel: "主题",
        lightModeSwitchTitle: "切换到浅色模式",
        darkModeSwitchTitle: "切换到深色模式",
      },
    },
  },
});
