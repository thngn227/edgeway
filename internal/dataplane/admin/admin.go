// Package admin serves the read-only admin API. It must never be exposed
// publicly without authentication (see docs/architecture/overview.md).
package admin

import (
	"encoding/json"
	"io"
	"net/http"

	"edgeway/internal/version"
)

func Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "OK")
	})

	mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"version": version.Version,
			"commit":  version.Commit,
		})
	})

	return mux
}
