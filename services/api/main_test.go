package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestListTasksReturnsSeededTasks(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response tasksResponse
	decodeJSON(t, rec.Body.Bytes(), &response)

	if len(response.Tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(response.Tasks))
	}
}

func TestSummaryReturnsRecommendedTask(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/summary", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response summary
	decodeJSON(t, rec.Body.Bytes(), &response)

	if response.RecommendedTask == nil {
		t.Fatal("expected recommended task")
	}

	if response.RecommendedTask.Task.ID != "task_quality_01" {
		t.Fatalf("expected task_quality_01, got %s", response.RecommendedTask.Task.ID)
	}
}

func TestPatchTaskUpdatesTask(t *testing.T) {
	server := newTestServer(t)

	body := taskPatchRequest{
		Status:         statusCompleted,
		CompletedBy:    ptr("abhishek"),
		ResolutionCode: ptr("battery_swapped"),
		Notes:          ptr("Swapped pack and rechecked lock."),
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_battery_01", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var updated task
	decodeJSON(t, rec.Body.Bytes(), &updated)

	if updated.Status != statusCompleted {
		t.Fatalf("expected completed, got %s", updated.Status)
	}

	if updated.CompletedBy == nil || *updated.CompletedBy != "abhishek" {
		t.Fatalf("expected completed_by=abhishek, got %#v", updated.CompletedBy)
	}

	if updated.ResolutionCode == nil || *updated.ResolutionCode != "battery_swapped" {
		t.Fatalf("expected resolution_code=battery_swapped, got %#v", updated.ResolutionCode)
	}

	if updated.StartedAt == nil {
		t.Fatal("expected started_at to be stamped on direct terminal update")
	}
}

func TestPatchTaskRejectsInvalidTransition(t *testing.T) {
	server := newTestServer(t)

	body := taskPatchRequest{
		Status:         statusCompleted,
		CompletedBy:    ptr("abhishek"),
		ResolutionCode: ptr("reparked"),
	}
	payload, _ := json.Marshal(body)

	firstReq := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_parking_01", bytes.NewReader(payload))
	firstRec := httptest.NewRecorder()
	server.ServeHTTP(firstRec, firstReq)

	if firstRec.Code != http.StatusOK {
		t.Fatalf("expected first update to pass, got %d", firstRec.Code)
	}

	secondBody := taskPatchRequest{Status: statusInProgress}
	secondPayload, _ := json.Marshal(secondBody)

	secondReq := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_parking_01", bytes.NewReader(secondPayload))
	secondRec := httptest.NewRecorder()
	server.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", secondRec.Code)
	}
}

func TestPatchTaskRejectsUnknownResolutionCode(t *testing.T) {
	server := newTestServer(t)

	body := taskPatchRequest{
		Status:         statusCompleted,
		CompletedBy:    ptr("abhishek"),
		ResolutionCode: ptr("made_it_up"),
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_battery_01", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestResetDataRestoresSeededTasks(t *testing.T) {
	server := newTestServer(t)

	completeBody := taskPatchRequest{
		Status:         statusCompleted,
		CompletedBy:    ptr("abhishek"),
		ResolutionCode: ptr("battery_swapped"),
	}
	payload, _ := json.Marshal(completeBody)

	completeReq := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_battery_01", bytes.NewReader(payload))
	completeRec := httptest.NewRecorder()
	server.ServeHTTP(completeRec, completeReq)

	resetReq := httptest.NewRequest(http.MethodPost, "/v1/dev/reset", nil)
	resetRec := httptest.NewRecorder()
	server.ServeHTTP(resetRec, resetReq)

	if resetRec.Code != http.StatusOK {
		t.Fatalf("expected reset 200, got %d", resetRec.Code)
	}

	tasksReq := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	tasksRec := httptest.NewRecorder()
	server.ServeHTTP(tasksRec, tasksReq)

	var response tasksResponse
	decodeJSON(t, tasksRec.Body.Bytes(), &response)

	if len(response.Tasks) != 4 {
		t.Fatalf("expected 4 tasks after reset, got %d", len(response.Tasks))
	}
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "shift-ops.db")
	store, err := newStore(dbPath)
	if err != nil {
		t.Fatalf("newStore: %v", err)
	}
	t.Cleanup(store.close)

	server := &apiServer{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tasks", server.handleListTasks)
	mux.HandleFunc("GET /v1/summary", server.handleSummary)
	mux.HandleFunc("PATCH /v1/tasks/{id}", server.handlePatchTask)
	mux.HandleFunc("POST /v1/dev/reset", server.handleResetData)
	mux.HandleFunc("POST /v1/dev/seed", server.handleSeedData)

	return mux
}

func decodeJSON(t *testing.T, payload []byte, target any) {
	t.Helper()

	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func ptr(value string) *string {
	return &value
}
