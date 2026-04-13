import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  GetAutoDetectedGameExe,
  GetConfigPath,
  GetGameExe,
  GetModsDirStatus,
  LaunchGame,
  OpenConfigFolder,
  PickExecutable,
  PickFolder,
  ResetModsDirToAuto,
  ResetGameExeToAuto,
  SetGameExe,
  SetModsDir,
} from '../../wailsjs/go/main/App'

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
  const configPath = ref('')
  const isLoaded = ref(false)
  const isLaunching = ref(false)
  const launchError = ref<string | null>(null)
  const lastLaunchAt = ref<number | null>(null)

  const modsDir = computed(() => modsDirStatus.value.effectiveDir)
  const requiresManualPaths = computed(() => isLoaded.value && !modsDirStatus.value.effectiveExists)

  async function fetch(): Promise<void> {
    const [status, exe, autoExe, cfg] = await Promise.all([GetModsDirStatus(), GetGameExe(), GetAutoDetectedGameExe(), GetConfigPath()])
    modsDirStatus.value = (status || emptyStatus) as ModsDirStatus
    gameExe.value = exe || ''
    autoDetectedGameExe.value = autoExe || ''
    configPath.value = cfg || ''
    isLoaded.value = true
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
    configPath,
    isLoaded,
    isLaunching,
    launchError,
    lastLaunchAt,
    requiresManualPaths,
    fetch,
    setModsDir,
    autoDetectModsDir,
    browseModsDir,
    setGameExecutable,
    browseGameExecutable,
    autoDetectGameExecutable,
    openConfigFolder,
    launchGame,
    clearLaunchError,
  }
})
