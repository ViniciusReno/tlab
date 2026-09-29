package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
	"math"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/cli"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
	"github.com/ViniciusReno/tlab/internal/web"
)

func TestIPCACLIAPIHTMLParityAndSelicDisplayOnly(t *testing.T) {
	ctx, dir := context.Background(), t.TempDir()
	store, err := sqlite.OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("../pricing/testdata/m3-quotes.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := tesouro.ParseDataset(f, app.IPCASource)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(ctx, rows.Quotes); err != nil {
		t.Fatal(err)
	}
	store.Close()
	for _, tc := range []struct {
		source, id, date, yield, basis string
		want                           float64
		days                           int
	}{
		{"demo", app.DemoIPCABond, "2012-02-17", "3.47", "purchase", 1901.4562720782521, 812},
		{"demo", app.DemoIPCABond, "2012-02-17", "3.51", "mark_to_market", 1897.2254119803258, 813},
		{"demo", app.DemoIPCABond, "2012-02-17", "3.51", "early_exit", 1899.0818683189307, 812},
		{"synced", "ipca:2029-05-15", "2026-09-04", "6.80", "purchase", 3974.0789506153073, 669},
		{"synced", "ipca:2029-05-15", "2026-09-04", "6.92", "mark_to_market", 3962.6249977780267, 670},
		{"synced", "ipca:2029-05-15", "2026-09-04", "6.92", "early_exit", 3962.6249977780267, 670},
		{"synced", "ipca:2050-08-15", "2026-09-04", "6.28", "purchase", 1111.5191713154372, 5994},
	} {
		t.Run(tc.source+"/"+tc.basis, func(t *testing.T) {
			args := []string{"analyze", tc.id, "--source", tc.source, "--basis", tc.basis, "--date", tc.date, "--yield", tc.yield, "--amount", "10000"}
			var s *app.Service
			if tc.source == "demo" {
				s, err = app.OpenDemo(ctx)
			} else {
				s, err = app.OpenPersistent(ctx, dir)
				args = append(args, "--data-dir", dir)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			h, err := web.New(s)
			if err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			if code := cli.Run(ctx, args, &stdout, &stderr); code != 0 {
				t.Fatalf("CLI: %s", stderr.String())
			}
			var c app.Result
			if err := json.Unmarshal(stdout.Bytes(), &c); err != nil {
				t.Fatal(err)
			}
			if c.YieldType != "real" || c.Kind != "ipca" || c.BusinessDays != tc.days || math.Abs(c.ScenarioPU-tc.want) > 1e-9 {
				t.Fatalf("CLI result: %+v", c)
			}
			api := httptest.NewRecorder()
			h.ServeHTTP(api, httptest.NewRequest("GET", "/api/scenario?"+c.Values().Encode(), nil))
			var a app.Result
			if api.Code != 200 {
				t.Fatalf("API: %d %s", api.Code, api.Body.String())
			}
			if err := json.Unmarshal(api.Body.Bytes(), &a); err != nil {
				t.Fatal(err)
			}
			if math.Abs(a.ScenarioPU-c.ScenarioPU) > 1e-9 || a.Quantity == nil || c.Quantity == nil || math.Abs(*a.Quantity-*c.Quantity) > 1e-12 || a.ScenarioAmount == nil || c.ScenarioAmount == nil || math.Abs(*a.ScenarioAmount-*c.ScenarioAmount) > 1e-9 {
				t.Fatal("IPCA parity")
			}
			page := httptest.NewRecorder()
			h.ServeHTTP(page, httptest.NewRequest("GET", "/playground?"+c.Values().Encode(), nil))
			body := html.UnescapeString(page.Body.String())
			for _, want := range []string{"Simulated annual real yield", "indexation base stays fixed", "No future inflation", tc.date, c.Command(), `value="` + tc.id + `"`} {
				if page.Code != 200 || !strings.Contains(body, want) {
					t.Fatalf("HTML missing %q: %d %s", want, page.Code, body)
				}
			}
			if strings.Contains(body, "fixed nominal amount at maturity") || strings.Contains(body, `id="yield-slider"`) {
				t.Fatal("Prefixado explanation or fixed slider leaked into IPCA")
			}
			if tc.source == "demo" && (!strings.Contains(body, app.IPCAFixtureVersion) || strings.Contains(body, "official Tesouro Direto methodology, pages 1–3")) {
				t.Fatal("IPCA provenance")
			}
		})
	}
	s, err := app.OpenPersistent(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h, err := web.New(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/market?matured=1", "/history?bond=selic:2029-03-01"} {
		page := httptest.NewRecorder()
		h.ServeHTTP(page, httptest.NewRequest("GET", path, nil))
		for _, want := range []string{"Selic scenarios unsupported", "BRL 19,795.28", "2026-09-04", "(spread)"} {
			if page.Code != 200 || !strings.Contains(page.Body.String(), want) {
				t.Fatalf("Selic view missing %q: %s", want, page.Body.String())
			}
		}
		if strings.Contains(page.Body.String(), "bond=selic%3A2029-03-01&amp;date=") {
			t.Fatal("Selic scenario shortcut exposed")
		}
	}
	for _, id := range []string{"selic:2029-03-01", "ipca_coupon:2045-05-15", "prefixado_coupon:2035-01-01"} {
		for _, basis := range []string{"purchase", "mark_to_market", "early_exit"} {
			page := httptest.NewRecorder()
			h.ServeHTTP(page, httptest.NewRequest("GET", "/api/scenario?bond="+id+"&basis="+basis+"&yield=7", nil))
			if page.Code != 422 || !strings.Contains(page.Body.String(), "unsupported") {
				t.Fatalf("unsupported math exposed: %s", page.Body.String())
			}
		}
	}
	// Future inflation is not an accepted scenario input or an implicit assumption.
	page := httptest.NewRecorder()
	h.ServeHTTP(page, httptest.NewRequest("GET", "/api/scenario?bond=ipca:2029-05-15&yield=6.8&inflation=4", nil))
	if page.Code != 422 || !strings.Contains(page.Body.String(), "invalid_input") {
		t.Fatal("unplanned inflation projection accepted")
	}
}
