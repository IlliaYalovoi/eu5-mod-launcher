// frontend/src/stores/activeGame.ts
import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useActiveGameStore = defineStore('activeGame', () => {
  const activeGameID = ref('eu5')
  const theme = ref('caesar')

  watch(activeGameID, (id) => {
    theme.value = id === 'eu5' ? 'caesar' : 'victoria'
    document.body.className = `theme-${theme.value}`
  }, { immediate: true })

  return { activeGameID, theme }
})
