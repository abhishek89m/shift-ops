CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  vehicle_id TEXT NOT NULL,
  title TEXT NOT NULL,
  type TEXT NOT NULL,
  status TEXT NOT NULL,
  urgency INTEGER NOT NULL,
  location_label TEXT NOT NULL,
  lat REAL NOT NULL,
  lng REAL NOT NULL,
  distance_meters INTEGER NOT NULL DEFAULT 0,
  blocked_access_severity INTEGER NOT NULL DEFAULT 0,
  battery_level INTEGER,
  notes TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  started_at TEXT,
  completed_at TEXT,
  completed_by TEXT,
  resolution_code TEXT,
  checklist_state TEXT NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS task_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  from_status TEXT,
  to_status TEXT,
  actor TEXT,
  resolution_code TEXT,
  notes TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
