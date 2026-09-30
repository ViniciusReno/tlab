package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ViniciusReno/tlab/internal/app"
)

func TestPortfolioFormsSecurityEscapingAndDemoReset(t *testing.T) {
	ctx := context.Background()
	s, err := app.OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h, err := newHandler(s, func() time.Time { return time.Date(2012, 2, 17, 12, 0, 0, 0, time.UTC) })
	if err != nil {
		t.Fatal(err)
	}
	get := func(path string) string {
		t.Helper()
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || !strings.Contains(w.Body.String(), "</html>") {
			t.Fatalf("render %d: %s", w.Code, w.Body.String())
		}
		return w.Body.String()
	}
	body := get("/portfolio")
	for _, text := range []string{"Demo data · temporary edits", "BRL 3,678.56", "Change since purchase is unavailable", "Composition of the valued positions"} {
		if !strings.Contains(body, text) {
			t.Fatalf("missing %s", text)
		}
	}
	token := regexp.MustCompile(`name="token" value="([a-f0-9]+)"`).FindStringSubmatch(body)[1]
	post := func(values url.Values, origin, media string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/portfolio", strings.NewReader(values.Encode()))
		r.Header.Set("Content-Type", media)
		r.Header.Set("Origin", origin)
		h.ServeHTTP(w, r)
		return w
	}
	v := url.Values{"token": {token}, "action": {"save"}, "bond": {app.DemoIPCABond}, "quantity": {"2"}, "note": {"<script>alert(1)</script>"}}
	if w := post(v, "http://hostile.example", "application/x-www-form-urlencoded"); w.Code != 403 {
		t.Fatal("cross-origin accepted")
	}
	v.Set("token", "invalid")
	if w := post(v, "", "application/x-www-form-urlencoded"); w.Code != 403 {
		t.Fatal("invalid token accepted")
	}
	v.Set("token", token)
	if w := post(v, "", "text/plain"); w.Code != 415 {
		t.Fatal("unexpected content type accepted")
	}
	v.Set("quantity", "0")
	if w := post(v, "", "application/x-www-form-urlencoded"); w.Code != 422 || !strings.Contains(w.Body.String(), "Invalid input") || !strings.Contains(w.Body.String(), "</html>") {
		t.Fatalf("validation %d %s", w.Code, w.Body.String())
	}
	v.Set("quantity", "2")
	if w := post(v, "", "application/x-www-form-urlencoded"); w.Code != 303 || w.Header().Get("Location") != "/portfolio" {
		t.Fatalf("save %d %s", w.Code, w.Body.String())
	}
	body = get("/portfolio")
	if strings.Contains(body, "<script>") || !strings.Contains(body, "&lt;script&gt;") {
		t.Fatal("note not escaped")
	}
	data, err := s.Portfolio(ctx, time.Date(2012, 2, 17, 12, 0, 0, 0, time.UTC))
	if err != nil || len(data.Positions) != 2 {
		t.Fatal(err)
	}
	id := data.Positions[1].ID
	body = get("/portfolio?edit=" + id)
	if !strings.Contains(body, "Edit position") || !strings.Contains(body, "Cancel editing") {
		t.Fatal("edit form absent")
	}
	v.Set("id", id)
	v.Set("quantity", "3")
	v.Set("note", "updated")
	if w := post(v, "", "application/x-www-form-urlencoded"); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	body = get("/portfolio?scenario=" + id + "&date=2012-02-17&yield=3.51")
	for _, text := range []string{"BRL 5,691.68", "Settlement 2012-02-17", "mark_to_market", "Inspect unit-price calculation"} {
		if !strings.Contains(body, text) {
			t.Fatalf("scenario missing %s: %s", text, body)
		}
	}
	if strings.Contains(body, "#ZgotmplZ") || strings.Contains(body, "note=updated") {
		t.Fatal("unsafe or private scenario URL")
	}
	removal := url.Values{"action": {"remove"}, "id": {id}, "token": {token}}
	if w := post(removal, "", "application/x-www-form-urlencoded"); w.Code != 422 {
		t.Fatal("removed without confirmation")
	}
	removal.Set("confirm", "yes")
	if w := post(removal, "", "application/x-www-form-urlencoded"); w.Code != 303 {
		t.Fatal(w.Body.String())
	}
	v.Del("id")
	v.Set("note", strings.Repeat("x", 17000))
	if w := post(v, "", "application/x-www-form-urlencoded"); w.Code != http.StatusBadRequest {
		t.Fatalf("oversized form: %d", w.Code)
	}
	get("/portfolio")
}

func TestGrossLossCurrencyFormatting(t *testing.T) {
	for _, tc := range []struct {
		value float64
		want  string
	}{
		{-100, "BRL -100.00"}, {-100000, "BRL -100,000.00"}, {-1234.5, "BRL -1,234.50"}, {1000, "BRL 1,000.00"}, {0, "BRL 0.00"},
	} {
		if got := money(tc.value); got != tc.want {
			t.Fatalf("got %s want %s", got, tc.want)
		}
	}
}
