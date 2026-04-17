import { computed, ref } from 'vue';
import type { Summary, Task, TaskPatchRequest, TasksResponse } from '../types';

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://127.0.0.1:8080';

export function useShiftOps() {
  const tasks = ref<Task[]>([]);
  const summary = ref<Summary | null>(null);
  const loading = ref(true);
  const error = ref('');
  const actionBusyId = ref<string | null>(null);

  const activeTask = computed(() => tasks.value.find((task) => task.status === 'in_progress') ?? null);

  async function refresh(options: { silent?: boolean } = {}) {
    const shouldShowLoading = !options.silent || tasks.value.length === 0;
    if (shouldShowLoading) {
      loading.value = true;
    }

    error.value = '';

    try {
      const [summaryResponse, tasksResponse] = await Promise.all([
        fetch(`${apiBaseUrl}/v1/summary`),
        fetch(`${apiBaseUrl}/v1/tasks`)
      ]);

      if (!summaryResponse.ok || !tasksResponse.ok) {
        throw new Error('Failed to load shift data.');
      }

      summary.value = (await summaryResponse.json()) as Summary;
      tasks.value = ((await tasksResponse.json()) as TasksResponse).tasks;
    } catch (reason) {
      error.value = reason instanceof Error ? reason.message : 'Unknown shift data error.';
    } finally {
      if (shouldShowLoading) {
        loading.value = false;
      }
    }
  }

  async function patchTask(id: string, payload: TaskPatchRequest) {
    actionBusyId.value = id;
    error.value = '';

    try {
      const response = await fetch(`${apiBaseUrl}/v1/tasks/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        const message = (await response.json().catch(() => null)) as { error?: string } | null;
        throw new Error(message?.error ?? 'Task update failed.');
      }

      await refresh({ silent: true });
    } catch (reason) {
      error.value = reason instanceof Error ? reason.message : 'Unknown task update error.';
      throw reason;
    } finally {
      actionBusyId.value = null;
    }
  }

  return {
    activeTask,
    actionBusyId,
    error,
    loading,
    refresh,
    summary,
    tasks,
    patchTask
  };
}
