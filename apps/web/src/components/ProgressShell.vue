<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import LanguageSelector from './LanguageSelector.vue';

const props = defineProps<{
  condensed: boolean;
  total: number;
  completed: number;
  skipped: number;
}>();

const { locale, t } = useI18n();
const progress = props.total === 0 ? 0 : Math.round(((props.completed + props.skipped) / props.total) * 100);
const titleDate = computed(() => new Intl.DateTimeFormat(locale.value, {
  weekday: 'short',
  day: 'numeric',
  month: 'short'
}).format(new Date()));
</script>

<template>
  <section
    class="product-shell"
    :class="{ 'is-condensed': condensed }"
  >
    <div class="product-shell-head">
      <div class="product-shell-title">
        <h1>{{ t('shell.tasksLabel') }} {{ titleDate }}</h1>
      </div>
      <LanguageSelector />
    </div>

    <div
      class="progress-track"
      :aria-label="t('shell.progressDone', { progress })"
      role="progressbar"
      :aria-valuemin="0"
      :aria-valuemax="100"
      :aria-valuenow="progress"
    >
      <span
        class="progress-fill"
        :style="{ width: `${progress}%` }"
      />
    </div>
  </section>
</template>
