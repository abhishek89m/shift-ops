<script setup lang="ts">
import type { LocaleCode, Task } from '../types';

defineProps<{
  currentTask: Task | null;
  locale: LocaleCode;
  nextTask: Task | null;
  open: boolean;
}>();

const emit = defineEmits<{
  cancel: [];
  confirm: [];
}>();
</script>

<template>
  <div
    v-if="open && currentTask && nextTask"
    class="dialog-backdrop"
  >
    <div class="dialog-panel surface">
      <span class="eyebrow">{{ locale === 'sv' ? 'Bekrafta byte' : 'Confirm switch' }}</span>
      <h2>
        {{
          locale === 'sv'
            ? 'Byt aktiv uppgift?'
            : 'Move the active task back to remaining?'
        }}
      </h2>
      <p class="muted-copy">
        {{
          locale === 'sv'
            ? `Detta flyttar "${currentTask.title}" tillbaka till kvar och markerar "${nextTask.title}" som pagar.`
            : `This moves "${currentTask.title}" back to remaining and starts "${nextTask.title}" instead.`
        }}
      </p>

      <div class="dialog-actions">
        <button
          type="button"
          class="secondary-button"
          @click="emit('cancel')"
        >
          {{ locale === 'sv' ? 'Avbryt' : 'Cancel' }}
        </button>
        <button
          type="button"
          class="primary-button"
          @click="emit('confirm')"
        >
          {{ locale === 'sv' ? 'Byt uppgift' : 'Switch task' }}
        </button>
      </div>
    </div>
  </div>
</template>
