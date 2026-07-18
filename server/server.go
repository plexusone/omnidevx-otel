// Package server composes the OmniDevX local telemetry plane: OTLP/JSON
// ingest (via omnidevx-core/providers/genericotel) plus a read-only query
// API that aggregates the canonical event store for dashboards.
package server

import (
	"encoding/json"
	"log"
	"net/http"

	core "github.com/plexusone/omnidevx-core"
	"github.com/plexusone/omnidevx-core/providers/genericotel"
	"github.com/plexusone/omnidevx-core/store"
)

// Options configures the server.
type Options struct {
	Store   *store.Store
	Subject core.SubjectRef
	Logf    func(format string, args ...any)
}

// New returns the combined HTTP handler: /v1/* (ingest) and /api/* (query).
func New(opts Options) (http.Handler, error) {
	logf := opts.Logf
	if logf == nil {
		logf = log.Printf
	}
	receiver, err := genericotel.New(genericotel.Options{
		Store:   opts.Store,
		Subject: opts.Subject,
		OnWrite: func(result *store.WriteResult, service string) {
			if result.Written > 0 {
				logf("%s: wrote %d events (%d duplicates)", service, result.Written, result.Duplicates)
			}
		},
	})
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/", receiver.Handler())
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"status": "ok", "store": opts.Store.Dir()})
	})
	mux.Handle("GET /api/series", &seriesHandler{store: opts.Store})
	return mux, nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Headers are already sent; nothing to do but log via the default
		// logger — the client sees a truncated body.
		log.Printf("server: encode response: %v", err)
	}
}
