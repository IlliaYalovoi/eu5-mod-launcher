import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useModsStore } from './mods'
import { useLoadOrderStore } from './loadorder'
import { logBackendCall } from '../utils/backendDebug'
import {
  GetActiveGameID,
  GetAutoDetectedGameExe,
  GetAvailableGames,
  GetConfigPath,
  GetGameExe,
  GetModsDirStatus,
  GetGameVersion,
  GetGameVersionOverride,
  SetGameVersionOverride,
  LaunchGame,
  OpenConfigFolder,
  PickExecutable,
  PickFolder,
  ResetModsDirToAuto,
  ResetGameExeToAuto,
  SetActiveGame,
  SetGameExe,
  SetModsDir,
} from '../../wailsjs/go/main/App'

type GameMetadata = {
  ID: () => string
}

type ModsDirStatus = {
  effectiveDir: string
  autoDetectedDir: string
  customDir: string
  usingCustomDir: boolean
  autoDetectedExists: boolean
  effectiveExists: boolean
}

const emptyStatus: ModsDirStatus = {
  effectiveDir: '',
  autoDetectedDir: '',
  customDir: '',
  usingCustomDir: false,
  autoDetectedExists: false,
  effectiveExists: false,
}

export const useSettingsStore = defineStore('settings', () => {
  const modsDirStatus = ref<ModsDirStatus>({ ...emptyStatus })
  const gameExe = ref('')
  const autoDetectedGameExe = ref('')
  const gameVersion = ref('unknown')
  const gameVersionOverride = ref('')
  const configPath = ref('')
  const isLoaded = ref(false)
  const isLaunching = ref(false)
  const launchError = ref<string | null>(null)
  const lastLaunchAt = ref<number | null>(null)
  const activeGameID = ref('eu5')
  const availableGames = ref<string[]>([])

  const modsDir = computed(() => modsDirStatus.value.effectiveDir)
  const requiresManualPaths = computed(() => isLoaded.value && !modsDirStatus.value.effectiveExists)

  async function fetch(): Promise<void> {
    const [status, exe, autoExe, cfg, activeID, games, ver, verOverride] = await Promise.all([
      logBackendCall('GetModsDirStatus', [], () => GetModsDirStatus()),
      logBackendCall('GetGameExe', [], () => GetGameExe()),
      logBackendCall('GetAutoDetectedGameExe', [], () => GetAutoDetectedGameExe()),
      logBackendCall('GetConfigPath', [], () => GetConfigPath()),
      logBackendCall('GetActiveGameID', [], () => GetActiveGameID()),
      logBackendCall('GetAvailableGames', [], () => GetAvailableGames()),
      logBackendCall('GetGameVersion', [], () => GetGameVersion()),
      logBackendCall('GetGameVersionOverride', [], () => GetGameVersionOverride()),
    ])
    modsDirStatus.value = (status || emptyStatus) as ModsDirStatus
    gameExe.value = exe || ''
    autoDetectedGameExe.value = autoExe || ''
    configPath.value = cfg || ''
    activeGameID.value = activeID || 'eu5'
    availableGames.value = (games as string[])
    gameVersion.value = ver || 'unknown'
    gameVersionOverride.value = verOverride || ''
    isLoaded.value = true
  }

  async function setGame(id: string): Promise<void> {
    await logBackendCall('SetActiveGame', [id], () => SetActiveGame(id))
    const modsStore = useModsStore()
    const loadOrderStore = useLoadOrderStore()
    await Promise.all([fetch(), modsStore.fetchAll(), loadOrderStore.fetch()])
  }

  async function setModsDir(path: string): Promise<void> {
    await SetModsDir(path)
    modsDirStatus.value = (await GetModsDirStatus()) as ModsDirStatus
  }

  async function autoDetectModsDir(): Promise<void> {
    await ResetModsDirToAuto()
    modsDirStatus.value = (await GetModsDirStatus()) as ModsDirStatus
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
    gameExe.value = path
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
    gameExe.value = await GetGameExe()
    autoDetectedGameExe.value = await GetAutoDetectedGameExe()
  }

  async function setGameVersionOverride(version: string): Promise<void> {
    await SetGameVersionOverride(version)
    gameVersionOverride.value = version
    gameVersion.value = await GetGameVersion()
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
