package app

import (
	"context"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/ViniciusReno/tlab/internal/bond"
)

func TestDemoScenarioAndReproduction(t *testing.T) {
	ctx := context.Background()
	s, err := OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	r, err := ParseRequest(url.Values{"bond": {DemoBond}, "yield": {"8.88"}, "amount": {"10000"}}, "demo")
	if err != nil {
		t.Fatal(err)
	}
	out, err := s.Analyze(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	if out.BusinessDays != 755 || out.SettlementDate != "2012-01-04" || out.QuoteDate != "2012-01-03" {
		t.Fatalf("%+v", out)
	}
	if math.Abs(out.ScenarioPU-774.9918980432596) > 1e-9 {
		t.Fatal(out.ScenarioPU)
	}
	replay, err := ParseRequest(out.Values(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.Analyze(ctx, replay)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(again.ScenarioPU-out.ScenarioPU) > 1e-9 {
		t.Fatal("replay mismatch")
	}
	for _, basis := range []string{"mark_to_market", "early_exit"} {
		r.Basis = basis
		if _, err := s.Analyze(ctx, r); err != bond.MissingQuote {
			t.Fatalf("%s: %v", basis, err)
		}
	}
	r.Basis = "purchase"
	r.Date = "2012-01-04"
	if _, err := s.Analyze(ctx, r); err != bond.MissingQuote {
		t.Fatal(err)
	}
	r.Date = ""
	r.Source = "synced"
	if _, err := s.Analyze(ctx, r); err != bond.SourceMismatch {
		t.Fatal(err)
	}
	r.Source = "demo"
	r.BondID = "selic:2015-01-01"
	if _, err := s.Analyze(ctx, r); err != bond.Unsupported {
		t.Fatal(err)
	}
}

func TestCanonicalInputs(t *testing.T) {
	for _, query := range []string{
		"bond=" + DemoBond + "&yield=NaN", "bond=" + DemoBond + "&yield=Inf", "bond=" + DemoBond + "&yield=-100",
		"bond=" + DemoBond + "&yield=12,00", "bond=" + DemoBond + "&yield=12%25", "bond=" + DemoBond + "&yield=12&amount=0",
		"bond=" + DemoBond + "&yield=12&amount=-1", "bond=" + DemoBond + "&yield=12&date=03/01/2012",
		"bond=" + DemoBond + "&yield=12&source=other", "bond=" + DemoBond + "&yield=12&basis=other",
		"bond=" + DemoBond + "&yield=12&yield=13", "bond=" + DemoBond + "&yield=12&unexpected=x",
	} {
		v, err := url.ParseQuery(query)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ParseRequest(v, "demo"); err == nil {
			t.Fatalf("accepted %s", query)
		}
	}
	r, err := ParseRequest(url.Values{"bond": {DemoBond}, "yield": {"12.00"}}, "synced")
	if err != nil || math.Abs(r.Yield-0.12) > 1e-12 || r.Source != "synced" || r.Basis != "purchase" {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestDemoDoesNotTouchUserData(t *testing.T) {
	// Deliberately use an invalid directory location: demo must never consult it.
	path := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(path, []byte("private portfolio sentinel"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_DATA_HOME", path)
	t.Setenv("XDG_CONFIG_HOME", path)
	t.Setenv("APPDATA", path)
	for i := 0; i < 2; i++ {
		s, err := OpenDemo(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		r, err := s.DefaultRequest(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		out, err := s.Analyze(context.Background(), r)
		s.Close()
		if err != nil || math.Abs(out.BasePU-733.86) > 1e-9 {
			t.Fatalf("%+v %v", out, err)
		}
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != "private portfolio sentinel" {
		t.Fatal("user data changed")
	}
}

func TestPersistentServiceRemainsEmptyAcrossDemo(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := OpenPersistent(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	demo, err := OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := demo.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = OpenPersistent(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if has, err := s.HasQuotes(ctx); err != nil || has || s.Source() != "synced" {
		t.Fatalf("persistent source contaminated: %v %v", has, err)
	}
	if _, err := s.DefaultRequest(ctx); err != bond.SourceUnavailable {
		t.Fatalf("persistent mode selected demo default: %v", err)
	}
}
