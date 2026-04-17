package main

import "time"

type taskStatus string

const (
	statusPending    taskStatus = "pending"
	statusInProgress taskStatus = "in_progress"
	statusCompleted  taskStatus = "completed"
	statusSkipped    taskStatus = "skipped"
)

type task struct {
	ID                    string     `json:"id"`
	VehicleID             string     `json:"vehicle_id"`
	Title                 string     `json:"title"`
	Type                  string     `json:"type"`
	Status                taskStatus `json:"status"`
	Urgency               int        `json:"urgency"`
	LocationLabel         string     `json:"location_label"`
	Lat                   float64    `json:"lat"`
	Lng                   float64    `json:"lng"`
	DistanceMeters        int        `json:"distance_meters"`
	BlockedAccessSeverity int        `json:"blocked_access_severity"`
	BatteryLevel          *int       `json:"battery_level,omitempty"`
	Notes                 string     `json:"notes"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	StartedAt             *time.Time `json:"started_at,omitempty"`
	CompletedAt           *time.Time `json:"completed_at,omitempty"`
	CompletedBy           *string    `json:"completed_by,omitempty"`
	ResolutionCode        *string    `json:"resolution_code,omitempty"`
}

type recommendation struct {
	Task    task     `json:"task"`
	Reasons []string `json:"reasons"`
}

type summary struct {
	Pending         int             `json:"pending"`
	InProgress      int             `json:"in_progress"`
	Completed       int             `json:"completed"`
	Skipped         int             `json:"skipped"`
	Total           int             `json:"total"`
	EventsToday     int             `json:"events_today"`
	RecommendedTask *recommendation `json:"recommended_task,omitempty"`
}

type tasksResponse struct {
	Tasks []task `json:"tasks"`
}

type taskPatchRequest struct {
	Status         taskStatus `json:"status"`
	CompletedBy    *string    `json:"completed_by,omitempty"`
	ResolutionCode *string    `json:"resolution_code,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type devActionResponse struct {
	OK     bool   `json:"ok"`
	Action string `json:"action"`
}
