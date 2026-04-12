import { ref } from 'vue'
import { defineStore } from 'pinia'
import { logger } from '../lib/logger'
import {
  ListSupportedGames,
  GetGameExe,
} from '../../wailsjs/go/launcher/App'

export const useGamesStore = defineStore('games', () => {
  const supportedGames = ref<Array<{ id: string; name: string; detected: boolean }>>([])
  const activeGameID = ref<string>('')

  async function fetchSupportedGames(): Promise<void> {
    logger.debug('games', 'Fetching supported games')
    const games = await ListSupportedGames()
    logger.info('games', `Found ${games.length} supported games`, { games })
    supportedGames.value = games.map((game: any) => ({
      id: game.id,
      name: game.name,
      detected: game.detected,
    }))

    // Set active game to first detected game if none set
    if (!activeGameID.value && supportedGames.value.length > 0) {
      const detectedGame = supportedGames.value.find(g => g.detected)
      if (detectedGame) {
        activeGameID.value = detectedGame.id
      } else {
        activeGameID.value = supportedGames.value[0].id
      }
    }
  }

  function setActiveGame(gameID: string): void {
    activeGameID.value = gameID
  }

  async function initialize(): Promise<void> {
    await fetchSupportedGames()
    // Sync with backend
    const exe = await GetGameExe()
    if (exe) {
      // Find game matching current exe and set as active
      const matchingGame = supportedGames.value.find(game =>
        game.detected && game.id === activeGameID.value
      )
      if (!matchingGame) {
        // If current active game doesn't match exe, find the game that does
        const exeGame = supportedGames.value.find(game => game.detected)
        if (exeGame) {
          activeGameID.value = exeGame.id
        }
      }
    }
  }

  return {
    supportedGames,
    activeGameID,
    fetchSupportedGames,
    setActiveGame,
    initialize,
  }
})