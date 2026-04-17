export type LocaleCode = 'en' | 'sv';

export type TaskStatus = 'pending' | 'in_progress' | 'completed' | 'skipped';

export interface Task {
  id: string;
  vehicle_id: string;
  title: string;
  type: string;
  status: TaskStatus;
  urgency: number;
  location_label: string;
  lat: number;
  lng: number;
  distance_meters: number;
  blocked_access_severity: number;
  battery_level?: number | null;
  notes: string;
  created_at: string;
  updated_at: string;
  started_at?: string | null;
  completed_at?: string | null;
  completed_by?: string | null;
  resolution_code?: string | null;
  checklist_state: boolean[];
}

export interface Recommendation {
  task: Task;
  reasons: string[];
}

export interface Summary {
  pending: number;
  in_progress: number;
  completed: number;
  skipped: number;
  total: number;
  events_today: number;
  recommended_task?: Recommendation | null;
}

export interface TasksResponse {
  tasks: Task[];
}

export interface TaskPatchRequest {
  status?: TaskStatus;
  completed_by?: string;
  resolution_code?: string;
  notes?: string;
  checklist_state?: boolean[];
}
