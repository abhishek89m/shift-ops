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

interface ApiResult {
  label: string;
  method: string;
  path: string;
  status: number | null;
  requestBody: string | null;
  responseBody: string;
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
const latestApiResult = ref<ApiResult | null>(null);

const patchTaskId = ref('');
const patchStatus = ref<TaskStatus>('in_progress');
const patchResolutionCode = ref('');
const patchCompletedBy = ref('playground');
const patchNotes = ref('');

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://127.0.0.1:8080';
const previewSrc = import.meta.env.VITE_APP_BASE_URL ?? 'http://127.0.0.1:4174/';
const activePreview = computed(() => previewModes.find((item) => item.id === previewMode.value) ?? previewModes[0]);
const patchRequestBody = computed(() => {
  const body: Record<string, string> = { status: patchStatus.value };

  if (patchCompletedBy.value.trim()) {
    body.completed_by = patchCompletedBy.value.trim();
  }

  if (patchResolutionCode.value.trim()) {
    body.resolution_code = patchResolutionCode.value.trim();
  }

  if (patchNotes.value.trim()) {
    body.notes = patchNotes.value.trim();
  }

  return body;
});

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

watch(tasks, (nextTasks) => {
  if (!patchTaskId.value && nextTasks.length > 0) {
    patchTaskId.value = nextTasks[0].id;
  }
}, { immediate: true });

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
    const fallback = 'API unavailable. Start the stack with `pnpm podman:up`, then reload this page.';
    error.value = err instanceof Error ? `${fallback} ${err.message}` : fallback;
  } finally {
    loading.value = false;
  }
}

