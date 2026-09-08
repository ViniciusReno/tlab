package web_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/web"
)

func TestPersistentEmptyStateAndSourceIsolation(t *testing.T) {
	s, err := app.OpenPersistent(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h, err := web.New(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/", "/playground"} {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 200 || !strings.Contains(r.Body.String(), "No official quotes yet") || !strings.Contains(r.Body.String(), "tesouro-lab demo") {
			t.Fatalf("empty page: %d %s", r.Code, r.Body.String())
		}
		for _, forbidden := range []string{"733.86", "Historical demo", "scenario-form", "<script"} {
			if strings.Contains(r.Body.String(), forbidden) {
				t.Fatalf("demo content in persistent page: %s", forbidden)
			}
		}
	}
	for _, tc := range []struct{ query, code string }{
		{"bond=" + app.DemoBond + "&yield=8.88&source=demo", "source_mismatch"},
		{"bond=" + app.DemoBond + "&yield=8.88", "source_unavailable"},
	} {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest("GET", "/api/scenario?"+tc.query, nil))
		if r.Code != 422 || !strings.Contains(r.Body.String(), tc.code) {
			t.Fatalf("API: %d %s", r.Code, r.Body.String())
		}
	}
	if has, err := s.HasQuotes(context.Background()); err != nil || has {
		t.Fatal("HTTP requests seeded persistent data")
	}
}
