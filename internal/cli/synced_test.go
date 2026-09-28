package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
	"math"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/cli"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
	"github.com/ViniciusReno/tlab/internal/web"
)

func TestSyncedCLIAPIHTMLAndMarketLinks(t *testing.T) {
	ctx := context.Background()
	dir := filepath.Join(t.TempDir(), "quotes ' $data")
	store, err := sqlite.OpenPersistent(ctx, dir, assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	f, err := os.Open("../pricing/testdata/quote-contexts.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dataset, err := tesouro.ParseDataset(f, "https://www.tesourotransparente.gov.br/ckan/")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(ctx, dataset.Quotes); err != nil {
		t.Fatal(err)
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
	for _, tc := range []struct {
		basis, yield, price, baseline string
		want                          float64
		days                          int
	}{
		{"purchase", "13.31", "BRL 516.85", "Official purchase price · D+1", 516.8482044401487, 1331},
		{"mark_to_market", "13.43", "BRL 513.71", "Official base PU · D0", 513.7089537612195, 1332},
		{"early_exit", "13.43", "BRL 513.71", "Official early-redemption PU · D0", 513.7089537612195, 1332},
	} {
		t.Run(tc.basis, func(t *testing.T) {
			args := []string{"analyze", "prefixado:2032-01-01", "--data-dir", dir, "--yield", tc.yield, "--amount", "10000"}
			// Purchase also exercises omitted source/basis/date defaults.
			if tc.basis != "purchase" {
				args = append(args, "--source", "synced", "--basis", tc.basis, "--date", "2026-09-04")
			}
			var stdout, stderr bytes.Buffer
			if code := cli.Run(ctx, args, &stdout, &stderr); code != 0 {
				t.Fatalf("CLI: %d %s", code, stderr.String())
			}
			var c app.Result
			if err := json.Unmarshal(stdout.Bytes(), &c); err != nil {
				t.Fatal(err)
			}
			if c.Source != "synced" || c.Basis != tc.basis || c.QuoteDate != "2026-09-04" || c.BusinessDays != tc.days || c.DataDirectory != dir || c.FixtureVersion != "" || math.Abs(c.ScenarioPU-tc.want) > 1e-9 {
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
			if math.Abs(a.ScenarioPU-c.ScenarioPU) > 1e-9 || math.Abs(a.Variation-c.Variation) > 1e-12 || a.SettlementDate != c.SettlementDate || a.CalendarVersion != c.CalendarVersion || a.Quantity == nil || c.Quantity == nil || a.ScenarioAmount == nil || c.ScenarioAmount == nil || math.Abs(*a.Quantity-*c.Quantity) > 1e-12 || math.Abs(*a.ScenarioAmount-*c.ScenarioAmount) > 1e-9 {
				t.Fatal("CLI/API mismatch")
			}
			page := httptest.NewRecorder()
			h.ServeHTTP(page, httptest.NewRequest("GET", "/playground?"+c.Values().Encode(), nil))
			body := html.UnescapeString(page.Body.String())
			if page.Code != 200 {
				t.Fatalf("HTML: %d %s", page.Code, body)
			}
			for _, want := range []string{tc.price, tc.baseline, "2026-09-04", "anbima-2002-2032-v1", c.Command(), `name="source" value="synced"`, "Imported at (UTC, not a market date)"} {
				if !strings.Contains(body, want) {
					t.Fatalf("HTML missing %q", want)
				}
			}
			if strings.Contains(body, "Historical demo") || strings.Contains(body, "prefixado-2012-v1") || strings.Contains(api.Body.String(), "fixture_version") {
				t.Fatal("demo metadata on synchronized result")
			}
			if runtime.GOOS != "windows" && !strings.Contains(c.Command(), "'\"'\"'") {
				t.Fatal("directory apostrophe is not shell-quoted")
			}
		})
	}
	page := httptest.NewRecorder()
	h.ServeHTTP(page, httptest.NewRequest("GET", "/history?bond=prefixado:2032-01-01", nil))
	links := regexp.MustCompile(`href="(/playground\?[^"]+)"`).FindAllStringSubmatch(page.Body.String(), -1)
	if len(links) != 3 {
		t.Fatalf("missing purchase/base/redemption links: %s", page.Body.String())
	}
	for _, link := range links {
		target := html.UnescapeString(link[1])
		u, err := url.Parse(target)
		if err != nil {
			t.Fatal(err)
		}
		v := u.Query()
		if v.Get("source") != "synced" || v.Get("date") != "2026-09-04" || v.Get("basis") == "" || v.Get("yield") == "" || v.Get("data-dir") != "" {
			t.Fatalf("implicit selection: %s", target)
		}
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest("GET", target, nil))
		if r.Code != 200 {
			t.Fatalf("scenario link: %d %s", r.Code, r.Body.String())
		}
	}
	for _, tc := range []struct{ extra, code string }{
		{"&date=2026-09-05", "missing_quote"},
		{"&source=demo", "source_mismatch"},
		{"&data-dir=somewhere", "invalid_input"},
	} {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest("GET", "/api/scenario?bond=prefixado:2032-01-01&yield=12"+tc.extra, nil))
		if r.Code != 422 || !strings.Contains(r.Body.String(), tc.code) {
			t.Fatalf("error contract: %d %s", r.Code, r.Body.String())
		}
	}
	// Synthetic numeric edge: an eligible selected scenario must survive an
	// ineligible optional shock grid, just as it does through the CLI/API.
	q := dataset.Quotes[0]
	pu, yield := 500.0, -0.99
	q.BuyPU, q.BuyYield, q.Source = &pu, &yield, "synthetic-test"
	if err := store.Upsert(ctx, []bond.Quote{q}); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRecorder()
	v := url.Values{"bond": {q.Bond.ID}, "yield": {"-99"}, "source": {"synced"}}
	h.ServeHTTP(r, httptest.NewRequest("GET", "/playground?"+v.Encode(), nil))
	if r.Code != 200 || !strings.Contains(r.Body.String(), "BRL 500.00") || !strings.Contains(r.Body.String(), "default shock range is unavailable") || strings.Contains(r.Body.String(), "NaN") {
		t.Fatalf("optional grid hid valid result: %d %s", r.Code, r.Body.String())
	}
	transition, err := os.Open("../pricing/testdata/redemption-transition.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer transition.Close()
	rows, err := tesouro.ParseDataset(transition, "https://www.tesourotransparente.gov.br/ckan/")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(ctx, rows.Quotes); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		date, yield, convention string
		want                    float64
	}{
		{"2021-09-10", "9.09", "D+1", 785.0360952493704},
		{"2021-09-13", "9.23", "D0", 782.241348739505},
	} {
		var stdout, stderr bytes.Buffer
		args := []string{"analyze", "prefixado:2024-07-01", "--source", "synced", "--basis", "early_exit", "--date", tc.date, "--yield", tc.yield, "--data-dir", dir}
		if code := cli.Run(ctx, args, &stdout, &stderr); code != 0 {
			t.Fatalf("historical CLI: %s", stderr.String())
		}
		var result app.Result
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if math.Abs(result.ScenarioPU-tc.want) > 1e-9 || result.SettlementConvention != tc.convention {
			t.Fatalf("historical result: %+v", result)
		}
		page := httptest.NewRecorder()
		h.ServeHTTP(page, httptest.NewRequest("GET", "/playground?"+result.Values().Encode(), nil))
		body := html.UnescapeString(page.Body.String())
		if page.Code != 200 || !strings.Contains(body, "Official early-redemption PU · "+tc.convention) || !strings.Contains(body, result.Command()) || !strings.Contains(body, "morning-redemption-2021-v1") {
			t.Fatalf("historical HTML: %d %s", page.Code, body)
		}
	}
	for _, path := range []string{"/api/scenario", "/playground"} {
		page := httptest.NewRecorder()
		h.ServeHTTP(page, httptest.NewRequest("GET", path+"?bond=prefixado:2025-01-01&basis=early_exit&date=2021-09-10&yield=9.2", nil))
		if page.Code != 422 || !strings.Contains(page.Body.String(), "does not validate") {
			t.Fatalf("historical mismatch not blocked: %d %s", page.Code, page.Body.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if code := cli.Run(ctx, []string{"analyze", "prefixado:2025-01-01", "--basis", "early_exit", "--date", "2021-09-10", "--yield", "9.2", "--data-dir", dir}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "calculation_not_validated") {
		t.Fatalf("CLI mismatch not blocked: %s", stderr.String())
	}
}
