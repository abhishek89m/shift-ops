import { createI18n } from 'vue-i18n';

export const localeStorageKey = 'shift-ops.locale';

const messages = {
  en: {
    shell: {
      eyebrow: 'Shift Ops',
      title: 'Keep the next field task obvious.',
      lede: 'Vue frontend scaffold for a mobile-first field operations flow. This first commit keeps the product shell light and leaves room for the real task flow.',
      progressLabel: 'Progress',
      progressDone: '{progress}% done',
      directionLabel: 'Current direction',
      directionTitle: 'Top progress shell',
      directionBody: 'The recommended-next card, task list, task detail, and preview harness will layer on top of this shell in later commits.'
    },
    settings: {
      label: 'Settings',
      language: 'Language',
      english: 'English',
      swedish: 'Swedish'
    }
  },
  sv: {
    shell: {
      eyebrow: 'Shift Ops',
      title: 'Gor nasta faltuppgift tydlig.',
      lede: 'Vue-frontendens grundskal for ett mobilforst faltflode. Forsta commiten haller produktskalet latt och lamnar plats for det riktiga uppgiftsflodet.',
      progressLabel: 'Framsteg',
      progressDone: '{progress}% klart',
      directionLabel: 'Nuvarande riktning',
      directionTitle: 'Ovra framstegsskalet',
      directionBody: 'Kortet for rekommenderad nasta uppgift, uppgiftslistan, detaljvyn och forhandsgranskningen laggs ovanpa detta skal i kommande commits.'
    },
    settings: {
      label: 'Installningar',
      language: 'Sprak',
      english: 'Engelska',
      swedish: 'Svenska'
    }
  }
} as const;

function detectLocale() {
  if (typeof window === 'undefined') {
    return 'en';
  }

  const stored = window.localStorage.getItem(localeStorageKey);
  if (stored === 'en' || stored === 'sv') {
    return stored;
  }

  const browserLocale = window.navigator.language.toLowerCase();
  return browserLocale.startsWith('sv') ? 'sv' : 'en';
}

export const i18n = createI18n({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages
});
