package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
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

func TestListTasksReturnsServerErrorWhenStoreFails(t *testing.T) {
	env := newTestEnv(t)
	env.store.close()

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	rec := httptest.NewRecorder()

	env.server.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestSummaryReturnsServerErrorWhenStoreFails(t *testing.T) {
	env := newTestEnv(t)
	env.store.close()

	req := httptest.NewRequest(http.MethodGet, "/v1/summary", nil)
	rec := httptest.NewRecorder()

	env.server.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestPatchTaskRejectsDirectTerminalUpdateFromPending(t *testing.T) {
	server := newTestServer(t)

	body := taskPatchRequest{
		Status:         ptrStatus(statusCompleted),
		CompletedBy:    ptr("abhishek"),
		ResolutionCode: ptr("battery_swapped"),
		Notes:          ptr("Swapped pack and rechecked lock."),
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_battery_01", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestPatchTaskRejectsInvalidTransition(t *testing.T) {
	server := newTestServer(t)

	startBody := taskPatchRequest{Status: ptrStatus(statusInProgress)}
	startPayload, _ := json.Marshal(startBody)

	startReq := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_parking_01", bytes.NewReader(startPayload))
	startRec := httptest.NewRecorder()
	server.ServeHTTP(startRec, startReq)

	if startRec.Code != http.StatusOK {
		t.Fatalf("expected first update to pass, got %d", startRec.Code)
	}

	body := taskPatchRequest{
		Status:         ptrStatus(statusCompleted),
		CompletedBy:    ptr("abhishek"),
		ResolutionCode: ptr("reparked"),
	}
	payload, _ := json.Marshal(body)

	firstReq := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_parking_01", bytes.NewReader(payload))
	firstRec := httptest.NewRecorder()
	server.ServeHTTP(firstRec, firstReq)

	if firstRec.Code != http.StatusOK {
		t.Fatalf("expected second update to pass, got %d", firstRec.Code)
	}

	secondBody := taskPatchRequest{Status: ptrStatus(statusInProgress)}
	secondPayload, _ := json.Marshal(secondBody)

	secondReq := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_parking_01", bytes.NewReader(secondPayload))
	secondRec := httptest.NewRecorder()
	server.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", secondRec.Code)
	}
}

func TestPatchTaskAllowsReturningActiveTaskToPending(t *testing.T) {
	server := newTestServer(t)

	body := taskPatchRequest{Status: ptrStatus(statusPending)}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_quality_01", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var updated task
	decodeJSON(t, rec.Body.Bytes(), &updated)

	if updated.Status != statusPending {
		t.Fatalf("expected pending, got %s", updated.Status)
	}

	if updated.StartedAt != nil || updated.CompletedAt != nil || updated.CompletedBy != nil || updated.ResolutionCode != nil {
		t.Fatalf("expected lifecycle fields cleared, got %#v", updated)
	}
}

func TestPatchTaskReturnsNotFoundForUnknownTask(t *testing.T) {
	server := newTestServer(t)

	body := taskPatchRequest{Status: ptrStatus(statusInProgress)}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_missing_01", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestPatchTaskRejectsUnknownResolutionCode(t *testing.T) {
	server := newTestServer(t)

	body := taskPatchRequest{
		Status:         ptrStatus(statusCompleted),
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

func TestPatchTaskRejectsMalformedJSON(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodPatch, "/v1/tasks/task_battery_01", strings.NewReader("{"))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestPatchTaskRejectsUnknownFields(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/v1/tasks/task_battery_01",
		strings.NewReader(`{"status":"in_progress","unexpected":true}`),
	)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestResetDataRestoresSeededTasks(t *testing.T) {
	server := newTestServer(t)

	completeBody := taskPatchRequest{
		Status:         ptrStatus(statusCompleted),
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

func TestSeedDataRestoresTasksWhenStoreIsEmpty(t *testing.T) {
	env := newTestEnv(t)

	if _, err := env.store.db.Exec(`DELETE FROM task_events;`); err != nil {
		t.Fatalf("delete task_events: %v", err)
	}
	if _, err := env.store.db.Exec(`DELETE FROM tasks;`); err != nil {
		t.Fatalf("delete tasks: %v", err)
	}

	seedReq := httptest.NewRequest(http.MethodPost, "/v1/dev/seed", nil)
	seedRec := httptest.NewRecorder()
	env.server.ServeHTTP(seedRec, seedReq)

	if seedRec.Code != http.StatusOK {
		t.Fatalf("expected seed 200, got %d, body=%s", seedRec.Code, seedRec.Body.String())
	}

	tasksReq := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	tasksRec := httptest.NewRecorder()
	env.server.ServeHTTP(tasksRec, tasksReq)

	var response tasksResponse
	decodeJSON(t, tasksRec.Body.Bytes(), &response)

	if len(response.Tasks) != 4 {
		t.Fatalf("expected 4 tasks after seed, got %d", len(response.Tasks))
	}
}

func TestPatchTaskPersistsChecklistState(t *testing.T) {
	server := newTestServer(t)

	body := taskPatchRequest{
		ChecklistState: &[]bool{true, false, true},
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

	if len(updated.ChecklistState) != 3 || !updated.ChecklistState[0] || updated.ChecklistState[1] || !updated.ChecklistState[2] {
		t.Fatalf("expected checklist state to persist, got %#v", updated.ChecklistState)
	}
}

func newTestServer(t *testing.T) http.Handler {
	return newTestEnv(t).server
}

type testEnv struct {
	server http.Handler
	store  *store
}

func newTestEnv(t *testing.T) testEnv {
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

	return testEnv{
		server: mux,
		store:  store,
	}
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

func ptrStatus(value taskStatus) *taskStatus {
	return &value
}
