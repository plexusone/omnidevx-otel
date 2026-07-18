// Command omnidevx-otel is the OmniDevX local telemetry plane: an
// OTLP/HTTP (JSON) ingest endpoint for AI-agent metrics plus a read-only
// query API over the canonical event store, consumed by Grafana.
//
// See docs/specs/ in this repository.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	core "github.com/plexusone/omnidevx-core"
	"github.com/plexusone/omnidevx-core/store"
	"github.com/plexusone/omnidevx-otel/server"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:4318", "listen address (keep on loopback)")
	person := flag.String("person", "", "subject person ID stamped on ingested events")
	dir := flag.String("store", "", "event store root (default ~/.plexusone/omnidevx/data)")
	flag.Parse()

	s, err := store.Open(store.Options{Dir: *dir})
	if err != nil {
		log.Fatal(err)
	}
	handler, err := server.New(server.Options{
		Store:   s,
		Subject: core.SubjectRef{PersonID: *person},
	})
	if err != nil {
		log.Fatal(err)
	}

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("omnidevx-otel listening on %s, store %s", *addr, s.Dir())
	log.Fatal(httpServer.ListenAndServe())
}
