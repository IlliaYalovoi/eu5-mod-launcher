import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { LauncherCategory, LauncherLayout } from '../types'
import {
  Autosort,
  CreateLauncherCategory,
  DeleteLauncherCategory,
  GetGameActivePlaysetIndex,
  GetLauncherActivePlaysetIndex,
  GetLauncherLayout,
  GetLoadOrder,
  GetPlaysetNames,
  SaveCompiledLoadOrder,
  SetLauncherActivePlaysetIndex,
  SetLauncherLayout,
  SetLoadOrder,
} from '../../wailsjs/go/main/App'

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

const emptyLauncherLayout: LauncherLayout = {
  ungrouped: [],
  categories: [],
  order: ['category:ungrouped'],
  collapsed: {},
}

export const useLoadOrderStore = defineStore('loadorder', () => {
  const orderedIDs = ref<string[]>([])
  const playsetNames = ref<string[]>([])
  const gameActivePlaysetIndex = ref(-1)
  const launcherActivePlaysetIndex = ref(-1)
  const autosortError = ref<string | null>(null)
  const isSorting = ref(false)
  const lastSortedAt = ref<number | null>(null)
  const launcherLayout = ref<LauncherLayout>({ ...emptyLauncherLayout })

  async function fetch(): Promise<void> {
    try {
      const [ids, names, gameIndex, launcherIndex, layout] = await Promise.all([
        GetLoadOrder(),
        GetPlaysetNames(),
        GetGameActivePlaysetIndex(),
        GetLauncherActivePlaysetIndex(),
        GetLauncherLayout(),
      ])
      orderedIDs.value = ids
      playsetNames.value = names
      gameActivePlaysetIndex.value = gameIndex
      launcherActivePlaysetIndex.value = launcherIndex
      launcherLayout.value = (layout || emptyLauncherLayout) as LauncherLayout
    } catch {
      orderedIDs.value = []
      playsetNames.value = []
      gameActivePlaysetIndex.value = -1
      launcherActivePlaysetIndex.value = -1
      launcherLayout.value = { ...emptyLauncherLayout }
    }
  }

  async function fetchLauncherLayout(): Promise<void> {
    launcherLayout.value = (await GetLauncherLayout()) as LauncherLayout
  }

  async function persist(ids: string[]): Promise<void> {
    await SetLoadOrder(ids)
    orderedIDs.value = [...ids]
  }

  async function persistLauncherLayout(next: LauncherLayout): Promise<void> {
    await SetLauncherLayout(next as any)
    launcherLayout.value = next
  }

  async function createCategory(name: string): Promise<LauncherCategory> {
    const created = (await CreateLauncherCategory(name)) as LauncherCategory
    await fetchLauncherLayout()
    return created
  }

  async function deleteCategory(categoryID: string): Promise<void> {
    await DeleteLauncherCategory(categoryID)
    await fetchLauncherLayout()
  }

  async function saveCompiled(): Promise<void> {
    orderedIDs.value = await SaveCompiledLoadOrder()
    await fetchLauncherLayout()
  }

  async function autosort(): Promise<void> {
    isSorting.value = true
    try {
      orderedIDs.value = await Autosort()
      await fetchLauncherLayout()

      autosortError.value = null
      lastSortedAt.value = Date.now()
    } catch (err) {
      autosortError.value = errorMessage(err)
    } finally {
      isSorting.value = false
    }
  }

  function clearAutosortError(): void {
    autosortError.value = null
  }

  async function setLauncherPlayset(index: number): Promise<void> {
    await SetLauncherActivePlaysetIndex(index)
    launcherActivePlaysetIndex.value = index
    orderedIDs.value = await GetLoadOrder()
    await fetchLauncherLayout()
  }

  return {
    orderedIDs,
    playsetNames,
    gameActivePlaysetIndex,
    launcherActivePlaysetIndex,
    launcherLayout,
    fetch,
    fetchLauncherLayout,
    persist,
    persistLauncherLayout,
    createCategory,
    deleteCategory,
    saveCompiled,
    autosort,
    clearAutosortError,
    setLauncherPlayset,
    autosortError,
    isSorting,
    lastSortedAt,
  }
})

