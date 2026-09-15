import { defineConfig } from "vitepress";
import { zhConfig } from "./config/zh";
import { enConfig } from "./config/en";

export default defineConfig({
  base: "/frp-panel/",
  locales: {
    root: { label: "简体中文", ...zhConfig },
    en: { label: "English", ...enConfig },
  },
  title: "frp-panel v2",
  description: "Onicc 维护的安全、跨平台 FRP 控制面",
  cleanUrls: true,
  lastUpdated: true,
});
