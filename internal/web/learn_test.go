package web_test

import (
	"context"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/web"
)

func TestLearnLocalJourneysAndSourceIsolation(t *testing.T) {
	for _, demo := range []bool{true, false} {
		name := "persistent"
		if demo {
			name = "demo"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			var s *app.Service
			var err error
			if demo {
				s, err = app.OpenDemo(ctx)
			} else {
				s, err = app.OpenPersistent(ctx, t.TempDir())
			}
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			h, err := web.New(s)
			if err != nil {
				t.Fatal(err)
			}
			get := func(path string) string {
				t.Helper()
				w := httptest.NewRecorder()
				h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
				if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "</html>") {
					t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
				}
				if w.Header().Get("Content-Security-Policy") == "" {
					t.Fatalf("missing security policy: %s", path)
				}
				return w.Body.String()
			}
			body := get("/learn")
			if strings.Contains(body, "<script") {
				t.Fatal("Learn should operate without JavaScript")
			}
			// Follow all local learning links, including scenario exercises, without
			// a browser script or live network. Check anchors against their pages.
			exercises := 0
			for _, match := range regexp.MustCompile(`<a\b[^>]*href="([^"]+)"`).FindAllStringSubmatch(body, -1) {
				link := html.UnescapeString(match[1])
				u, err := url.Parse(link)
				if err != nil {
					t.Fatal(err)
				}
				if u.IsAbs() {
					continue
				}
				target := body
				if u.Path != "" {
					target = get(u.RequestURI())
				}
				if u.Fragment != "" && !strings.Contains(target, `id="`+u.Fragment+`"`) {
					t.Fatalf("missing learning anchor: %s", link)
				}
				if u.Path == "/playground" && u.RawQuery != "" {
					exercises++
					v := u.Query()
					if !demo || v.Get("source") != "demo" || v.Get("date") == "" || v.Get("basis") == "" || v.Get("yield") == "" {
						t.Fatalf("exercise lacks explicit demo inputs: %s", link)
					}
					if !strings.Contains(target, "Historical demo") || !strings.Contains(target, "Hypothetical scenario price") || !strings.Contains(target, v.Get("date")) {
						t.Fatalf("exercise did not render historical calculation: %s", link)
					}
				}
			}
			if demo && exercises < 6 {
				t.Fatalf("missing instrument, shock, or base-PU exercises: %d", exercises)
			}
			paths := []string{"/", "/playground", "/portfolio", "/learn"}
			if !demo {
				paths = append(paths, "/market")
			}
			for _, path := range paths {
				page := get(path)
				for _, want := range []string{`href="/learn"`, `href="#main"`, `id="main"`, `aria-label="Main navigation"`} {
					if !strings.Contains(page, want) {
						t.Fatalf("%s missing navigation %s", path, want)
					}
				}
			}
			if !demo {
				if has, err := s.HasQuotes(ctx); err != nil || has {
					t.Fatalf("learning journey changed persistent quotes: %v, %v", has, err)
				}
				if strings.Contains(body, "source=demo") || strings.Contains(body, "quote 2012-") {
					t.Fatal("persistent learning page contains demo exercises")
				}
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/learn", nil))
			if w.Code != http.StatusMethodNotAllowed {
				t.Fatalf("Learn accepted POST: %d", w.Code)
			}
		})
	}
}