async function runApi(label: string, method: 'GET' | 'POST' | 'PATCH', path: string, body?: Record<string, string>) {
  actionBusy.value = true;
  error.value = '';

  try {
    const response = await fetch(`${apiBaseUrl}${path}`, {
      method,
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined
    });

    const raw = await response.text();
    let pretty = raw;

    try {
      pretty = JSON.stringify(JSON.parse(raw), null, 2);
    } catch {
      // Keep text as-is when not JSON.
    }

    latestApiResult.value = {
      label,
      method,
      path,
      status: response.status,
      requestBody: body ? JSON.stringify(body, null, 2) : null,
      responseBody: pretty
    };

    if (!response.ok) {
      throw new Error(`${method} ${path} failed with ${response.status}.`);
    }

    if (path === '/v1/tasks') {
      tasks.value = ((JSON.parse(raw)) as { tasks: Task[] }).tasks;
    } else if (path === '/v1/summary') {
      summary.value = JSON.parse(raw) as Summary;
    } else {
      await refreshData();
      reloadPreview();
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown API error.';
  } finally {
    actionBusy.value = false;
  }
}

async function runPatchTask() {
  if (!patchTaskId.value) {
    error.value = 'Choose a task before sending a PATCH request.';
    return;
  }

  await runApi(
    'PATCH task',
    'PATCH',
    `/v1/tasks/${patchTaskId.value}`,
    patchRequestBody.value
  );
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
        <p class="muted-copy">Repo-local app + API testing shell for `ShiftOps`.</p>
      </div>
      <div class="topbar-actions">
        <div class="page-nav" role="tablist" aria-label="Playground pages">
          <button type="button" class="nav-pill" :class="{ 'is-active': page === 'app' }" @click="goTo('app')">
            App
          </button>
          <button type="button" class="nav-pill" :class="{ 'is-active': page === 'backend' }" @click="goTo('backend')">
            API
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
            <span class="eyebrow">API endpoints</span>
            <h2>Run the live backend like a small Swagger lab.</h2>
          </div>
          <button class="secondary-button" type="button" :disabled="actionBusy || loading" @click="refreshData">
            Refresh snapshot
          </button>
        </div>

        <p v-if="error" class="error-copy">{{ error }}</p>

        <div class="endpoint-grid">
          <article class="endpoint-card">
            <div class="endpoint-head">
              <span class="method-pill method-get">GET</span>
              <code>/healthz</code>
            </div>
            <p class="muted-copy">Service health, name, version.</p>
            <button class="secondary-button" type="button" :disabled="actionBusy" @click="runApi('Health', 'GET', '/healthz')">
              Run
            </button>
          </article>

          <article class="endpoint-card">
            <div class="endpoint-head">
              <span class="method-pill method-get">GET</span>
              <code>/v1/summary</code>
            </div>
            <p class="muted-copy">Progress counts and recommended task.</p>
            <button class="secondary-button" type="button" :disabled="actionBusy" @click="runApi('Summary', 'GET', '/v1/summary')">
              Run
            </button>
          </article>

          <article class="endpoint-card">
            <div class="endpoint-head">
              <span class="method-pill method-get">GET</span>
              <code>/v1/tasks</code>
            </div>
            <p class="muted-copy">Full task list with current statuses.</p>
            <button class="secondary-button" type="button" :disabled="actionBusy" @click="runApi('Tasks', 'GET', '/v1/tasks')">
              Run
            </button>
          </article>

          <article class="endpoint-card">
            <div class="endpoint-head">
              <span class="method-pill method-post">POST</span>
              <code>/v1/dev/seed</code>
            </div>
            <p class="muted-copy">Seed data if empty.</p>
            <button class="secondary-button" type="button" :disabled="actionBusy" @click="runApi('Seed if empty', 'POST', '/v1/dev/seed')">
              Run
            </button>
          </article>

          <article class="endpoint-card">
            <div class="endpoint-head">
              <span class="method-pill method-post">POST</span>
              <code>/v1/dev/reset</code>
            </div>
            <p class="muted-copy">Reset and reseed the SQLite file.</p>
            <button class="primary-button" type="button" :disabled="actionBusy" @click="runApi('Reset and seed', 'POST', '/v1/dev/reset')">
              Run
            </button>
          </article>
        </div>

        <article class="patch-console">
          <div class="section-head compact">
            <div>
              <span class="eyebrow">PATCH tester</span>
              <h2>Try a task update</h2>
            </div>
          </div>

          <div class="form-grid">
            <label class="field">
              <span class="muted-label">Task</span>
              <select v-model="patchTaskId">
                <option v-for="task in tasks" :key="task.id" :value="task.id">
                  {{ task.id }} · {{ task.title }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="muted-label">Status</span>
              <select v-model="patchStatus">
                <option value="pending">pending</option>
                <option value="in_progress">in_progress</option>
                <option value="completed">completed</option>
                <option value="skipped">skipped</option>
              </select>
            </label>

            <label class="field">
              <span class="muted-label">Resolution</span>
              <input v-model="patchResolutionCode" placeholder="checked_ok / reparked / ..." />
            </label>

            <label class="field">
              <span class="muted-label">Completed by</span>
              <input v-model="patchCompletedBy" placeholder="playground" />
            </label>
          </div>

          <label class="field">
            <span class="muted-label">Notes</span>
            <textarea v-model="patchNotes" rows="3" placeholder="Optional operator note" />
          </label>

          <div class="request-preview">
            <span class="muted-label">Request body preview</span>
            <pre class="code-block compact"><code>{{ JSON.stringify(patchRequestBody, null, 2) }}</code></pre>
          </div>

          <div class="button-row">
            <button class="primary-button" type="button" :disabled="actionBusy" @click="runPatchTask">
              Send PATCH
            </button>
          </div>
        </article>
      </article>

      <article class="surface section-card api-side-column">
        <div class="section-head compact">
          <div>
            <span class="eyebrow">Latest result</span>
            <h2>Response viewer</h2>
          </div>
        </div>

        <div v-if="latestApiResult" class="response-meta">
          <div class="endpoint-head">
            <span class="method-pill" :class="`method-${latestApiResult.method.toLowerCase()}`">{{ latestApiResult.method }}</span>
            <code>{{ latestApiResult.path }}</code>
          </div>
          <p class="muted-copy">{{ latestApiResult.label }} · status {{ latestApiResult.status ?? 'n/a' }}</p>

          <div v-if="latestApiResult.requestBody" class="request-preview">
            <span class="muted-label">Request body</span>
            <pre class="code-block compact"><code>{{ latestApiResult.requestBody }}</code></pre>
          </div>

          <div class="request-preview">
            <span class="muted-label">Response body</span>
            <pre class="code-block compact"><code>{{ latestApiResult.responseBody }}</code></pre>
          </div>
        </div>
        <p v-else class="muted-copy">Run any endpoint to inspect the live JSON here.</p>

        <section class="schema-overview">
          <div class="section-head compact">
            <div>
              <span class="eyebrow">Schema overview</span>
              <h2>SQLite model at a glance</h2>
            </div>
          </div>

          <div class="schema-grid">
            <article class="schema-card">
              <strong>tasks</strong>
              <ul class="schema-list">
                <li>one row per field task</li>
                <li><code>id</code>, <code>type</code>, <code>status</code>, <code>title</code></li>
                <li><code>vehicle_id</code>, <code>location_label</code>, <code>distance_meters</code></li>
                <li><code>battery_level</code>, <code>urgency</code>, <code>blocked_access_severity</code></li>
                <li><code>started_at</code>, <code>completed_at</code>, <code>completed_by</code>, <code>resolution_code</code></li>
              </ul>
            </article>

            <article class="schema-card">
              <strong>task_events</strong>
              <ul class="schema-list">
                <li>append-only audit trail</li>
                <li>captures each task state change</li>
                <li>feeds <code>events_today</code> and later reporting</li>
                <li>stores notes and resolution history</li>
              </ul>
            </article>
          </div>

          <article class="schema-card">
            <strong>status model</strong>
            <ul class="schema-list">
              <li><code>pending</code> = queued work</li>
              <li><code>in_progress</code> = active field task</li>
              <li><code>completed</code> / <code>skipped</code> = terminal states</li>
              <li>UI can move the active task back to <code>pending</code> when switching tasks</li>
            </ul>
          </article>
        </section>

        <details class="schema-details">
          <summary>SQLite schema reference</summary>
          <pre class="code-block compact"><code>{{ schemaText }}</code></pre>
        </details>
      </article>
    </section>
  </main>
</template>
