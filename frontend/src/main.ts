import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './assets/main.css'
import { useSnapshotsStore } from './stores/snapshots'
import { initializeTheme } from './utils/theme'

const app = createApp(App)
const pinia = createPinia()

initializeTheme()

app.use(pinia)
app.mount('#app')

async function bootstrapData(): Promise<void> {
  const snapshotsStore = useSnapshotsStore(pinia)
  await snapshotsStore.bootstrap()
}

void bootstrapData()
