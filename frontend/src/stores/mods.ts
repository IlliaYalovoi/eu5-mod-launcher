import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useLoadOrderStore } from './loadorder'
import type { Mod, WorkshopItem } from '../types'
import { DisableMod, EnableMod, FetchWorkshopMetadataForMod, GetAllMods } from '../../wailsjs/go/main/App'

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

export const useModsStore = defineStore('mods', () => {
  const allMods = ref<Mod[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const selectedModID = ref<string>('')
  const workshopOpenError = ref<string | null>(null)
  const steamByModID = ref<Record<string, WorkshopItem | null>>({})
  const steamLoadingByModID = ref<Record<string, boolean>>({})
  const steamErrorByModID = ref<Record<string, string | null>>({})
  const metadataRequests = new Map<string, Promise<void>>()

  async function fetchAll(): Promise<void> {
    isLoading.value = true
    error.value = null

    try {
      allMods.value = (await GetAllMods()) as Mod[]
      if (!allMods.value.some((mod) => mod.ID === selectedModID.value)) {
        selectedModID.value = allMods.value[0]?.ID || ''
      }
    } catch (err) {
      error.value = errorMessage(err)
      allMods.value = []
      selectedModID.value = ''
    } finally {
      isLoading.value = false
    }
  }

  function selectMod(id: string): void {
    selectedModID.value = id
    workshopOpenError.value = null
  }

  function setWorkshopOpenError(message: string): void {
    workshopOpenError.value = message
  }

  function clearWorkshopOpenError(): void {
    workshopOpenError.value = null
  }

  async function fetchSteamMetadata(modID: string): Promise<void> {
    if (!modID || steamByModID.value[modID] !== undefined) {
      return
    }
    if (metadataRequests.has(modID)) {
      await metadataRequests.get(modID)
      return
    }

    steamLoadingByModID.value[modID] = true
    steamErrorByModID.value[modID] = null

    const request = (async () => {
      try {
        const item = (await FetchWorkshopMetadataForMod(modID)) as WorkshopItem
        steamByModID.value[modID] = item.itemId ? item : null
      } catch (err) {
        steamErrorByModID.value[modID] = errorMessage(err)
      } finally {
        steamLoadingByModID.value[modID] = false
      }
    })()

    metadataRequests.set(modID, request)
    try {
      await request
    } finally {
      metadataRequests.delete(modID)
    }
  }

  async function setEnabled(id: string, enabled: boolean): Promise<void> {
    isLoading.value = true
    error.value = null

    try {
      if (enabled) {
        await EnableMod(id)
      } else {
        await DisableMod(id)
      }
      const loadOrderStore = useLoadOrderStore()
      await Promise.all([fetchAll(), loadOrderStore.fetch()])
    } catch (err) {
      error.value = errorMessage(err)
    } finally {
      isLoading.value = false
    }
  }

  const enabledMods = computed(() => allMods.value.filter((m) => m.Enabled))
  const selectedMod = computed<Mod | null>(() => allMods.value.find((mod) => mod.ID === selectedModID.value) || null)
  const selectedSteamMetadata = computed<WorkshopItem | null>(() => {
    const id = selectedModID.value
    return id ? steamByModID.value[id] ?? null : null
  })
  const selectedSteamLoading = computed<boolean>(() => {
    const id = selectedModID.value
    return id ? steamLoadingByModID.value[id] ?? false : false
  })
  const selectedSteamError = computed<string | null>(() => {
    const id = selectedModID.value
    return id ? steamErrorByModID.value[id] ?? null : null
  })

  return {
    allMods,
    enabledMods,
    selectedModID,
    selectedMod,
    selectedSteamMetadata,
    selectedSteamLoading,
    selectedSteamError,
    workshopOpenError,
    isLoading,
    error,
    fetchAll,
    setEnabled,
    selectMod,
    fetchSteamMetadata,
    setWorkshopOpenError,
    clearWorkshopOpenError,
  }
})
