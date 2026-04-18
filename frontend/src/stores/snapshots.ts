import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type { GameSnapshot } from '../types'
import {
  GetActiveGameID,
  GetGameSnapshot,
  SetActiveGameAndGetSnapshot,
  WarmNonActiveGameSnapshots,
} from '../../wailsjs/go/main/App'

type StartupState = 'cold' | 'warming' | 'ready'
type SwitchState = 'idle' | 'loading' | 'committing'

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

export const useSnapshotsStore = defineStore('snapshots', () => {
  const snapshotsByGameID = ref<Record<string, GameSnapshot>>({})
  const activeGameID = ref('')
  const visibleGameID = ref('')
  const startupState = ref<StartupState>('cold')
  const switchState = ref<SwitchState>('idle')
  const switchError = ref<string | null>(null)
  const lastFailedGameID = ref('')
  const latestRequestID = ref(0)
  const warmTimer = ref<number | null>(null)

  function applySnapshot(next: GameSnapshot): void {
    const current = snapshotsByGameID.value[next.gameID]
    if (current && next.meta.revision < current.meta.revision) {
      return
    }

    snapshotsByGameID.value = {
      ...snapshotsByGameID.value,
      [next.gameID]: next,
    }
  }

  async function bootstrap(): Promise<void> {
    startupState.value = 'cold'

    const active = await GetActiveGameID()
    const snapshot = (await GetGameSnapshot(active || '')) as GameSnapshot
    applySnapshot(snapshot)

    activeGameID.value = snapshot.gameID
    visibleGameID.value = snapshot.gameID
    startupState.value = 'warming'

    void warmNonActive()
    startWarmLoop()
    startupState.value = 'ready'
  }

  async function switchGame(gameID: string): Promise<void> {
    if (!gameID || switchState.value !== 'idle' || gameID === activeGameID.value) {
      return
    }

    const requestID = latestRequestID.value + 1
    latestRequestID.value = requestID
    switchState.value = 'loading'
    switchError.value = null
    lastFailedGameID.value = ''

    try {
      const snapshot = (await SetActiveGameAndGetSnapshot(gameID)) as GameSnapshot
      if (requestID !== latestRequestID.value) {
        return
      }

      applySnapshot(snapshot)
      activeGameID.value = snapshot.gameID
      switchState.value = 'committing'

      window.setTimeout(() => {
        if (requestID !== latestRequestID.value) {
          return
        }
        visibleGameID.value = snapshot.gameID
        switchState.value = 'idle'
      }, 220)
    } catch (err) {
      if (requestID === latestRequestID.value) {
        switchState.value = 'idle'
        switchError.value = errorMessage(err)
        lastFailedGameID.value = gameID
      }
    }
  }

  async function refreshActive(): Promise<void> {
    const gameID = activeGameID.value || visibleGameID.value
    if (!gameID) {
      return
    }

    const snapshot = (await GetGameSnapshot(gameID)) as GameSnapshot
    applySnapshot(snapshot)
  }

  async function warmNonActive(): Promise<void> {
    const result = (await WarmNonActiveGameSnapshots()) as Record<string, GameSnapshot>
    for (const key of Object.keys(result)) {
      applySnapshot(result[key])
    }
  }

  function startWarmLoop(): void {
    if (warmTimer.value !== null) {
      window.clearInterval(warmTimer.value)
    }

    warmTimer.value = window.setInterval(() => {
      void warmNonActive()
    }, 7 * 60 * 1000)
  }

  function clearSwitchError(): void {
    switchError.value = null
    lastFailedGameID.value = ''
  }

  async function retryLastSwitch(): Promise<void> {
    if (!lastFailedGameID.value || switchState.value !== 'idle') {
      return
    }
    await switchGame(lastFailedGameID.value)
  }

  const visibleSnapshot = computed<GameSnapshot | null>(() => snapshotsByGameID.value[visibleGameID.value] || null)
  const activeSnapshot = computed<GameSnapshot | null>(() => {
    if (switchState.value !== 'idle') {
      return visibleSnapshot.value
    }
    return snapshotsByGameID.value[activeGameID.value] || visibleSnapshot.value || null
  })
  const hasColdStart = computed(() => startupState.value === 'cold' || !visibleSnapshot.value)

  return {
    snapshotsByGameID,
    activeGameID,
    visibleGameID,
    startupState,
    switchState,
    switchError,
    visibleSnapshot,
    activeSnapshot,
    hasColdStart,
    bootstrap,
    switchGame,
    refreshActive,
    warmNonActive,
    startWarmLoop,
    clearSwitchError,
    retryLastSwitch,
  }
})
