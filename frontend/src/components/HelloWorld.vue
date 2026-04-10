<script lang="ts" setup>
import { reactive } from 'vue'

const data = reactive({
  name: '',
  resultText: 'Data layer initialized. UI tasks will replace this template.',
})
</script>

<template>
  <main>
    <div id="result" class="result">{{ data.resultText }}</div>
    <div id="input" class="input-box">
      <input id="name" v-model="data.name" autocomplete="off" class="input" type="text" />
      <button class="btn" type="button" disabled>Greet</button>
    </div>
  </main>
</template>

<style scoped>
.result {
  height: var(--size-legacy-result-height);
  line-height: var(--size-legacy-result-height);
  margin: 1.5rem auto;
}

.input-box .btn {
  width: var(--size-legacy-btn-width);
  height: var(--size-legacy-control-height);
  line-height: var(--size-legacy-control-height);
  border-radius: var(--radius-legacy-sm);
  border: none;
  margin: 0 0 0 var(--space-5);
  padding: 0 var(--space-2);
  cursor: pointer;
}

.input-box .btn:hover {
  background-image: var(--gradient-legacy-btn-hover);
  color: var(--color-legacy-text-dark);
}

.input-box .input {
  border: none;
  border-radius: var(--radius-legacy-sm);
  outline: none;
  height: var(--size-legacy-control-height);
  line-height: var(--size-legacy-control-height);
  padding: 0 var(--space-legacy-input-x);
  background-color: var(--color-legacy-input-bg);
  -webkit-font-smoothing: antialiased;
}

.input-box .input:hover {
  border: none;
  background-color: var(--color-legacy-input-bg-hover);
}

.input-box .input:focus {
  border: none;
  background-color: var(--color-legacy-input-bg-hover);
}
</style>
