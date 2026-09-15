import type { DefaultTheme, LocaleSpecificConfig } from "vitepress";

export const zhConfig: LocaleSpecificConfig<DefaultTheme.Config> = {
  lang: "zh-CN",
  title: "frp-panel v2",
  description: "安全、跨平台的 FRP 控制面",
  themeConfig: {
    nav: [
      { text: "首页", link: "/" },
      { text: "部署", link: "/deployment" },
      { text: "源码", link: "https://github.com/Onicc/frp-panel" },
    ],
    sidebar: [
      { text: "快速开始", link: "/quick-start" },
      { text: "部署指南", link: "/deployment" },
      { text: "Client / Agent", link: "/agent" },
      { text: "配置", link: "/configuration" },
      { text: "架构", link: "/ARCHITECTURE_V2" },
      { text: "优化审查", link: "/OPTIMIZATION" },
      { text: "平台支持", link: "/SUPPORT_MATRIX" },
      { text: "安全", link: "/SECURITY" },
      { text: "API 错误", link: "/api-problems" },
    ],
    socialLinks: [{ icon: "github", link: "https://github.com/Onicc/frp-panel" }],
  },
};
