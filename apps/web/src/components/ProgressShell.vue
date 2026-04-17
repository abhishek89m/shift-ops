<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import LanguageSelector from './LanguageSelector.vue';

const props = defineProps<{
  completed: number;
  total: number;
}>();

const { t } = useI18n();
const progress = props.total === 0 ? 0 : Math.round((props.completed / props.total) * 100);
</script>

<template>
  <main class="app-shell">
    <div class="app-toolbar">
      <LanguageSelector />
    </div>

    <section class="product-shell">
      <div class="product-shell-head">
        <div class="product-copy">
          <span class="eyebrow">{{ t('shell.eyebrow') }}</span>
          <h1>{{ t('shell.title') }}</h1>
          <p class="lede">
            {{ t('shell.lede') }}
          </p>
        </div>
        <div class="progress-summary">
          <span class="meta-label">{{ t('shell.progressLabel') }}</span>
          <strong>{{ completed }}/{{ total }}</strong>
          <span>{{ t('shell.progressDone', { progress }) }}</span>
        </div>
      </div>

      <div
        class="progress-track"
        aria-hidden="true"
      >
        <span
          class="progress-fill"
          :style="{ width: `${progress}%` }"
        />
      </div>
    </section>

    <section class="preview-card">
      <span class="meta-label">{{ t('shell.directionLabel') }}</span>
      <h2>{{ t('shell.directionTitle') }}</h2>
      <p>{{ t('shell.directionBody') }}</p>
    </section>
  </main>
</template>
