import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useLoadOrderStore } from './loadorder'
import type { Mod } from '../types'
import { GetAllMods, SetModEnabled } from '../../wailsjs/go/main/App'

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

export const useModsStore = defineStore('mods', () => {
  const allMods = ref<Mod[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  async function fetchAll(): Promise<void> {
    isLoading.value = true
    error.value = null

    try {
      allMods.value = (await GetAllMods()) as Mod[]
    } catch (err) {
      error.value = errorMessage(err)
      allMods.value = []
    } finally {
      isLoading.value = false
    }
  }

  async function setEnabled(id: string, enabled: boolean): Promise<void> {
    isLoading.value = true
    error.value = null

    try {
      await SetModEnabled(id, enabled)
      const loadOrderStore = useLoadOrderStore()
      await Promise.all([fetchAll(), loadOrderStore.fetch()])
    } catch (err) {
      error.value = errorMessage(err)
    } finally {
      isLoading.value = false
    }
  }

  const enabledMods = computed(() => allMods.value.filter((m) => m.Enabled))

  return { allMods, enabledMods, isLoading, error, fetchAll, setEnabled }
})

