<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import apiDoc from '../../docs/api.md?raw';
import architectureDoc from '../../docs/architecture.md?raw';
import guideDoc from '../../docs/guide.md?raw';
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
type StudioPage = 'app' | 'backend' | 'docs';
type DocTab = 'architecture' | 'guide' | 'api';

const previewMode = ref<PreviewMode>(resolveMode());
const page = ref<StudioPage>(resolvePage());
const docTab = ref<DocTab>(resolveDocTab());
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
const patchCompletedBy = ref('studio');
const patchNotes = ref('');

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://127.0.0.1:8080';
const previewSrc = import.meta.env.VITE_APP_BASE_URL ?? 'http://127.0.0.1:4174/';
const activePreview = computed(() => previewModes.find((item) => item.id === previewMode.value) ?? previewModes[0]);
const docsByTab: Record<DocTab, { title: string; source: string }> = {
  architecture: { title: 'Architecture', source: architectureDoc },
  guide: { title: 'Guide', source: guideDoc },
  api: { title: 'API docs', source: apiDoc }
};
const activeDoc = computed(() => docsByTab[docTab.value]);
const renderedDoc = computed(() => renderMarkdown(activeDoc.value.source));
const patchRequestBody = computed(() => {
  const body: Record<string, string> = { status: patchStatus.value };
  const isTerminal = patchStatus.value === 'completed' || patchStatus.value === 'skipped';

  if (isTerminal && patchCompletedBy.value.trim()) {
    body.completed_by = patchCompletedBy.value.trim();
  }

  if (isTerminal && patchResolutionCode.value.trim()) {
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

function resolvePage(): StudioPage {
  const hash = window.location.hash || '#/app';
  if (hash.startsWith('#/backend')) {
    return 'backend';
  }
  if (hash.startsWith('#/docs')) {
    return 'docs';
  }
  return 'app';
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
  docTab.value = resolveDocTab();
}

function goTo(nextPage: StudioPage) {
  if (nextPage === 'backend') {
    window.location.hash = '#/backend';
  } else if (nextPage === 'docs') {
    window.location.hash = `#/docs?tab=${docTab.value}`;
  } else {
    window.location.hash = `#/app?mode=${previewMode.value}`;
  }
  page.value = nextPage;
}

function setPreviewMode(nextMode: PreviewMode) {
  previewMode.value = nextMode;
  window.location.hash = `#/app?mode=${nextMode}`;
}

function resolveDocTab(): DocTab {
  const hash = window.location.hash || '#/docs?tab=architecture';
  const query = hash.split('?')[1] ?? '';
  const params = new URLSearchParams(query);
  const requestedTab = params.get('tab');

  return requestedTab === 'guide' || requestedTab === 'api' ? requestedTab : 'architecture';
}

function setDocTab(nextTab: DocTab) {
  docTab.value = nextTab;
  window.location.hash = `#/docs?tab=${nextTab}`;
}

function renderMarkdown(source: string) {
  const lines = source.split('\n');
  const html: string[] = [];
  let inList = false;
  let inCode = false;

  const escapeHtml = (value: string) => value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;');

  const renderInline = (value: string) => escapeHtml(value).replace(/`([^`]+)`/g, '<code>$1</code>');

  for (const rawLine of lines) {
    const line = rawLine.trimEnd();

    if (line.startsWith('```')) {
      if (inList) {
        html.push('</ul>');
        inList = false;
      }
      html.push(inCode ? '</code></pre>' : '<pre class="doc-code"><code>');
      inCode = !inCode;
      continue;
    }

    if (inCode) {
      html.push(`${escapeHtml(rawLine)}\n`);
      continue;
    }

    if (line === '') {
      if (inList) {
        html.push('</ul>');
        inList = false;
      }
      continue;
    }

    if (line.startsWith('## ')) {
      if (inList) {
        html.push('</ul>');
        inList = false;
      }
      html.push(`<h2>${renderInline(line.slice(3))}</h2>`);
      continue;
    }

    if (line.startsWith('# ')) {
      if (inList) {
        html.push('</ul>');
        inList = false;
      }
      html.push(`<h1>${renderInline(line.slice(2))}</h1>`);
      continue;
    }

    if (line.startsWith('- ')) {
      if (!inList) {
        html.push('<ul>');
        inList = true;
      }
      html.push(`<li>${renderInline(line.slice(2))}</li>`);
      continue;
    }

    if (inList) {
      html.push('</ul>');
      inList = false;
    }

    html.push(`<p>${renderInline(line)}</p>`);
  }

  if (inList) {
    html.push('</ul>');
  }

  if (inCode) {
    html.push('</code></pre>');
  }

  return html.join('');
}
</script>

