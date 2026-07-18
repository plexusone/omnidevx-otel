package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	core "github.com/plexusone/omnidevx-core"
	"github.com/plexusone/omnidevx-core/store"
)

func event(id string, t time.Time, eventType core.EventType, mode core.CollectionMode, attrs map[string]any) core.Event {
	return core.Event{
		ID:         id,
		Type:       eventType,
		Timestamp:  t,
		Source:     core.Source{Provider: "test", Product: "claude-code"},
		Attributes: attrs,
		Provenance: core.Provenance{CollectionMode: mode, Confidence: 1},
	}
}

// fixtureHandler builds a server over a temp store seeded with events
// spanning a week boundary (Sunday 2026-07-12 / Monday 2026-07-13).
func fixtureHandler(t *testing.T) http.Handler {
	t.Helper()
	s, err := store.Open(store.Options{Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	sun := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC) // Sunday
	mon := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC) // Monday
	events := []core.Event{
		event("m1", sun, core.EventMessageCompleted, core.ModeHistory,
			map[string]any{core.AttrInputTokens: 100, core.AttrOutputTokens: 10}),
		event("m2", mon, core.EventMessageCompleted, core.ModeHistory,
			map[string]any{core.AttrInputTokens: 200, core.AttrOutputTokens: 20}),
		event("u1", mon, core.EventUsageRecorded, core.ModeOTel,
			map[string]any{core.AttrInputTokens: 50, core.AttrCostUSD: 0.25}),
		event("s1", mon, core.EventSessionStarted, core.ModeHistory, nil),
		event("p1", mon, core.EventPromptSubmitted, core.ModeHistory, nil),
		event("p2", mon, core.EventPromptSubmitted, core.ModeHistory, nil),
		event("c1", sun, core.EventChangeCommitted, core.ModeHistory,
			map[string]any{core.AttrAIAssisted: true}),
		event("c2", sun, core.EventChangeCommitted, core.ModeHistory,
			map[string]any{core.AttrAIAssisted: false}),
	}
	if _, err := s.Write(context.Background(), events); err != nil {
		t.Fatal(err)
	}
	h, err := New(Options{Store: s, Logf: func(string, ...any) {}})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

type seriesResponse struct {
	Metric string           `json:"metric"`
	Bucket string           `json:"bucket"`
	Rows   []map[string]any `json:"rows"`
}

func getSeries(t *testing.T, h http.Handler, query string) (*seriesResponse, int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet,
		"/api/series?since=2026-07-01T00:00:00Z&until=2026-08-01T00:00:00Z&"+query, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		return nil, w.Code
	}
	var resp seriesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v: %s", err, w.Body.String())
	}
	return &resp, w.Code
}

func rowByT(rows []map[string]any, t string) map[string]any {
	for _, r := range rows {
		if r["t"] == t {
			return r
		}
	}
	return nil
}

func TestTokensDaily(t *testing.T) {
	h := fixtureHandler(t)
	resp, code := getSeries(t, h, "metric=tokens&bucket=day")
	if code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	sun := rowByT(resp.Rows, "2026-07-12")
	mon := rowByT(resp.Rows, "2026-07-13")
	if sun == nil || sun["input"] != float64(100) {
		t.Errorf("sunday: %+v", sun)
	}
	// Monday combines message.completed (200) and usage.recorded (50).
	if mon == nil || mon["input"] != float64(250) || mon["output"] != float64(20) {
		t.Errorf("monday: %+v", mon)
	}
}

func TestModeFilter(t *testing.T) {
	h := fixtureHandler(t)
	resp, _ := getSeries(t, h, "metric=tokens&bucket=day&mode=otel")
	mon := rowByT(resp.Rows, "2026-07-13")
	if mon == nil || mon["input"] != float64(50) {
		t.Errorf("otel-only monday: %+v", mon)
	}
	if sun := rowByT(resp.Rows, "2026-07-12"); sun != nil {
		t.Errorf("sunday had no otel events: %+v", sun)
	}
}

func TestWeekBucketBoundary(t *testing.T) {
	h := fixtureHandler(t)
	resp, _ := getSeries(t, h, "metric=tokens&bucket=week")
	// Sunday 07-12 belongs to the week starting Monday 07-06; Monday 07-13
	// starts its own week. They must land in different buckets.
	if len(resp.Rows) != 2 {
		t.Fatalf("weeks: got %d rows: %+v", len(resp.Rows), resp.Rows)
	}
	if resp.Rows[0]["t"] != "2026-07-06" || resp.Rows[1]["t"] != "2026-07-13" {
		t.Errorf("week starts: %v, %v", resp.Rows[0]["t"], resp.Rows[1]["t"])
	}
}

func TestCommitsAIShare(t *testing.T) {
	h := fixtureHandler(t)
	resp, _ := getSeries(t, h, "metric=commits&bucket=day")
	sun := rowByT(resp.Rows, "2026-07-12")
	if sun == nil || sun["commits"] != float64(2) || sun["ai_share"] != 0.5 {
		t.Errorf("commits sunday: %+v", sun)
	}
}

func TestActivityAndCost(t *testing.T) {
	h := fixtureHandler(t)
	resp, _ := getSeries(t, h, "metric=activity&bucket=day")
	mon := rowByT(resp.Rows, "2026-07-13")
	if mon == nil || mon["sessions"] != float64(1) || mon["prompts"] != float64(2) {
		t.Errorf("activity monday: %+v", mon)
	}
	resp, _ = getSeries(t, h, "metric=cost&bucket=day")
	if mon := rowByT(resp.Rows, "2026-07-13"); mon == nil || mon["cost_usd"] != 0.25 {
		t.Errorf("cost monday: %+v", mon)
	}
}

func TestUnknownMetricAndHealth(t *testing.T) {
	h := fixtureHandler(t)
	if _, code := getSeries(t, h, "metric=nope"); code != http.StatusBadRequest {
		t.Errorf("unknown metric: got %d, want 400", code)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health: got %d", w.Code)
	}
}
