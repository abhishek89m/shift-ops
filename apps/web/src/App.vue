<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import ProgressShell from './components/ProgressShell.vue';
import TaskDetailPanel from './components/TaskDetailPanel.vue';
import TaskListPanel from './components/TaskListPanel.vue';
import TaskSwitchDialog from './components/TaskSwitchDialog.vue';
import { useShiftOps } from './composables/useShiftOps';
import { checklistForTask, resolutionForAction } from './task-ui';
import type { LocaleCode, Task } from './types';

const splitQuery = typeof window !== 'undefined' ? window.matchMedia('(min-width: 768px)') : null;

const selectedTaskId = ref<string | null>(null);
const checklistState = ref<Record<string, boolean[]>>({});
const isWide = ref(splitQuery?.matches ?? false);
const isCondensed = ref(false);
const switchCandidate = ref<Task | null>(null);
const { locale: i18nLocale } = useI18n();

const workerHandle = 'field-worker';

const {
  activeTask,
  actionBusyId,
  error,
  loading,
  patchTask,
  refresh,
  summary,
  tasks
} = useShiftOps();

const locale = computed<LocaleCode>(() => (i18nLocale.value.startsWith('sv') ? 'sv' : 'en'));

const selectedTask = computed(() => {
  if (selectedTaskId.value) {
    return tasks.value.find((task) => task.id === selectedTaskId.value) ?? null;
  }

  if (isWide.value) {
    return preferredTask(tasks.value);
  }

  return null;
});

const taskSections = computed(() => [
  {
    id: 'in_progress' as const,
    label: locale.value === 'sv' ? 'Pagar' : 'In progress',
    tasks: tasks.value.filter((task) => task.status === 'in_progress')
  },
  {
    id: 'pending' as const,
    label: locale.value === 'sv' ? 'Kvar' : 'Remaining',
    tasks: tasks.value.filter((task) => task.status === 'pending')
  },
  {
    id: 'done' as const,
    label: locale.value === 'sv' ? 'Klart eller hoppad' : 'Done or skipped',
    tasks: tasks.value.filter((task) => task.status === 'completed' || task.status === 'skipped')
  }
]);

const showMobileDetail = computed(() => !isWide.value && selectedTask.value !== null);

onMounted(() => {
  void refresh();
  handleScroll();
  window.addEventListener('scroll', handleScroll, { passive: true });
  splitQuery?.addEventListener('change', syncViewport);
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
  splitQuery?.removeEventListener('change', syncViewport);
});

watch(tasks, (nextTasks) => {
  syncChecklist(nextTasks);
  syncSelection(nextTasks);
}, { immediate: true });

function syncViewport() {
  isWide.value = splitQuery?.matches ?? false;
  syncSelection(tasks.value);
}

function handleScroll() {
  isCondensed.value = typeof window !== 'undefined' && window.scrollY > 28;
}

function preferredTask(nextTasks: Task[]) {
  return (
    nextTasks.find((task) => task.status === 'in_progress')
    ?? (summary.value?.recommended_task ? nextTasks.find((task) => task.id === summary.value?.recommended_task?.task.id) : null)
    ?? nextTasks[0]
    ?? null
  );
}

function syncChecklist(nextTasks: Task[]) {
  const nextState: Record<string, boolean[]> = {};

  for (const task of nextTasks) {
    const size = checklistForTask(task.type, locale.value).length;
    const current = checklistState.value[task.id] ?? [];
    nextState[task.id] = Array.from({ length: size }, (_, index) => current[index] ?? false);
  }

  checklistState.value = nextState;
}

function syncSelection(nextTasks: Task[]) {
  if (selectedTaskId.value && nextTasks.some((task) => task.id === selectedTaskId.value)) {
    return;
  }

  if (!isWide.value) {
    selectedTaskId.value = null;
    return;
  }

  selectedTaskId.value = preferredTask(nextTasks)?.id ?? null;
}

function toggleChecklist(index: number) {
  const currentTask = selectedTask.value;
  if (!currentTask) {
    return;
  }

  checklistState.value[currentTask.id] = checklistState.value[currentTask.id].map((value, currentIndex) => (
    currentIndex === index ? !value : value
  ));
}

function openTask(task: Task) {
  if (task.status === 'pending' && activeTask.value && activeTask.value.id !== task.id) {
    switchCandidate.value = task;
    return;
  }

  selectedTaskId.value = task.id;
}

async function startTask(task: Task) {
  if (activeTask.value && activeTask.value.id !== task.id) {
    switchCandidate.value = task;
    return;
  }

  await patchTask(task.id, { status: 'in_progress' });
  selectedTaskId.value = task.id;
}

async function completeTask(task: Task) {
  await patchTask(task.id, {
    status: 'completed',
    completed_by: workerHandle,
    resolution_code: resolutionForAction(task, 'complete')
  });
  selectedTaskId.value = task.id;
}

async function skipTask(task: Task) {
  await patchTask(task.id, {
    status: 'skipped',
    completed_by: workerHandle,
    resolution_code: resolutionForAction(task, 'skip')
  });
  selectedTaskId.value = task.id;
}

async function confirmSwitchTask() {
  const currentTask = activeTask.value;
  const nextTask = switchCandidate.value;
  if (!currentTask || !nextTask) {
    switchCandidate.value = null;
    return;
  }

  await patchTask(currentTask.id, { status: 'pending' });
  await patchTask(nextTask.id, { status: 'in_progress' });
  selectedTaskId.value = nextTask.id;
  switchCandidate.value = null;
}
</script>

<template>
  <main class="app-shell">
    <ProgressShell
      :completed="summary?.completed ?? 0"
      :condensed="isCondensed"
      :skipped="summary?.skipped ?? 0"
      :total="summary?.total ?? tasks.length"
    />

    <div class="app-content">
      <p
        v-if="error"
        class="error-banner"
      >
        {{ error }}
      </p>

      <p
        v-if="loading"
        class="loading-banner"
      >
        {{ locale === 'sv' ? 'Laddar arbetsflode...' : 'Loading shift flow...' }}
      </p>

      <section
        v-else
        class="app-layout"
        :class="{ 'is-wide': isWide }"
      >
        <TaskListPanel
          v-if="isWide || !showMobileDetail"
          :locale="locale"
          :recommended="summary?.recommended_task ?? null"
          :sections="taskSections"
          :selected-task-id="selectedTaskId"
          @select="openTask"
        />

        <TaskDetailPanel
          v-if="isWide || showMobileDetail"
          :action-busy="actionBusyId === selectedTask?.id"
          :checklist-state="selectedTask ? checklistState[selectedTask.id] ?? [] : []"
          :is-wide="isWide"
          :locale="locale"
          :task="selectedTask"
          @back="selectedTaskId = null"
          @complete="completeTask"
          @skip="skipTask"
          @start="startTask"
          @toggle-checklist="toggleChecklist"
        />
      </section>
    </div>

    <TaskSwitchDialog
      :current-task="activeTask"
      :locale="locale"
      :next-task="switchCandidate"
      :open="switchCandidate !== null"
      @cancel="switchCandidate = null"
      @confirm="confirmSwitchTask"
    />
  </main>
</template>