<template>
  <main class="studio-shell">
    <header class="studio-topbar surface">
      <div>
        <span class="brand-mark">studio</span>
        <p class="muted-copy">Repo-local app + API testing shell for `ShiftOps`.</p>
      </div>
      <div class="topbar-actions">
        <div class="page-nav" role="tablist" aria-label="Studio pages">
          <button type="button" class="nav-pill" :class="{ 'is-active': page === 'app' }" @click="goTo('app')">
            App
          </button>
          <button type="button" class="nav-pill" :class="{ 'is-active': page === 'backend' }" @click="goTo('backend')">
            API
          </button>
          <button type="button" class="nav-pill" :class="{ 'is-active': page === 'docs' }" @click="goTo('docs')">
            Docs
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

    <section v-else-if="page === 'backend'" class="studio-grid">
      <article class="surface section-card">
        <div class="section-head">
          <div>
            <span class="eyebrow">API endpoints</span>
            <h2>Run the live backend like a small Swagger lab.</h2>
          </div>
          <div class="button-row">
            <button class="secondary-button" type="button" :disabled="actionBusy || loading" @click="refreshData">
              Refresh live data
            </button>
            <button class="secondary-button" type="button" :disabled="actionBusy" @click="runApi('Seed if empty', 'POST', '/v1/dev/seed')">
              Seed if empty
            </button>
            <button class="primary-button" type="button" :disabled="actionBusy" @click="runApi('Reset and seed', 'POST', '/v1/dev/reset')">
              Reset demo data
            </button>
          </div>
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
              <input v-model="patchCompletedBy" placeholder="studio" />
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

          <div class="schema-code-stack">
            <div>
              <span class="muted-label">tasks</span>
              <pre class="code-block compact"><code>tasks
  one row per field task
  id, type, status, title
  vehicle_id, location_label, distance_meters
  battery_level, urgency, blocked_access_severity
  checklist_state
  started_at, completed_at, completed_by, resolution_code</code></pre>
            </div>

            <div>
              <span class="muted-label">task_events</span>
              <pre class="code-block compact"><code>task_events
  append-only audit trail
  captures each task state change
  feeds events_today and later reporting
  stores notes and resolution history</code></pre>
            </div>

            <div>
              <span class="muted-label">status model</span>
              <pre class="code-block compact"><code>pending      queued work
in_progress  active field task
completed    terminal state
skipped      terminal state

ui rule
  switching tasks can move the live task back to pending</code></pre>
            </div>
          </div>
        </section>

        <details class="schema-details">
          <summary>SQLite schema reference</summary>
          <pre class="code-block compact"><code>{{ schemaText }}</code></pre>
        </details>
      </article>
    </section>

    <section v-else class="surface section-card docs-shell">
      <div class="section-head">
        <div>
          <span class="eyebrow">Docs</span>
          <h1>Architecture, guide, and API docs in one place.</h1>
        </div>
      </div>

      <div class="doc-tab-row" role="tablist" aria-label="Docs tabs">
        <button type="button" class="preview-tab" :class="{ 'is-active': docTab === 'architecture' }" @click="setDocTab('architecture')">
          Architecture
        </button>
        <button type="button" class="preview-tab" :class="{ 'is-active': docTab === 'guide' }" @click="setDocTab('guide')">
          Guide
        </button>
        <button type="button" class="preview-tab" :class="{ 'is-active': docTab === 'api' }" @click="setDocTab('api')">
          API docs
        </button>
      </div>

      <article class="doc-panel">
        <div class="section-head compact">
          <div>
            <span class="eyebrow">{{ activeDoc.title }}</span>
          </div>
        </div>
        <div class="doc-markdown" v-html="renderedDoc" />
      </article>
    </section>
  </main>
</template>
