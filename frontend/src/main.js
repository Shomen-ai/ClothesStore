// Application entry point: bootstraps the Vue app and wires in global plugins
// (Pinia store, Element Plus UI kit + icons, vue-router) before mounting to #app.
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import ru from 'element-plus/es/locale/lang/ru' // Russian locale for Element Plus (date picker months/weekdays, etc.)
import * as ElIcons from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router/index.js'
import './assets/styles/main.css'
import { useThemeStore } from './stores/theme.js'

const app = createApp(App)
app.use(createPinia())                       // global state container (must be installed before any store is used)
app.use(ElementPlus, { size: 'default', locale: ru })    // Element Plus component library (Russian locale)
app.use(router)                              // client-side routing

// Register every Element Plus icon as a global component (e.g. <Close />, used in ProductForm).
Object.keys(ElIcons).forEach(key => app.component(key, ElIcons[key]))

// Apply the persisted/system color theme before first paint to avoid a flash of the wrong theme.
// Note: relies on Pinia being installed above, since useThemeStore() resolves the active Pinia instance.
useThemeStore().init()

app.mount('#app')
