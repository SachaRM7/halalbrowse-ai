package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"halalbrowse-ai/core/blocklist"
	"halalbrowse-ai/core/ml"
	"halalbrowse-ai/core/prayertimes"
)

func TestScoreResponseBlocksKnownHaramText(t *testing.T) {
	proxy := Proxy{
		Config:        Config{Threshold: 0.70},
		TextScorer:    ml.NewTextScorer(),
		Blocklist:     blocklist.Blocklist{},
		PrayerManager: prayertimes.Manager{StrictDelta: 0.15},
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	rec := httptest.NewRecorder()
	decision := proxy.scoreResponse(rec, req, []byte("online casino with explicit adult content"), false)
	if !decision.Blocked {
		t.Fatalf("expected response to be blocked")
	}
	if decision.Reason == "" {
		t.Fatalf("expected a blocking reason")
	}
}

func TestScoreResponseAllowsSafeText(t *testing.T) {
	proxy := Proxy{
		Config:        Config{Threshold: 0.70},
		TextScorer:    ml.NewTextScorer(),
		Blocklist:     blocklist.Blocklist{},
		PrayerManager: prayertimes.Manager{StrictDelta: 0.15},
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	rec := httptest.NewRecorder()
	decision := proxy.scoreResponse(rec, req, []byte("Learn tajweed and read Quran today"), false)
	if decision.Blocked {
		t.Fatalf("expected safe content to be allowed")
	}
}
