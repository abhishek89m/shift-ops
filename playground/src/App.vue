<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import schemaText from '../../services/api/schema.sql?raw';

type TaskStatus = 'pending' | 'in_progress' | 'completed' | 'skipped';

interface Task {
  id: string;
  title: string;
  type: string;
  status: TaskStatus;
  urgency: number;
  location_label: string;
  distance_meters: number;
  battery_level?: number | null;
  notes: string;
  resolution_code?: string | null;
  completed_by?: string | null;
}

interface Recommendation {
  task: Task;
  reasons: string[];
}

interface Summary {
  pending: number;
  in_progress: number;
  completed: number;
  skipped: number;
  total: number;
  events_today: number;
  recommended_task?: Recommendation | null;
}

const previewModes = [
  { id: 'mobile', label: 'Mobile', size: '390 × 844' },
  { id: 'tablet', label: 'Tablet', size: '820 × 1180' },
  { id: 'desktop', label: 'Desktop', size: '1440 × 1024' }
] as const;

type PreviewMode = typeof previewModes[number]['id'];
type PlaygroundPage = 'app' | 'backend';

const previewMode = ref<PreviewMode>(resolveMode());
const page = ref<PlaygroundPage>(resolvePage());
const loading = ref(false);
const actionBusy = ref(false);
const error = ref('');
const summary = ref<Summary | null>(null);
const tasks = ref<Task[]>([]);
const iframeKey = ref(0);

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://127.0.0.1:8080';
const previewSrc = import.meta.env.VITE_APP_BASE_URL ?? 'http://127.0.0.1:4174/';
const activePreview = computed(() => previewModes.find((item) => item.id === previewMode.value) ?? previewModes[0]);

onMounted(() => {
  window.addEventListener('hashchange', syncPageFromHash);

  if (page.value === 'backend') {
    void refreshData();
  }
});

onUnmounted(() => {
  window.removeEventListener('hashchange', syncPageFromHash);
});

watch(page, async (nextPage) => {
  if (nextPage === 'backend' && summary.value === null) {
    await refreshData();
  }
});

async function refreshData() {
  loading.value = true;
  error.value = '';

  try {
    const [summaryResponse, tasksResponse] = await Promise.all([
      fetch(`${apiBaseUrl}/v1/summary`),
      fetch(`${apiBaseUrl}/v1/tasks`)
    ]);

    if (!summaryResponse.ok || !tasksResponse.ok) {
      throw new Error('Failed to load API data.');
    }

    summary.value = (await summaryResponse.json()) as Summary;
    tasks.value = ((await tasksResponse.json()) as { tasks: Task[] }).tasks;
  } catch (err) {
    const fallback = 'API unavailable. Start the stack with `pnpm podman:up` or run the local app/API commands, then reload this page.';
    error.value = err instanceof Error ? `${fallback} ${err.message}` : fallback;
  } finally {
    loading.value = false;
  }
}

async function postAction(path: '/api/v1/dev/reset' | '/api/v1/dev/seed') {
  actionBusy.value = true;
  error.value = '';

  try {
    const response = await fetch(`${apiBaseUrl}${path.replace('/api', '')}`, { method: 'POST' });
    if (!response.ok) {
      throw new Error('Dev action failed.');
    }

    await refreshData();
    reloadPreview();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown dev action error.';
  } finally {
    actionBusy.value = false;
  }
}

function reloadPreview() {
  iframeKey.value += 1;
}

function resolvePage(): PlaygroundPage {
  const hash = window.location.hash || '#/app';
  return hash.startsWith('#/backend') ? 'backend' : 'app';
}

function resolveMode(): PreviewMode {
  const hash = window.location.hash || '#/app';
  const modeValue = hash.split('?')[1] ?? '';
  const params = new URLSearchParams(modeValue);
  const requestedMode = params.get('mode');

  return previewModes.some((item) => item.id === requestedMode) ? (requestedMode as PreviewMode) : 'mobile';
}

function syncPageFromHash() {
  page.value = resolvePage();
  previewMode.value = resolveMode();
}

function goTo(nextPage: PlaygroundPage) {
  window.location.hash = nextPage === 'backend' ? '#/backend' : `#/app?mode=${previewMode.value}`;
  page.value = nextPage;
}

function setPreviewMode(nextMode: PreviewMode) {
  previewMode.value = nextMode;
  window.location.hash = `#/app?mode=${nextMode}`;
}
</script>

