<script setup lang="ts">
import { computed } from 'vue';
import { displayTaskId, labelForStatus, labelForTaskType } from '../task-ui';
import type { LocaleCode, Recommendation, Task } from '../types';

interface TaskSection {
  id: 'in_progress' | 'pending' | 'done';
  label: string;
  tasks: Task[];
}

const props = defineProps<{
  locale: LocaleCode;
  recommended: Recommendation | null;
  sections: TaskSection[];
  selectedTaskId: string | null;
}>();

const emit = defineEmits<{
  select: [task: Task];
}>();

const recommendedTaskId = computed(() => props.recommended?.task.id ?? null);

function sectionTone(id: TaskSection['id']) {
  switch (id) {
    case 'in_progress':
      return 'status-tone-in-progress';
    case 'done':
      return 'status-tone-done';
    case 'pending':
    default:
      return 'status-tone-pending';
  }
}

function taskTone(status: Task['status']) {
  switch (status) {
    case 'in_progress':
      return 'status-tone-in-progress';
    case 'completed':
      return 'status-tone-done';
    case 'skipped':
      return 'status-tone-skipped';
    case 'pending':
    default:
      return 'status-tone-pending';
  }
}
</script>

<template>
  <section class="task-list-shell">
    <article
      v-if="recommended"
      class="recommended-card"
    >
      <div class="recommended-head">
        <div class="recommended-copy">
          <span class="eyebrow">{{ locale === 'sv' ? 'Rekommenderad' : 'Recommended' }}</span>
          <h2>{{ recommended.task.title }}</h2>
          <p class="muted-copy">
            {{ recommended.task.location_label }}
          </p>
          <p class="task-id-copy">
            {{ displayTaskId(recommended.task) }}
          </p>
        </div>
        <button
          type="button"
          class="primary-button"
          @click="emit('select', recommended.task)"
        >
          {{ locale === 'sv' ? 'Oppna uppgift' : 'Open task' }}
        </button>
      </div>

      <div class="chip-row">
        <span
          v-for="reason in recommended.reasons"
          :key="reason"
          class="chip"
        >
          {{ reason }}
        </span>
      </div>
    </article>

    <section
      v-for="section in sections"
      :key="section.id"
      class="task-section"
    >
      <header class="task-section-head">
        <h3>{{ section.label }}</h3>
        <span
          class="section-count"
          :class="sectionTone(section.id)"
        >
          {{ section.tasks.length }}
        </span>
      </header>

      <div
        v-if="section.tasks.length === 0"
        class="task-empty"
      >
        {{ locale === 'sv' ? 'Inga uppgifter i denna sektion.' : 'No tasks in this section.' }}
      </div>

      <button
        v-for="task in section.tasks"
        :key="task.id"
        type="button"
        class="task-row"
        :class="{
          'is-selected': selectedTaskId === task.id,
          'is-recommended': recommendedTaskId === task.id
        }"
        @click="emit('select', task)"
      >
        <div class="task-row-head">
          <strong>{{ task.title }}</strong>
          <span
            class="status-pill"
            :class="taskTone(task.status)"
          >
            {{ labelForStatus(task.status, locale) }}
          </span>
        </div>

        <div class="task-row-id">
          {{ displayTaskId(task) }}
        </div>

        <div class="task-row-meta">
          <span>{{ labelForTaskType(task.type, locale) }}</span>
          <span>{{ task.location_label }}</span>
        </div>

        <div class="task-row-foot">
          <span>{{ task.distance_meters }}m</span>
          <span v-if="task.battery_level !== null && task.battery_level !== undefined">{{ task.battery_level }}%</span>
          <span>{{ locale === 'sv' ? 'Prio' : 'Urgency' }} {{ task.urgency }}</span>
        </div>
      </button>
    </section>
  </section>
</template>
