<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void
}>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>

<template>
  <div class="toggle" :class="{ on: modelValue }" @click.stop="toggle"></div>
</template>

<style scoped>
.toggle {
  width: 36px;
  height: 18px;
  background: #444;
  border-radius: 10px;
  position: relative;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.2s;
}
.toggle.on {
  background: var(--color-success);
}
.toggle::after {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  background: white;
  border-radius: 50%;
  top: 2px;
  left: 2px;
  transition: 0.2s;
}
.toggle.on::after {
  left: 20px;
}
</style>
