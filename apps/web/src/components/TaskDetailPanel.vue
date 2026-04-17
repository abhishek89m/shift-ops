<script setup lang="ts">
import { computed } from 'vue';
import { checklistForTask, detailOutcomeCopy, displayTaskId, labelForTaskType } from '../task-ui';
import type { LocaleCode, Task } from '../types';

const props = defineProps<{
  actionBusy: boolean;
  canComplete: boolean;
  checklistState: boolean[];
  isWide: boolean;
  locale: LocaleCode;
  task: Task | null;
}>();

const emit = defineEmits<{
  back: [];
  complete: [task: Task];
  skip: [task: Task];
  start: [task: Task];
  toggleChecklist: [index: number];
}>();

const checklist = computed(() => {
  if (!props.task) {
    return [];
  }

  return checklistForTask(props.task.type, props.locale);
});

const outcomeCopy = computed(() => {
  if (!props.task) {
    return { complete: '', skip: '' };
  }

  return detailOutcomeCopy(props.task, props.locale);
});

const checklistLocked = computed(() => (
  props.actionBusy
  || !props.task
  || props.task.status !== 'in_progress'
));
</script>

<template>
  <article class="detail-shell surface">
    <template v-if="task">
      <div class="detail-head">
        <button
          v-if="!isWide"
          type="button"
          class="back-link"
          @click="emit('back')"
        >
          {{ locale === 'sv' ? 'Tillbaka' : 'Back' }}
        </button>

        <span class="eyebrow">{{ labelForTaskType(task.type, locale) }}</span>
        <h2>{{ task.title }}</h2>
        <p class="muted-copy">
          {{ task.location_label }}
        </p>
        <p class="task-id-copy">
          {{ displayTaskId(task) }}
        </p>
      </div>

      <div class="detail-meta-grid">
        <div class="detail-meta-card">
          <span class="meta-label">{{ locale === 'sv' ? 'Fordon' : 'Vehicle' }}</span>
          <strong>{{ task.vehicle_id }}</strong>
        </div>
        <div class="detail-meta-card">
          <span class="meta-label">{{ locale === 'sv' ? 'Avstand' : 'Distance' }}</span>
          <strong>{{ task.distance_meters }}m</strong>
        </div>
        <div class="detail-meta-card">
          <span class="meta-label">{{ locale === 'sv' ? 'Prioritet' : 'Urgency' }}</span>
          <strong>{{ task.urgency }}/3</strong>
        </div>
        <div
          v-if="task.battery_level !== null && task.battery_level !== undefined"
          class="detail-meta-card"
        >
          <span class="meta-label">{{ locale === 'sv' ? 'Batteri' : 'Battery' }}</span>
          <strong>{{ task.battery_level }}%</strong>
        </div>
      </div>

      <section class="detail-copy-block">
        <span class="meta-label">{{ locale === 'sv' ? 'Faltsammanhang' : 'Field context' }}</span>
        <p>{{ task.notes }}</p>
      </section>

      <section class="detail-copy-block">
        <div class="checklist-head">
          <span class="meta-label">{{ locale === 'sv' ? 'Checklista' : 'Checklist' }}</span>
          <strong>{{ checklistState.filter(Boolean).length }}/{{ checklist.length }}</strong>
        </div>

        <label
          v-for="(item, index) in checklist"
          :key="item"
          class="checklist-row"
        >
          <input
            type="checkbox"
            :checked="checklistState[index] ?? false"
            :disabled="checklistLocked"
            @change="emit('toggleChecklist', index)"
          >
          <span>{{ item }}</span>
        </label>
      </section>

      <section class="detail-copy-block">
        <span class="meta-label">{{ locale === 'sv' ? 'Automatisk logg' : 'Action log' }}</span>
        <p>{{ outcomeCopy.complete }}</p>
        <p>{{ outcomeCopy.skip }}</p>
      </section>

      <footer class="detail-actions">
        <button
          type="button"
          class="secondary-button"
          :disabled="actionBusy || task.status !== 'pending'"
          @click="emit('start', task)"
        >
          {{ locale === 'sv' ? 'Starta' : 'Start' }}
        </button>
        <button
          type="button"
          class="primary-button"
          :disabled="actionBusy || !canComplete"
          @click="emit('complete', task)"
        >
          {{ locale === 'sv' ? 'Klarmarkera' : 'Complete' }}
        </button>
        <button
          type="button"
          class="danger-button"
          :disabled="actionBusy || task.status !== 'in_progress'"
          @click="emit('skip', task)"
        >
          {{ locale === 'sv' ? 'Hoppa over' : 'Skip' }}
        </button>
      </footer>
    </template>

    <template v-else>
      <div class="detail-empty">
        <span class="eyebrow">{{ locale === 'sv' ? 'Detaljer' : 'Detail' }}</span>
        <h2>{{ locale === 'sv' ? 'Valj en uppgift.' : 'Choose a task.' }}</h2>
        <p class="muted-copy">
          {{
            locale === 'sv'
              ? 'Nar du valjer en uppgift visas arbetssteg, sammanhang och handlingar har.'
              : 'Once you choose a task, its context, checklist, and actions will appear here.'
          }}
        </p>
      </div>
    </template>
  </article>
</template>
