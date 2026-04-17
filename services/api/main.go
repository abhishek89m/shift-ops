package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/abhishek89m/shift-ops/services/api/internal/buildinfo"
)

type healthResponse struct {
	OK      bool   `json:"ok"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type apiServer struct {
	store *store
}

func main() {
	addr := envOrDefault("API_ADDR", ":8080")
	dbPath := envOrDefault("API_DB_PATH", "./shift-ops.db")

	store, err := newStore(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer store.close()

	server := &apiServer{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			OK:      true,
			Service: "shift-ops-api",
			Version: buildinfo.Version,
		})
	})
	mux.HandleFunc("GET /v1/tasks", server.handleListTasks)
	mux.HandleFunc("GET /v1/summary", server.handleSummary)
	mux.HandleFunc("PATCH /v1/tasks/{id}", server.handlePatchTask)
	mux.HandleFunc("POST /v1/dev/reset", server.handleResetData)
	mux.HandleFunc("POST /v1/dev/seed", server.handleSeedData)

	log.Printf("shift-ops api listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
