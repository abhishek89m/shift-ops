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

func main() {
	addr := envOrDefault("API_ADDR", ":8080")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			OK:      true,
			Service: "shift-ops-api",
			Version: buildinfo.Version,
		})
	})

	log.Printf("shift-ops api listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
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
