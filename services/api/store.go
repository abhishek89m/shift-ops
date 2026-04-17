package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql seed.sql
var migrations embed.FS

var (
	errTaskNotFound      = errors.New("task not found")
	errInvalidTransition = errors.New("invalid task status transition")
	errValidation        = errors.New("invalid task update")
)

var allowedResolutionCodes = []string{
	"battery_swapped",
	"reparked",
	"checked_ok",
	"collected_for_repair",
	"unable_to_locate",
	"blocked_access",
	"duplicate_task",
	"other",
}

type store struct {
	db *sql.DB
}

func newStore(dbPath string) (*store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := initSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &store{db: db}, nil
}

func (s *store) close() {
	_ = s.db.Close()
}

func initSchema(db *sql.DB) error {
	schemaSQL, err := migrations.ReadFile("schema.sql")
	if err != nil {
		return err
	}

	seedSQL, err := migrations.ReadFile("seed.sql")
	if err != nil {
		return err
	}

	if _, err := db.Exec(string(schemaSQL)); err != nil {
		return err
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tasks;`).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err = db.Exec(string(seedSQL))
	return err
}

func (s *store) listTasks() ([]task, error) {
	rows, err := s.db.Query(`
		SELECT
			id,
			vehicle_id,
			title,
			type,
			status,
			urgency,
			location_label,
			lat,
			lng,
			distance_meters,
			blocked_access_severity,
			battery_level,
			notes,
			created_at,
			updated_at,
			started_at,
			completed_at,
			completed_by,
			resolution_code
		FROM tasks
		ORDER BY
			CASE status
				WHEN 'in_progress' THEN 0
				WHEN 'pending' THEN 1
				WHEN 'completed' THEN 2
				WHEN 'skipped' THEN 3
				ELSE 4
			END,
			urgency DESC,
			distance_meters ASC,
			created_at ASC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]task, 0, 8)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (s *store) seedData() error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM tasks;`).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	return runSeedSQL(s.db)
}

func (s *store) resetData() error {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(`DELETE FROM task_events;`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM tasks;`); err != nil {
		return err
	}
	if err := runSeedSQL(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *store) getSummary() (summary, error) {
	tasks, err := s.listTasks()
	if err != nil {
		return summary{}, err
	}

	var result summary
	result.Total = len(tasks)

	for _, task := range tasks {
		switch task.Status {
		case statusPending:
			result.Pending++
		case statusInProgress:
			result.InProgress++
		case statusCompleted:
			result.Completed++
		case statusSkipped:
			result.Skipped++
		}
	}

	if err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM task_events
		WHERE date(created_at) = date('now');
	`).Scan(&result.EventsToday); err != nil {
		return summary{}, err
	}

	result.RecommendedTask = chooseRecommendation(tasks)
	return result, nil
}

func (s *store) updateTask(id string, req taskPatchRequest) (task, error) {
	if err := validatePatchRequest(req); err != nil {
		return task{}, err
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return task{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	current, err := getTaskByID(ctx, tx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task{}, errTaskNotFound
		}
		return task{}, err
	}

	if !canTransition(current.Status, req.Status) {
		return task{}, fmt.Errorf("%w: %s -> %s", errInvalidTransition, current.Status, req.Status)
	}

	now := time.Now().UTC()
	updated := current
	updated.Status = req.Status
	updated.UpdatedAt = now

	if req.Notes != nil {
		updated.Notes = strings.TrimSpace(*req.Notes)
	}

	if req.Status == statusInProgress && updated.StartedAt == nil {
		startedAt := now
		updated.StartedAt = &startedAt
	}

	if req.Status == statusCompleted || req.Status == statusSkipped {
		// Direct terminal updates should still leave an accountability trail.
		if updated.StartedAt == nil {
			startedAt := now
			updated.StartedAt = &startedAt
		}

		completedAt := now
		updated.CompletedAt = &completedAt

		completedBy := strings.TrimSpace(*req.CompletedBy)
		resolutionCode := strings.TrimSpace(*req.ResolutionCode)
		updated.CompletedBy = &completedBy
		updated.ResolutionCode = &resolutionCode
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE tasks
		SET
			status = ?,
			notes = ?,
			updated_at = ?,
			started_at = ?,
			completed_at = ?,
			completed_by = ?,
			resolution_code = ?
		WHERE id = ?;
	`,
		updated.Status,
		updated.Notes,
		formatTime(updated.UpdatedAt),
		formatNullableTime(updated.StartedAt),
		formatNullableTime(updated.CompletedAt),
		nullableString(updated.CompletedBy),
		nullableString(updated.ResolutionCode),
		id,
	); err != nil {
		return task{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO task_events (
			task_id,
			event_type,
			from_status,
			to_status,
			actor,
			resolution_code,
			notes,
			created_at
		)
		VALUES (?, 'status_changed', ?, ?, ?, ?, ?, ?);
	`,
		id,
		current.Status,
		updated.Status,
		nullableString(updated.CompletedBy),
		nullableString(updated.ResolutionCode),
		updated.Notes,
		formatTime(now),
	); err != nil {
		return task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task{}, err
	}

	return updated, nil
}

func getTaskByID(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, id string) (task, error) {
	row := queryer.QueryRowContext(ctx, `
		SELECT
			id,
			vehicle_id,
			title,
			type,
			status,
			urgency,
			location_label,
			lat,
			lng,
			distance_meters,
			blocked_access_severity,
			battery_level,
			notes,
			created_at,
			updated_at,
			started_at,
			completed_at,
			completed_by,
			resolution_code
		FROM tasks
		WHERE id = ?;
	`, id)

	return scanTask(row)
}

func scanTask(scanner interface{ Scan(...any) error }) (task, error) {
	var (
		record          task
		batteryLevel    sql.NullInt64
		createdAtText   string
		updatedAtText   string
		startedAtText   sql.NullString
		completedAtText sql.NullString
		completedBy     sql.NullString
		resolutionCode  sql.NullString
	)

	if err := scanner.Scan(
		&record.ID,
		&record.VehicleID,
		&record.Title,
		&record.Type,
		&record.Status,
		&record.Urgency,
		&record.LocationLabel,
		&record.Lat,
		&record.Lng,
		&record.DistanceMeters,
		&record.BlockedAccessSeverity,
		&batteryLevel,
		&record.Notes,
		&createdAtText,
		&updatedAtText,
		&startedAtText,
		&completedAtText,
		&completedBy,
		&resolutionCode,
	); err != nil {
		return task{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtText)
	if err != nil {
		return task{}, err
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtText)
	if err != nil {
		return task{}, err
	}

	record.CreatedAt = createdAt
	record.UpdatedAt = updatedAt

	if batteryLevel.Valid {
		value := int(batteryLevel.Int64)
		record.BatteryLevel = &value
	}
	if startedAtText.Valid {
		value, err := time.Parse(time.RFC3339, startedAtText.String)
		if err != nil {
			return task{}, err
		}
		record.StartedAt = &value
	}
	if completedAtText.Valid {
		value, err := time.Parse(time.RFC3339, completedAtText.String)
		if err != nil {
			return task{}, err
		}
		record.CompletedAt = &value
	}
	if completedBy.Valid {
		value := completedBy.String
		record.CompletedBy = &value
	}
	if resolutionCode.Valid {
		value := resolutionCode.String
		record.ResolutionCode = &value
	}

	return record, nil
}

func validatePatchRequest(req taskPatchRequest) error {
	if req.Status != statusPending && req.Status != statusInProgress && req.Status != statusCompleted && req.Status != statusSkipped {
		return fmt.Errorf("%w: status must be pending, in_progress, completed, or skipped", errValidation)
	}

	if req.Status == statusCompleted || req.Status == statusSkipped {
		if req.CompletedBy == nil || strings.TrimSpace(*req.CompletedBy) == "" {
			return fmt.Errorf("%w: completed_by is required for terminal updates", errValidation)
		}
		if req.ResolutionCode == nil || strings.TrimSpace(*req.ResolutionCode) == "" {
			return fmt.Errorf("%w: resolution_code is required for terminal updates", errValidation)
		}
		// Keep reporting/metrics clean by constraining terminal outcomes to known codes.
		if !slices.Contains(allowedResolutionCodes, strings.TrimSpace(*req.ResolutionCode)) {
			return fmt.Errorf("%w: resolution_code must be one of %s", errValidation, strings.Join(allowedResolutionCodes, ", "))
		}
		return nil
	}

	if req.CompletedBy != nil || req.ResolutionCode != nil {
		return fmt.Errorf("%w: completed_by and resolution_code are only valid for terminal updates", errValidation)
	}

	return nil
}

func canTransition(from, to taskStatus) bool {
	allowed := map[taskStatus][]taskStatus{
		statusPending:    {statusInProgress, statusCompleted, statusSkipped},
		statusInProgress: {statusPending, statusCompleted, statusSkipped},
	}

	return slices.Contains(allowed[from], to)
}

func chooseRecommendation(tasks []task) *recommendation {
	bestScore := -1 << 30
	var bestTask task
	var bestReasons []string
	found := false

	for _, task := range tasks {
		if task.Status != statusPending && task.Status != statusInProgress {
			continue
		}

		score, reasons := scoreTask(task)
		if !found || score > bestScore {
			bestScore = score
			bestTask = task
			bestReasons = reasons
			found = true
		}
	}

	if !found {
		return nil
	}

	return &recommendation{
		Task:    bestTask,
		Reasons: bestReasons,
	}
}

func scoreTask(task task) (int, []string) {
	score := task.Urgency * 100
	reasons := make([]string, 0, 4)

	if task.Status == statusInProgress {
		score += 220
		reasons = append(reasons, "Continue active task")
	}

	if task.Urgency >= 3 {
		reasons = append(reasons, "High urgency")
	}

	if task.BlockedAccessSeverity > 0 {
		score += task.BlockedAccessSeverity * 35
		reasons = append(reasons, "Blocked access risk")
	}

	if task.Type == "battery_swap" && task.BatteryLevel != nil && *task.BatteryLevel <= 20 {
		score += 70
		reasons = append(reasons, "Low battery")
	}

	distancePenalty := min(task.DistanceMeters, 1200) / 12
	score -= distancePenalty
	if task.DistanceMeters <= 250 {
		reasons = append(reasons, "Close by")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "Next best available task")
	}

	return score, reasons
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

func formatNullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func runSeedSQL(execable interface {
	Exec(string, ...any) (sql.Result, error)
}) error {
	seedSQL, err := migrations.ReadFile("seed.sql")
	if err != nil {
		return err
	}

	_, err = execable.Exec(string(seedSQL))
	return err
}