<template>
  <main class="playground-shell">
    <header class="playground-topbar surface">
      <div>
        <span class="brand-mark">playground</span>
        <p class="muted-copy">Repo-local app testing shell for `ShiftOps`.</p>
      </div>
      <div class="topbar-actions">
        <div class="page-nav" role="tablist" aria-label="Playground pages">
          <button type="button" class="nav-pill" :class="{ 'is-active': page === 'app' }" @click="goTo('app')">
            App
          </button>
          <button type="button" class="nav-pill" :class="{ 'is-active': page === 'backend' }" @click="goTo('backend')">
            Backend
          </button>
        </div>
      </div>
    </header>

    <section v-if="page === 'app'" class="surface section-card">
      <div class="section-head">
        <div>
          <span class="eyebrow">App</span>
          <h1>Inspect the real app in one testing tab.</h1>
        </div>
        <button class="secondary-button" type="button" @click="reloadPreview">Reload iframe</button>
      </div>

      <div class="preview-tab-row" role="tablist" aria-label="Preview sizes">
        <button
          v-for="mode in previewModes"
          :key="mode.id"
          type="button"
          class="preview-tab"
          :class="{ 'is-active': previewMode === mode.id }"
          @click="setPreviewMode(mode.id)"
        >
          <strong>{{ mode.label }}</strong>
          <span>{{ mode.size }}</span>
        </button>
      </div>

      <article class="device-frame" :class="`${previewMode}-frame`">
        <div class="device-topline">
          <span class="eyebrow subtle">Shift Ops app</span>
          <strong>{{ activePreview.size }}</strong>
        </div>
        <iframe
          :key="iframeKey"
          class="preview-viewport"
          :class="`${previewMode}-viewport`"
          :title="`${activePreview.label} preview`"
          :src="previewSrc"
        />
      </article>
    </section>

    <section v-else class="playground-grid">
      <article class="surface section-card">
        <div class="section-head">
          <div>
            <span class="eyebrow">Live data</span>
            <h2>Summary + task snapshot</h2>
          </div>
          <div class="button-row">
            <button class="secondary-button" type="button" :disabled="actionBusy" @click="refreshData">Refresh data</button>
            <button class="secondary-button" type="button" :disabled="actionBusy" @click="postAction('/api/v1/dev/seed')">Seed if empty</button>
            <button class="primary-button" type="button" :disabled="actionBusy" @click="postAction('/api/v1/dev/reset')">Reset + seed</button>
          </div>
        </div>

        <p v-if="error" class="error-copy">{{ error }}</p>
        <p v-else-if="loading" class="muted-copy">Loading API snapshot...</p>

        <template v-else>
          <div class="metric-grid">
            <div class="metric-card">
              <span class="muted-label">Pending</span>
              <strong>{{ summary?.pending ?? 0 }}</strong>
            </div>
            <div class="metric-card">
              <span class="muted-label">In progress</span>
              <strong>{{ summary?.in_progress ?? 0 }}</strong>
            </div>
            <div class="metric-card">
              <span class="muted-label">Done</span>
              <strong>{{ (summary?.completed ?? 0) + (summary?.skipped ?? 0) }}</strong>
            </div>
            <div class="metric-card">
              <span class="muted-label">Events today</span>
              <strong>{{ summary?.events_today ?? 0 }}</strong>
            </div>
          </div>

          <div v-if="summary?.recommended_task" class="recommended-card">
            <span class="muted-label">Recommended</span>
            <strong>{{ summary.recommended_task.task.title }}</strong>
            <p class="muted-copy">{{ summary.recommended_task.task.location_label }}</p>
            <div class="chip-row">
              <span v-for="reason in summary.recommended_task.reasons" :key="reason" class="chip">{{ reason }}</span>
            </div>
          </div>

          <div class="task-list">
            <article v-for="task in tasks" :key="task.id" class="task-card">
              <div class="task-head">
                <strong>{{ task.title }}</strong>
                <span class="status-pill">{{ task.status }}</span>
              </div>
              <div class="task-meta">
                <span>{{ task.type }}</span>
                <span>{{ task.location_label }}</span>
                <span>{{ task.distance_meters }}m</span>
                <span v-if="task.battery_level !== null && task.battery_level !== undefined">{{ task.battery_level }}%</span>
              </div>
            </article>
          </div>
        </template>
      </article>

      <article class="surface section-card">
        <div class="section-head">
          <div>
            <span class="eyebrow">Backend schema</span>
            <h2>Current SQLite shape</h2>
          </div>
        </div>
        <pre class="code-block"><code>{{ schemaText }}</code></pre>
      </article>
    </section>
  </main>
</template>
