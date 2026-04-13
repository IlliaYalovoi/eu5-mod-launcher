import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './assets/main.css'
import { useModsStore } from './stores/mods'
import { useLoadOrderStore } from './stores/loadorder'
import { useConstraintsStore } from './stores/constraints'
import { useSettingsStore } from './stores/settings'
import { initializeTheme } from './utils/theme'

const app = createApp(App)
const pinia = createPinia()

initializeTheme()

app.use(pinia)
app.mount('#app')

async function bootstrapData(): Promise<void> {
  const modsStore = useModsStore(pinia)
  const loadOrderStore = useLoadOrderStore(pinia)
  const constraintsStore = useConstraintsStore(pinia)
  const settingsStore = useSettingsStore(pinia)

  await Promise.allSettled([
	settingsStore.fetch(),
	loadOrderStore.fetch(),
	constraintsStore.fetch(),
	modsStore.fetchAll(),
  ])
}

void bootstrapData()
