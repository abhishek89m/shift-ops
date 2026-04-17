package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

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
			resolution_code,
			checklist_state
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

	now := time.Now().UTC()
	updated, statusChanged, err := applyTaskPatch(current, req, now)
	if err != nil {
		return task{}, err
	}

	if err := persistTaskUpdate(ctx, tx, id, updated); err != nil {
		return task{}, err
	}

	if statusChanged {
		if err := insertStatusChangedEvent(ctx, tx, current, updated, now); err != nil {
			return task{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return task{}, err
	}

	return updated, nil
}

func applyTaskPatch(current task, req taskPatchRequest, now time.Time) (task, bool, error) {
	updated := current
	updated.UpdatedAt = now

	statusChanged := req.Status != nil && *req.Status != current.Status
	if statusChanged && !canTransition(current.Status, *req.Status) {
		return task{}, false, fmt.Errorf("%w: %s -> %s", errInvalidTransition, current.Status, *req.Status)
	}

	if req.Notes != nil {
		updated.Notes = strings.TrimSpace(*req.Notes)
	}

	if req.ChecklistState != nil {
		updated.ChecklistState = slices.Clone(*req.ChecklistState)
	}

	if req.Status != nil {
		updated.Status = *req.Status
	}

	applyTaskLifecycle(&updated, req, now, statusChanged)
	return updated, statusChanged, nil
}

func applyTaskLifecycle(updated *task, req taskPatchRequest, now time.Time, statusChanged bool) {
	if !statusChanged {
		return
	}

	switch updated.Status {
	case statusInProgress:
		if updated.StartedAt == nil {
			startedAt := now
			updated.StartedAt = &startedAt
		}
	case statusPending:
		// Returning work to the queue should clear the previous run lifecycle fields.
		updated.StartedAt = nil
		updated.CompletedAt = nil
		updated.CompletedBy = nil
		updated.ResolutionCode = nil
	case statusCompleted, statusSkipped:
		if updated.StartedAt == nil {
			startedAt := now
			updated.StartedAt = &startedAt
		}

		completedAt := now
		completedBy := strings.TrimSpace(*req.CompletedBy)
		resolutionCode := strings.TrimSpace(*req.ResolutionCode)

		updated.CompletedAt = &completedAt
		updated.CompletedBy = &completedBy
		updated.ResolutionCode = &resolutionCode
	}
}

func persistTaskUpdate(ctx context.Context, tx *sql.Tx, id string, updated task) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE tasks
		SET
			status = ?,
			notes = ?,
			updated_at = ?,
			started_at = ?,
			completed_at = ?,
			completed_by = ?,
			resolution_code = ?,
			checklist_state = ?
		WHERE id = ?;
	`,
		updated.Status,
		updated.Notes,
		formatTime(updated.UpdatedAt),
		formatNullableTime(updated.StartedAt),
		formatNullableTime(updated.CompletedAt),
		nullableString(updated.CompletedBy),
		nullableString(updated.ResolutionCode),
		mustMarshalChecklist(updated.ChecklistState),
		id,
	)
	return err
}

func insertStatusChangedEvent(ctx context.Context, tx *sql.Tx, current, updated task, now time.Time) error {
	_, err := tx.ExecContext(ctx, `
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
		current.ID,
		current.Status,
		updated.Status,
		nullableString(updated.CompletedBy),
		nullableString(updated.ResolutionCode),
		updated.Notes,
		formatTime(now),
	)
	return err
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
			resolution_code,
			checklist_state
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
		checklistState  string
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
		&checklistState,
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
	if checklistState != "" {
		if err := json.Unmarshal([]byte(checklistState), &record.ChecklistState); err != nil {
			return task{}, err
		}
	}

	return record, nil
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

func mustMarshalChecklist(value []bool) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}

	return string(payload)
}
