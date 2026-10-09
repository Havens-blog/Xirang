import { createPinia } from "pinia";
import { createApp } from "vue";
import App from "./App.vue";
import { router } from "./router";

// 运维 Agent 独立前端（子项目独立仓库，不并入平台控制台 haven-console）。
// cyan 深色设计系统：作用域 .opsagent-page，AppShell 根部即携带该类，
// 全部页面 token 命中本文件（不污染平台其它应用）。
import "./assets/styles/opsagent-theme.css";

const app = createApp(App);

app.use(createPinia());
app.use(router);

app.mount("#app");