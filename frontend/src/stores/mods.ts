import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { logger } from '../lib/logger'
import { useLoadOrderStore } from './loadorder'
import type { Mod, WorkshopItem } from '../types'
import {
  DisableMod,
  EnableMod,
  FetchWorkshopMetadataForMod,
  GetAllMods,
  IsUnsubscribeEnabled,
  UnsubscribeWorkshopMod,
} from '../../wailsjs/go/main/App'

type UnsubscribeNotice = {
  type: 'success' | 'error'
  message: string
}

const steamAppID = '3450310'

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
  const unsubscribeLoadingByModID = ref<Record<string, boolean>>({})
  const unsubscribeErrorByModID = ref<Record<string, string | null>>({})
  const unsubscribeNotice = ref<UnsubscribeNotice | null>(null)
  const unsubscribeFeatureEnabled = ref(false)
  const unsubscribeCapabilityLoaded = ref(false)
  const metadataRequests = new Map<string, Promise<void>>()
  const unsubscribeRequests = new Map<string, Promise<void>>()

  logger.debug('mods', 'Mods store initialized')

  async function ensureUnsubscribeCapability(): Promise<void> {
    if (unsubscribeCapabilityLoaded.value) {
      return
    }
    unsubscribeFeatureEnabled.value = await IsUnsubscribeEnabled()
    unsubscribeCapabilityLoaded.value = true
    logger.debug('mods', `Unsubscribe capability loaded: ${unsubscribeFeatureEnabled.value}`)
  }

  function modByID(id: string): Mod | null {
    return allMods.value.find((mod) => mod.ID === id) || null
  }

  function workshopItemIDFromDirPath(dirPath: string): string {
    const normalized = dirPath.trim().replace(/\\+/g, '/').replace(/\/+/g, '/')
    if (!normalized) {
      return ''
    }

    const steamContentPattern = new RegExp(`/workshop/content/${steamAppID}/(\\d+)(?:/|$)`, 'i')
    const match = normalized.match(steamContentPattern)
    if (match && match[1]) {
      return match[1]
    }

    return ''
  }

  function workshopItemIDForMod(modID: string): string {
    if (!modID) {
      return ''
    }
    const metadataItemID = steamByModID.value[modID]?.itemId?.trim() || ''
    if (metadataItemID) {
      return metadataItemID
    }
    const mod = modByID(modID)
    if (!mod) {
      return ''
    }
    return workshopItemIDFromDirPath(mod.DirPath || '')
  }

  function isWorkshopMod(modID: string): boolean {
    return workshopItemIDForMod(modID) !== ''
  }

  function isUnsubscribeLoading(modID: string): boolean {
    return modID ? unsubscribeLoadingByModID.value[modID] ?? false : false
  }

  async function fetchAll(): Promise<void> {
    isLoading.value = true
    error.value = null
    logger.debug('mods', 'Fetching all mods')

    try {
      await ensureUnsubscribeCapability()
      allMods.value = (await GetAllMods()) as Mod[]
      logger.info('mods', `Fetched ${allMods.value.length} mods`)
      if (!allMods.value.some((mod) => mod.ID === selectedModID.value)) {
        selectedModID.value = allMods.value[0]?.ID || ''
      }
    } catch (err) {
      error.value = errorMessage(err)
      logger.error('mods', `Fetch failed: ${errorMessage(err)}`)
      allMods.value = []
      selectedModID.value = ''
    } finally {
      isLoading.value = false
    }
  }

  function selectMod(id: string): void {
    selectedModID.value = id
    workshopOpenError.value = null
    logger.debug('mods', `Selected mod: ${id}`)
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

  async function ensureWorkshopMetadata(modID: string): Promise<void> {
    if (!modID || workshopItemIDForMod(modID)) {
      return
    }
    await fetchSteamMetadata(modID)
  }

  async function unsubscribeWorkshop(modID: string): Promise<void> {
    if (!modID) {
      return
    }
    await ensureUnsubscribeCapability()
    if (!unsubscribeFeatureEnabled.value) {
      const disabledMessage = 'Unsubscribe is currently disabled in this build.'
      unsubscribeErrorByModID.value[modID] = disabledMessage
      unsubscribeNotice.value = { type: 'error', message: disabledMessage }
      return
    }
    if (unsubscribeRequests.has(modID)) {
      await unsubscribeRequests.get(modID)
      return
    }

    unsubscribeLoadingByModID.value[modID] = true
    unsubscribeErrorByModID.value[modID] = null

    const request = (async () => {
      try {
        await ensureWorkshopMetadata(modID)
        const itemID = workshopItemIDForMod(modID)
        if (!itemID) {
          throw new Error('Selected mod is not linked to Steam Workshop.')
        }

        await UnsubscribeWorkshopMod(itemID)
        const loadOrderStore = useLoadOrderStore()
        await Promise.all([fetchAll(), loadOrderStore.fetch()])
        unsubscribeNotice.value = { type: 'success', message: 'Unsubscribe request sent to Steam.' }
      } catch (err) {
        const message = errorMessage(err)
        unsubscribeErrorByModID.value[modID] = message
        unsubscribeNotice.value = { type: 'error', message }
        throw err
      } finally {
        unsubscribeLoadingByModID.value[modID] = false
      }
    })()

    unsubscribeRequests.set(modID, request)
    try {
      await request
    } finally {
      unsubscribeRequests.delete(modID)
    }
  }

  function clearUnsubscribeNotice(): void {
    unsubscribeNotice.value = null
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
  const selectedUnsubscribeLoading = computed<boolean>(() => {
    const id = selectedModID.value
    return id ? unsubscribeLoadingByModID.value[id] ?? false : false
  })
  const selectedUnsubscribeError = computed<string | null>(() => {
    const id = selectedModID.value
    return id ? unsubscribeErrorByModID.value[id] ?? null : null
  })

  return {
    allMods,
    enabledMods,
    selectedModID,
    selectedMod,
    selectedSteamMetadata,
    selectedSteamLoading,
    selectedSteamError,
    selectedUnsubscribeLoading,
    selectedUnsubscribeError,
    workshopOpenError,
    unsubscribeNotice,
    unsubscribeFeatureEnabled,
    isLoading,
    error,
    fetchAll,
    setEnabled,
    selectMod,
    fetchSteamMetadata,
    ensureWorkshopMetadata,
    workshopItemIDForMod,
    isWorkshopMod,
    isUnsubscribeLoading,
    unsubscribeWorkshop,
    setWorkshopOpenError,
    clearWorkshopOpenError,
    clearUnsubscribeNotice,
  }
})
