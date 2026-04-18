<script setup lang="ts">
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useSettingsStore } from '../stores/settings'
import { useLoadOrderStore } from '../stores/loadorder'
import { useModsStore } from '../stores/mods'
import GameSettingsModal from './GameSettingsModal.vue'
import LaunchButton from './LaunchButton.vue'

const emit = defineEmits<{
  (event: 'open-settings'): void
}>()

const settingsStore = useSettingsStore()
const loadOrderStore = useLoadOrderStore()
const modsStore = useModsStore()

const { playsetNames, launcherActivePlaysetIndex } = storeToRefs(loadOrderStore)

const activeCountLabel = computed(() => {
  const count = modsStore.enabledMods.length
  return `${count} Active Mod${count === 1 ? '' : 's'}`
})

const gameIcons: Record<string, string> = {
  eu5: '⚜️',
  hoi4: '🎖️',
  ck3: '👑',
  stellaris: '🚀',
  vic3: '🎩',
}

const gameNames: Record<string, string> = {
  eu5: 'Project Caesar',
  hoi4: 'Hearts of Iron IV',
  ck3: 'Crusader Kings III',
  stellaris: 'Stellaris',
  vic3: 'Victoria 3',
}

const gameSettingsModal = ref({
  open: false,
  gameID: '',
})

const isSwitchingPlayset = ref(false)

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

async function onLauncherPlaysetChange(event: Event) {
  const target = event.target as HTMLSelectElement
  const index = parseInt(target.value, 10)
  if (Number.isNaN(index)) return

  isSwitchingPlayset.value = true
  try {
    await loadOrderStore.setLauncherPlayset(index)
  } finally {
    isSwitchingPlayset.value = false
  }
}

const hasPlaysetChoices = computed(() => playsetNames.value.length > 0)
</script>

<template>
  <div class="sidebar-wrapper">
    <div class="sidebar-section header">
      <h2>MOD ORGANIZER</h2>
      <button class="settings-btn" type="button" @click="emit('open-settings')" title="Settings">⚙️</button>
    </div>

    <nav class="game-nav sidebar-section">
      <div
        v-for="gameID in settingsStore.availableGames"
        :key="gameID"
        class="game-item"
      >
        <button
          class="game-btn"
          :class="{ active: settingsStore.activeGameID === gameID }"
          :title="(gameNames[gameID] || gameID.toUpperCase()) + ' (Right click for settings)'"
          @click="selectGame(gameID)"
          @contextmenu.prevent="openGameSettings(gameID)"
        >
          <span class="icon">{{ gameIcons[gameID] || '🎮' }}</span>
          <span>{{ gameNames[gameID] || gameID.toUpperCase() }}</span>
        </button>

        <div v-if="settingsStore.activeGameID === gameID" class="playset-selector">
          <select
            class="playset-dropdown"
            :disabled="!hasPlaysetChoices || isSwitchingPlayset"
            :value="launcherActivePlaysetIndex"
            @change="onLauncherPlaysetChange"
          >
            <option v-for="(name, index) in playsetNames" :key="`${name}-${index}`" :value="index">
              {{ name }}
            </option>
          </select>
        </div>
      </div>

      <GameSettingsModal
        v-if="gameSettingsModal.open"
        :open="gameSettingsModal.open"
        :game-i-d="gameSettingsModal.gameID"
        @close="closeGameSettings"
      />
    </nav>

    <div class="sidebar-footer">
      <div class="play-area">
        <div class="stats-mini">
          <span>{{ activeCountLabel }}</span>
        </div>
        <LaunchButton />
        <div class="status-indicator">
          ● LOAD ORDER READY
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sidebar-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.sidebar-section {
  padding: 20px;
  border-bottom: 1px solid var(--color-border);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header h2 {
  font-size: 1.2rem;
  color: var(--color-accent);
  margin: 0;
  font-family: var(--font-display);
}

.settings-btn {
  background: transparent;
  border: none;
  color: var(--color-text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px;
  transition: color 0.2s;
}

.settings-btn:hover {
  color: var(--color-text-primary);
}

.game-nav {
  flex: 1;
  overflow-y: auto;
}

.game-item {
  margin-bottom: 8px;
}

.game-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  border-radius: 4px;
  text-align: left;
  transition: all 0.2s;
  font-family: var(--font-body);
}

.game-btn.active {
  color: var(--color-text-primary);
  background: var(--color-bg-elevated);
  border-color: var(--color-accent);
  box-shadow: 0 0 10px var(--color-accent-glow);
}

.game-btn .icon {
  font-size: 1.2rem;
}

.playset-selector {
  margin-top: 10px;
  margin-left: 32px;
}

.playset-dropdown {
  width: 100%;
  background: var(--color-bg-base);
  color: var(--color-accent);
  border: 1px solid var(--color-border);
  padding: 8px;
  border-radius: 2px;
  font-size: 0.9rem;
  outline: none;
  font-family: var(--font-body);
}

.sidebar-footer {
  margin-top: auto;
  padding: 20px;
  background: rgba(0,0,0,0.2);
  border-top: 2px solid var(--color-border);
}

.play-area {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.stats-mini {
  font-size: 0.8rem;
  color: var(--color-text-muted);
  display: flex;
  justify-content: space-between;
}

.status-indicator {
  font-size: 0.75rem;
  text-align: center;
  color: var(--color-success);
  margin-top: 5px;
}
</style>
