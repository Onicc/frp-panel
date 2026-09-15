import type { DefaultTheme, LocaleSpecificConfig } from "vitepress";

export const enConfig: LocaleSpecificConfig<DefaultTheme.Config> = {
  lang: "en-US",
  title: "frp-panel v2",
  description: "A secure, cross-platform FRP control plane",
  themeConfig: {
    nav: [
      { text: "Home", link: "/en/" },
      { text: "Deployment", link: "/en/deployment" },
      { text: "Source", link: "https://github.com/Onicc/frp-panel" },
    ],
    sidebar: [
      { text: "Quick start", link: "/en/quick-start" },
      { text: "Deployment", link: "/en/deployment" },
      { text: "Client / Agent", link: "/en/agent" },
      { text: "Configuration", link: "/en/configuration" },
      { text: "Architecture", link: "/ARCHITECTURE_V2" },
      { text: "Optimization review", link: "/OPTIMIZATION" },
      { text: "Platform support", link: "/SUPPORT_MATRIX" },
      { text: "Security", link: "/SECURITY" },
      { text: "API errors", link: "/api-problems" },
    ],
    socialLinks: [{ icon: "github", link: "https://github.com/Onicc/frp-panel" }],
  },
};
