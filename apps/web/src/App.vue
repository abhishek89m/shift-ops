<script setup lang="ts">
import ProgressShell from './components/ProgressShell.vue';

const completed = 2;
const total = 4;
</script>

<template>
  <ProgressShell
    :completed="completed"
    :total="total"
  />
</template>
