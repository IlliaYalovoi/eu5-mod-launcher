<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useLoadOrderStore } from '../stores/loadorder'
import BaseModal from './ui/BaseModal.vue'
import BaseButton from './ui/BaseButton.vue'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (event: 'close'): void
}>()

const loadOrderStore = useLoadOrderStore()
const { launcherLayout } = storeToRefs(loadOrderStore)

const newCategoryName = ref('')
const error = ref('')

async function createCategory() {
  if (!newCategoryName.value.trim()) return
  try {
    await loadOrderStore.createCategory(newCategoryName.value.trim())
    newCategoryName.value = ''
  } catch (e: any) {
    error.value = e.message
  }
}

async function deleteCategory(id: string) {
  try {
    await loadOrderStore.deleteCategory(id)
  } catch (e: any) {
    error.value = e.message
  }
}
</script>

<template>
  <BaseModal :open="open" @close="emit('close')">
    <div class="manage-groups-modal">
      <h2>Manage Mod Groups</h2>

      <div class="create-form">
        <input v-model="newCategoryName" type="text" placeholder="New group name..." class="input" @keyup.enter="createCategory" />
        <BaseButton @click="createCategory">Add Group</BaseButton>
      </div>

      <p v-if="error" class="error">{{ error }}</p>

      <ul class="group-list">
        <li v-for="category in launcherLayout.categories" :key="category.id" class="group-item">
          <span>{{ category.name }}</span>
          <button class="delete-btn" type="button" @click="deleteCategory(category.id)">Delete</button>
        </li>
      </ul>

      <div class="actions">
        <BaseButton variant="ghost" @click="emit('close')">Done</BaseButton>
      </div>
    </div>
  </BaseModal>
</template>

<style scoped>
.manage-groups-modal {
  background: var(--color-bg-panel);
  padding: 30px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
  width: 400px;
}

h2 {
  margin-top: 0;
  color: var(--color-accent);
  font-family: var(--font-display);
}

.create-form {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.input {
  flex: 1;
  background: var(--color-bg-base);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  padding: 8px;
  border-radius: 4px;
}

.group-list {
  list-style: none;
  padding: 0;
  margin: 0 0 20px 0;
  max-height: 300px;
  overflow-y: auto;
}

.group-item {
  display: flex;
  justify-content: space-between;
  padding: 10px;
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
}

.delete-btn {
  background: transparent;
  border: none;
  color: var(--color-danger);
  cursor: pointer;
}

.actions {
  display: flex;
  justify-content: flex-end;
}

.error {
  color: var(--color-danger);
  font-size: 13px;
  margin-bottom: 10px;
}
</style>