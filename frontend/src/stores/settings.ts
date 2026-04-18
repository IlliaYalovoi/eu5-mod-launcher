import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useSnapshotsStore } from './snapshots'
import type { SnapshotModsDirStatus } from '../types'
import {
  LaunchGame,
  OpenConfigFolder,
  PickExecutable,
  PickFolder,
  ResetModsDirToAuto,
  ResetGameExeToAuto,
  SetGameExe,
  SetGameVersionOverride,
  SetModsDir,
} from '../../wailsjs/go/main/App'

const emptyStatus: SnapshotModsDirStatus = {
  effectiveDir: '',
  autoDetectedDir: '',
  customDir: '',
  usingCustomDir: false,
  autoDetectedExists: false,
  effectiveExists: false,
}

export const useSettingsStore = defineStore('settings', () => {
  const snapshotsStore = useSnapshotsStore()

  const isLaunching = ref(false)
  const launchError = ref<string | null>(null)
  const lastLaunchAt = ref<number | null>(null)

  const activeSnapshot = computed(() => snapshotsStore.activeSnapshot)
  const modsDirStatus = computed<SnapshotModsDirStatus>(() => activeSnapshot.value?.settings.modsDirStatus || emptyStatus)
  const gameExe = computed(() => activeSnapshot.value?.settings.gameExe || '')
  const autoDetectedGameExe = computed(() => activeSnapshot.value?.settings.autoDetectedGameExe || '')
  const gameVersion = computed(() => activeSnapshot.value?.settings.gameVersion || 'unknown')
  const gameVersionOverride = computed(() => activeSnapshot.value?.settings.gameVersionOverride || '')
  const configPath = computed(() => activeSnapshot.value?.settings.configPath || '')
  const activeGameID = computed(() => activeSnapshot.value?.gameID || 'eu5')
  const availableGames = computed(() => activeSnapshot.value?.settings.availableGames || [])

  const isLoaded = computed(() => snapshotsStore.startupState === 'ready')
  const modsDir = computed(() => modsDirStatus.value.effectiveDir)
  const requiresManualPaths = computed(() => isLoaded.value && !modsDirStatus.value.effectiveExists)

  async function fetch(): Promise<void> {
    if (!snapshotsStore.activeSnapshot) {
      await snapshotsStore.bootstrap()
      return
    }
    await snapshotsStore.refreshActive()
  }

  async function setGame(id: string): Promise<void> {
    await snapshotsStore.switchGame(id)
  }

  async function setModsDir(path: string): Promise<void> {
    await SetModsDir(path)
    await snapshotsStore.refreshActive()
  }

  async function autoDetectModsDir(): Promise<void> {
    await ResetModsDirToAuto()
    await snapshotsStore.refreshActive()
  }

  async function browseModsDir(): Promise<void> {
    const picked = await PickFolder()
    if (!picked) {
      return
    }
    await setModsDir(picked)
  }

  async function setGameExecutable(path: string): Promise<void> {
    await SetGameExe(path)
    await snapshotsStore.refreshActive()
  }

  async function browseGameExecutable(): Promise<void> {
    const picked = await PickExecutable()
    if (!picked) {
      return
    }
    await setGameExecutable(picked)
  }

  async function autoDetectGameExecutable(): Promise<void> {
    await ResetGameExeToAuto()
    await snapshotsStore.refreshActive()
  }

  async function setGameVersionOverride(version: string): Promise<void> {
    await SetGameVersionOverride(version)
    await snapshotsStore.refreshActive()
  }

  async function openConfigFolder(): Promise<void> {
    await OpenConfigFolder()
  }

  async function launchGame(): Promise<void> {
    isLaunching.value = true
    launchError.value = null
    try {
      await LaunchGame()
      lastLaunchAt.value = Date.now()
    } catch (err) {
      launchError.value = err instanceof Error ? err.message : String(err)
    } finally {
      isLaunching.value = false
    }
  }

  function clearLaunchError(): void {
    launchError.value = null
  }

  return {
    modsDir,
    modsDirStatus,
    gameExe,
    autoDetectedGameExe,
    gameVersion,
    gameVersionOverride,
    configPath,
    isLoaded,
    isLaunching,
    launchError,
    lastLaunchAt,
    requiresManualPaths,
    activeGameID,
    availableGames,
    fetch,
    setGame,
    setModsDir,
    autoDetectModsDir,
    browseModsDir,
    setGameExecutable,
    browseGameExecutable,
    autoDetectGameExecutable,
    setGameVersionOverride,
    openConfigFolder,
    launchGame,
    clearLaunchError,
  }
})
