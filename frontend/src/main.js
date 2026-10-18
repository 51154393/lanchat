import {createApp} from 'vue'
import {createPinia} from 'pinia'
import App from './App.vue'
import './style.css'

if (import.meta.env.DEV) {
  await import('./dev-mock.js')
}

const app = createApp(App)
app.use(createPinia())
app.mount('#app')
