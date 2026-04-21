import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type { LauncherCategory, LauncherLayout } from '../types'
import { useSnapshotsStore } from './snapshots'
import {
  Autosort,
  CreateLauncherCategory,
  DeleteLauncherCategory,
  SaveCompiledLoadOrder,
  SetLauncherActivePlaysetIndex,
  SetLauncherLayout,
  SetLoadOrder,
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'

type LoadOrderMutationOptions = {
  scheduleAutosort?: boolean
}

const autosortDebounceMs = 250

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

export const useLoadOrderStore = defineStore('loadorder', () => {
  const snapshotsStore = useSnapshotsStore()

  const autosortError = ref<string | null>(null)
  const isSorting = ref(false)
  const lastSortedAt = ref<number | null>(null)

  let autosortTimer: number | null = null
  let autosortQueued = false
  let autosortRunning = false
  let mutationQueue: Promise<void> = Promise.resolve()

  const orderedIDs = computed(() => snapshotsStore.activeSnapshot?.loadOrder || [])
  const playsetNames = computed(() => snapshotsStore.activeSnapshot?.playsetNames || [])
  const gameActivePlaysetIndex = computed(() => snapshotsStore.activeSnapshot?.gameActivePlaysetIndex ?? -1)
  const launcherActivePlaysetIndex = computed(() => snapshotsStore.activeSnapshot?.launcherActivePlaysetIndex ?? -1)
  const launcherLayout = computed<LauncherLayout>(() => {
    const nextLayout = snapshotsStore.activeSnapshot?.launcherLayout as Partial<LauncherLayout> | undefined
    return {
      ungrouped: nextLayout?.ungrouped || [],
      categories: nextLayout?.categories || [],
      order: nextLayout?.order || ['category:ungrouped'],
      collapsed: nextLayout?.collapsed || {},
    }
  })

  function scheduleAutosort(): void {
    autosortQueued = true
    if (autosortRunning) {
      return
    }
    if (autosortTimer !== null) {
      window.clearTimeout(autosortTimer)
    }
    autosortTimer = window.setTimeout(() => {
      autosortTimer = null
      void flushAutosortQueue()
    }, autosortDebounceMs)
  }

  async function waitForMutationQueue(): Promise<void> {
    while (true) {
      const pending = mutationQueue
      await pending
      if (pending === mutationQueue) {
        return
      }
    }
  }

  async function runAutosortPass(): Promise<void> {
    isSorting.value = true
    try {
      await Autosort()
      await snapshotsStore.refreshActive()
      autosortError.value = null
      lastSortedAt.value = Date.now()
    } catch (err) {
      autosortError.value = errorMessage(err)
    } finally {
      isSorting.value = false
    }
  }

  async function flushAutosortQueue(): Promise<void> {
    if (autosortRunning) {
      return
    }

    autosortRunning = true
    try {
      while (autosortQueued) {
        autosortQueued = false
        await waitForMutationQueue()
        await runAutosortPass()
      }
    } finally {
      autosortRunning = false
    }
  }

  async function runLoadOrderMutation(
    mutation: () => Promise<void>,
    options: LoadOrderMutationOptions = {},
  ): Promise<void> {
    const shouldScheduleAutosort = options.scheduleAutosort ?? true
    const nextMutation = mutationQueue.then(async () => {
      await mutation()
      if (shouldScheduleAutosort) {
        scheduleAutosort()
      }
    })
    mutationQueue = nextMutation.catch(() => undefined)
    return nextMutation
  }

  async function fetch(): Promise<void> {
    await snapshotsStore.refreshActive()
  }

  async function fetchLauncherLayout(): Promise<void> {
    await snapshotsStore.refreshActive()
  }

  async function persist(ids: string[]): Promise<void> {
    await runLoadOrderMutation(async () => {
      await SetLoadOrder(ids)
      await snapshotsStore.refreshActive()
    })
  }

  async function persistLauncherLayout(next: LauncherLayout): Promise<void> {
    await SetLauncherLayout(main.LauncherLayout.createFrom(next))
    await snapshotsStore.refreshActive()
  }

  async function createCategory(name: string): Promise<LauncherCategory> {
    const created = (await CreateLauncherCategory(name)) as LauncherCategory
    await snapshotsStore.refreshActive()
    return created
  }

  async function deleteCategory(categoryID: string): Promise<void> {
    await DeleteLauncherCategory(categoryID)
    await snapshotsStore.refreshActive()
  }

  async function saveCompiled(): Promise<void> {
    await SaveCompiledLoadOrder()
    await snapshotsStore.refreshActive()
  }

  async function autosort(): Promise<void> {
    autosortQueued = true
    if (autosortTimer !== null) {
      window.clearTimeout(autosortTimer)
      autosortTimer = null
    }
    await flushAutosortQueue()
  }

  function clearAutosortError(): void {
    autosortError.value = null
  }

  async function setLauncherPlayset(index: number): Promise<void> {
    await SetLauncherActivePlaysetIndex(index)
    await snapshotsStore.refreshActive()
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
    runLoadOrderMutation,
    autosortError,
    isSorting,
    lastSortedAt,
  }
})
