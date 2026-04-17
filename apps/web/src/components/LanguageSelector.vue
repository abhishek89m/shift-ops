<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { localeStorageKey } from '../i18n';

const isOpen = ref(false);
const { locale, t } = useI18n();
const root = ref<HTMLElement | null>(null);

watch(locale, (value) => {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(localeStorageKey, value);
});

function setLocale(nextLocale: 'en' | 'sv') {
  locale.value = nextLocale;
  isOpen.value = false;
}

function handleDocumentClick(event: MouseEvent) {
  if (!root.value?.contains(event.target as Node)) {
    isOpen.value = false;
  }
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick);
});

onUnmounted(() => {
  document.removeEventListener('click', handleDocumentClick);
});
</script>

<template>
  <div
    ref="root"
    class="settings-menu"
  >
    <button
      type="button"
      class="settings-trigger"
      :aria-expanded="isOpen"
      :aria-label="t('settings.label')"
      @click="isOpen = !isOpen"
    >
      <svg
        class="settings-trigger-icon"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path
          d="M4 7h10M18 7h2M14 7a2 2 0 1 1 4 0a2 2 0 0 1-4 0ZM4 12h2M10 12h10M6 12a2 2 0 1 1 4 0a2 2 0 0 1-4 0ZM4 17h10M18 17h2M14 17a2 2 0 1 1 4 0a2 2 0 0 1-4 0Z"
        />
      </svg>
    </button>

    <div
      v-if="isOpen"
      class="settings-popup"
    >
      <div class="settings-row">
        <div class="settings-copy">
          <span class="meta-label">{{ t('settings.label') }}</span>
          <strong>{{ t('settings.language') }}</strong>
        </div>

        <div class="settings-options">
          <button
            type="button"
            class="settings-option"
            :class="{ 'is-active': locale === 'en' }"
            @click="setLocale('en')"
          >
            {{ t('settings.english') }}
          </button>
          <button
            type="button"
            class="settings-option"
            :class="{ 'is-active': locale === 'sv' }"
            @click="setLocale('sv')"
          >
            {{ t('settings.swedish') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
