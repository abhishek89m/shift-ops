import { createI18n } from 'vue-i18n';

export const localeStorageKey = 'shift-ops.locale';

const messages = {
  en: {
    shell: {
      progressLabel: 'Shift',
      tasksLabel: 'Tasks',
      doneLabel: 'Done',
      progressDone: '{progress}% done',
      activeLabel: 'Active',
      remainingLabel: 'Remaining',
      eventsLabel: 'Events today'
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
      progressLabel: 'Pass',
      tasksLabel: 'Uppgifter',
      doneLabel: 'Klart',
      progressDone: '{progress}% klart',
      activeLabel: 'Pagar',
      remainingLabel: 'Kvar',
      eventsLabel: 'Handelser idag'
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
