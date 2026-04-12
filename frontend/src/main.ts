import { createApp } from 'vue'
import App from './App.vue'
import './assets/main.css'
import { initializeTheme } from './utils/theme'

initializeTheme()
createApp(App).mount('#app')
