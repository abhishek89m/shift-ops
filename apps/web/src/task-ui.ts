import type { LocaleCode, Task } from './types';

type ChecklistCopy = Record<LocaleCode, string[]>;

const checklistByType: Record<string, ChecklistCopy> = {
  parking_fix: {
    en: ['Clear the lane or walkway', 'Stand scooter inside the parking area', 'Confirm rider path is open'],
    sv: ['Frigor cykelbanan eller gangytan', 'Placera scootern inom parkeringsytan', 'Bekrafta att passagen ar fri']
  },
  battery_swap: {
    en: ['Unlock battery bay', 'Swap in charged battery', 'Wake vehicle and confirm power'],
    sv: ['Las upp batterifacket', 'Byt till laddat batteri', 'Vack fordonet och bekrafta strom']
  },
  quality_check: {
    en: ['Check brakes and bell', 'Inspect stem and deck', 'Confirm scooter is ride-ready'],
    sv: ['Kontrollera bromsar och ringklocka', 'Inspektera stam och platta', 'Bekrafta att scootern ar redo']
  },
  retrieve_damaged: {
    en: ['Verify vehicle ID', 'Mark scooter as unavailable', 'Load scooter for repair pickup'],
    sv: ['Bekrafta fordons-ID', 'Markera scootern som otillganglig', 'Lasta scootern for verkstad']
  }
};

const taskTypeLabels: Record<string, Record<LocaleCode, string>> = {
  parking_fix: { en: 'Parking fix', sv: 'Parkering' },
  battery_swap: { en: 'Battery swap', sv: 'Batteribyte' },
  quality_check: { en: 'Quality check', sv: 'Kvalitetskontroll' },
  retrieve_damaged: { en: 'Repair pickup', sv: 'Hamta for reparation' }
};

const statusLabels: Record<string, Record<LocaleCode, string>> = {
  pending: { en: 'Remaining', sv: 'Kvar' },
  in_progress: { en: 'In progress', sv: 'Pagar' },
  completed: { en: 'Done', sv: 'Klart' },
  skipped: { en: 'Skipped', sv: 'Hoppad' }
};

const taskIdPrefixes: Record<string, string> = {
  parking_fix: 'TP',
  battery_swap: 'TB',
  quality_check: 'TQ',
  retrieve_damaged: 'TR'
};

export function checklistForTask(taskType: string, locale: LocaleCode) {
  return checklistByType[taskType]?.[locale] ?? checklistByType.quality_check[locale];
}

export function labelForTaskType(taskType: string, locale: LocaleCode) {
  return taskTypeLabels[taskType]?.[locale] ?? taskType;
}

export function labelForStatus(status: string, locale: LocaleCode) {
  return statusLabels[status]?.[locale] ?? status;
}

export function displayTaskId(task: Task) {
  const prefix = taskIdPrefixes[task.type] ?? 'TS';
  const suffixMatch = task.id.match(/(\d+)$/);
  const numericSuffix = suffixMatch ? suffixMatch[1].padStart(3, '0') : '000';

  return `${prefix}-${numericSuffix}`;
}

export function resolutionForAction(task: Task, action: 'complete' | 'skip') {
  if (action === 'skip') {
    return task.blocked_access_severity > 0 ? 'blocked_access' : 'unable_to_locate';
  }

  switch (task.type) {
    case 'parking_fix':
      return 'reparked';
    case 'battery_swap':
      return 'battery_swapped';
    case 'retrieve_damaged':
      return 'collected_for_repair';
    case 'quality_check':
    default:
      return 'checked_ok';
  }
}

export function detailOutcomeCopy(task: Task, locale: LocaleCode) {
  const completeOutcome: Record<LocaleCode, string> = {
    en: `Complete logs "${resolutionForAction(task, 'complete')}".`,
    sv: `Klarmarkering loggar "${resolutionForAction(task, 'complete')}".`
  };
  const skipOutcome: Record<LocaleCode, string> = {
    en: `Skip logs "${resolutionForAction(task, 'skip')}".`,
    sv: `Hoppa over loggar "${resolutionForAction(task, 'skip')}".`
  };

  return {
    complete: completeOutcome[locale],
    skip: skipOutcome[locale]
  };
}
