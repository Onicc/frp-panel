import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { i18n, setLocale } from './i18n'
import { initializeTheme } from './theme'
import './styles.css'

initializeTheme()
setLocale(localStorage.getItem('frp-panel.locale') || 'zh-CN')
createApp(App).use(createPinia()).use(router).use(i18n).mount('#app')
