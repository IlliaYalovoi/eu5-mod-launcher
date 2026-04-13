<script setup lang="ts">
import { ref } from 'vue'
import { useSettingsStore } from '../stores/settings'
import GameSettingsModal from './GameSettingsModal.vue'

const settingsStore = useSettingsStore()

const gameIcons: Record<string, string> = {
  eu5: '🌍',
  hoi4: '⚔️',
  ck3: '👑',
  stellaris: '🚀',
  vic3: '📜',
}

const gameNames: Record<string, string> = {
  eu5: 'Europa Universalis V',
  hoi4: 'Hearts of Iron IV',
  ck3: 'Crusader Kings III',
  stellaris: 'Stellaris',
  vic3: 'Victoria 3',
}

const gameSettingsModal = ref({
  open: false,
  gameID: '',
})

function selectGame(id: string) {
  settingsStore.setGame(id)
}

async function openGameSettings(id: string) {
  await settingsStore.setGame(id)
  gameSettingsModal.value = {
    open: true,
    gameID: id,
  }
}

function closeGameSettings() {
  gameSettingsModal.value.open = false
}
</script>

<template>
  <nav class="game-sidebar">
    <button
      v-for="gameID in settingsStore.availableGames"
      :key="gameID"
      class="game-icon"
      :class="{ active: settingsStore.activeGameID === gameID }"
      :title="(gameNames[gameID] || gameID.toUpperCase()) + ' (Right click for settings)'"
      @click="selectGame(gameID)"
      @contextmenu.prevent="openGameSettings(gameID)"
    >
      {{ gameIcons[gameID] || '🎮' }}
    </button>

    <GameSettingsModal
      v-if="gameSettingsModal.open"
      :open="gameSettingsModal.open"
      :game-i-d="gameSettingsModal.gameID"
      @close="closeGameSettings"
    />
  </nav>
</template>

<style scoped>
.game-sidebar {
  width: 4rem;
  background: var(--color-bg-sidebar);
  border-right: var(--border-width) solid var(--color-border);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--space-4) 0;
  gap: var(--space-4);
}

.game-icon {
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-panel);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: 1.25rem;
  cursor: pointer;
  transition: all 0.2s;
}

.game-icon:hover {
  border-color: var(--color-accent);
  background: var(--color-bg-elevated);
}

.game-icon.active {
  background: var(--color-accent);
  border-color: var(--color-accent);
  color: #000;
  box-shadow: 0 0 10px var(--color-accent);
}
</style>
