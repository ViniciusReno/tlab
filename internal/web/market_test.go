package web_test

import (
	"context"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
	"github.com/ViniciusReno/tlab/internal/web"
)

func TestMarketAndHistoryOfficialFieldsAndIsolation(t *testing.T) {
	ctx, dir := context.Background(), t.TempDir()
	store, err := sqlite.OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../datasource/tesouro/testdata/official-quotes.csv")
	if err != nil {
		t.Fatal(err)
	}
	dataset, err := tesouro.ParseDataset(strings.NewReader(string(data)), "https://www.tesourotransparente.gov.br/test")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(ctx, dataset.Quotes); err != nil {
		t.Fatal(err)
	}
	store.Close()
	s, err := app.OpenPersistent(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h, err := web.New(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/market?matured=1", "/history?bond=prefixado:2032-01-01"} {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 200 {
			t.Fatalf("%s: %d %s", path, r.Code, r.Body.String())
		}
		for _, want := range []string{"2026-09-04", "BRL 493.41", "BRL 490.42", "Official base PU", "Imported at (UTC)", "Unavailable", "<summary>", "https://www.tesourotransparente.gov.br/test"} {
			if !strings.Contains(r.Body.String(), want) {
				t.Fatalf("%s missing %s", path, want)
			}
		}
		for _, unwanted := range []string{"scenario-form", "<script", "Historical demo", "#ZgotmplZ"} {
			if strings.Contains(r.Body.String(), unwanted) {
				t.Fatalf("unexpected %s", unwanted)
			}
		}
	}
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/market?matured=1", nil))
	if !strings.Contains(r.Body.String(), "Matured on 2015-01-01") || !strings.Contains(r.Body.String(), "Historical view only") {
		t.Fatal("matured market value exposed")
	}
	r = httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/", nil))
	if strings.Contains(r.Body.String(), "2015-01-01") {
		t.Fatal("matured bond in default market")
	}
	for _, path := range []string{"/history", "/history?bond=unknown", "/history?bond=prefixado:2032-01-01&before=2026-02-30", "/history?bond=a&bond=b", "/market?matured=0", "/market?source=demo", "/market?file=/tmp/data", "/market?%zz"} {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 422 {
			t.Fatalf("accepted %s: %d", path, r.Code)
		}
	}
	demo, err := app.OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer demo.Close()
	dh, err := web.New(demo)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/market", "/history?bond=prefixado:2032-01-01"} {
		r := httptest.NewRecorder()
		dh.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != 422 || !strings.Contains(r.Body.String(), "does not match this server") {
			t.Fatalf("demo source isolation: %d", r.Code)
		}
	}
}
