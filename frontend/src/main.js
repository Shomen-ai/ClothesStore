import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElIcons from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router/index.js'
import './assets/styles/main.css'

const app = createApp(App)
app.use(createPinia())
app.use(ElementPlus, { size: 'default' })
app.use(router)

Object.keys(ElIcons).forEach(key => app.component(key, ElIcons[key]))

app.mount('#app')
