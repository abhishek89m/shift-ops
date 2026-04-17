package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

func (s *apiServer) handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.store.listTasks()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to load tasks"})
		return
	}

	writeJSON(w, http.StatusOK, tasksResponse{Tasks: tasks})
}

func (s *apiServer) handleSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.store.getSummary()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to load summary"})
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (s *apiServer) handlePatchTask(w http.ResponseWriter, r *http.Request) {
	var req taskPatchRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	updatedTask, err := s.store.updateTask(r.PathValue("id"), req)
	if err == nil {
		writeJSON(w, http.StatusOK, updatedTask)
		return
	}

	switch {
	case errors.Is(err, errTaskNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, errInvalidTransition), errors.Is(err, errValidation):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to update task"})
	}
}

func (s *apiServer) handleResetData(w http.ResponseWriter, r *http.Request) {
	if err := s.store.resetData(); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to reset data"})
		return
	}

	writeJSON(w, http.StatusOK, devActionResponse{OK: true, Action: "reset"})
}

func (s *apiServer) handleSeedData(w http.ResponseWriter, r *http.Request) {
	if err := s.store.seedData(); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to seed data"})
		return
	}

	writeJSON(w, http.StatusOK, devActionResponse{OK: true, Action: "seed"})
}
