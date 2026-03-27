import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { createPinia } from 'pinia'
import adminRouter from './admin-router'
import AdminApp from './admin-app.vue'
import './style.css'

// Initialize admin secret for API requests
localStorage.setItem('admin_secret', 'admin123')

const app = createApp(AdminApp)

// Register all Element Plus icons globally
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(ElementPlus)
app.use(createPinia())
app.use(adminRouter)
app.mount('#app')
