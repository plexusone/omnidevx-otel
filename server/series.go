package server

import (
	"net/http"
	"sort"
	"time"

	core "github.com/plexusone/omnidevx-core"
	"github.com/plexusone/omnidevx-core/store"
)

// seriesHandler serves GET /api/series — time-bucketed aggregations over
// the canonical store, computed on demand (the store stays canon; nothing
// is cached).
type seriesHandler struct {
	store *store.Store
}

type row = map[string]any

func (h *seriesHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	metric := q.Get("metric")
	if metric == "" {
		metric = "tokens"
	}
	bucket := q.Get("bucket")
	if bucket != "day" && bucket != "week" {
		bucket = "day"
	}
	now := time.Now().UTC()
	since := now.AddDate(0, 0, -30)
	if s := q.Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			since = t
		}
	}
	until := now
	if s := q.Get("until"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			until = t
		}
	}
	modeFilter := q.Get("mode")

	read, err := h.store.Read(req.Context(), store.Query{
		Period: core.Period{Start: since, End: until},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buckets := map[string]row{}
	get := func(t time.Time) row {
		key := bucketKey(t, bucket)
		r := buckets[key]
		if r == nil {
			r = row{"t": key}
			buckets[key] = r
		}
		return r
	}
	addInt := func(r row, field string, delta int64) {
		if v, ok := r[field].(int64); ok {
			r[field] = v + delta
		} else {
			r[field] = delta
		}
	}
	addFloat := func(r row, field string, delta float64) {
		if v, ok := r[field].(float64); ok {
			r[field] = v + delta
		} else {
			r[field] = delta
		}
	}

	for _, e := range read.Events {
		if modeFilter != "" && string(e.Provenance.CollectionMode) != modeFilter {
			continue
		}
		switch metric {
		case "tokens":
			if e.Type != core.EventMessageCompleted && e.Type != core.EventUsageRecorded {
				continue
			}
			r := get(e.Timestamp)
			addInt(r, "input", intAttr(e, core.AttrInputTokens))
			addInt(r, "output", intAttr(e, core.AttrOutputTokens))
			addInt(r, "cache_read", intAttr(e, core.AttrCacheReadTokens))
			addInt(r, "cache_creation", intAttr(e, core.AttrCacheCreationTokens))
		case "cost":
			if e.Type != core.EventUsageRecorded {
				continue
			}
			if cost, ok := e.Attributes[core.AttrCostUSD].(float64); ok {
				addFloat(get(e.Timestamp), "cost_usd", cost)
			}
		case "activity":
			switch e.Type {
			case core.EventSessionStarted:
				addInt(get(e.Timestamp), "sessions", 1)
			case core.EventPromptSubmitted:
				addInt(get(e.Timestamp), "prompts", 1)
			case core.EventToolCompleted:
				addInt(get(e.Timestamp), "tool_calls", 1)
			}
		case "commits":
			if e.Type != core.EventChangeCommitted {
				continue
			}
			r := get(e.Timestamp)
			addInt(r, "commits", 1)
			if assisted, _ := e.Attributes[core.AttrAIAssisted].(bool); assisted {
				addInt(r, "ai_assisted", 1)
			}
		default:
			http.Error(w, "unknown metric: "+metric, http.StatusBadRequest)
			return
		}
	}

	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := make([]row, 0, len(keys))
	for _, k := range keys {
		r := buckets[k]
		if metric == "commits" {
			commits, _ := r["commits"].(int64)
			ai, _ := r["ai_assisted"].(int64)
			if commits > 0 {
				r["ai_share"] = float64(ai) / float64(commits)
			}
		}
		rows = append(rows, r)
	}
	writeJSON(w, map[string]any{"metric": metric, "bucket": bucket, "rows": rows})
}

// bucketKey truncates a timestamp to its day, or to the Monday of its ISO
// week.
func bucketKey(t time.Time, bucket string) string {
	t = t.UTC()
	if bucket == "week" {
		weekday := (int(t.Weekday()) + 6) % 7 // Monday = 0
		t = t.AddDate(0, 0, -weekday)
	}
	return t.Format("2006-01-02")
}

// intAttr reads a numeric attribute; JSON round-trips deliver float64.
func intAttr(e core.Event, key string) int64 {
	switch v := e.Attributes[key].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	}
	return 0
}
