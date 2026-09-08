// Package app selects official inputs and shares the same scenario workflow across CLI and HTTP.
package app

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/calendar"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/pricing"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
)

const (
	DemoBond       = "prefixado:2015-01-01"
	FixtureVersion = "prefixado-2012-v1"
	OfficialSource = "https://www.tesourodireto.com.br/documents/d/guest/tesouro_prefixado"
	BuildVersion   = "0.1.0-dev"
)

type Service struct {
	store    *sqlite.Store
	calendar *calendar.Calendar
	source   string
}

// DefaultDataDir follows the OS user configuration directory convention.
func DefaultDataDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tesouro-lab"), nil
}

func OpenPersistent(ctx context.Context, directory string) (*Service, error) {
	store, err := sqlite.OpenPersistent(ctx, directory, assets.Files)
	if err != nil {
		return nil, err
	}
	return &Service{store: store, source: "synced"}, nil
}

func (s *Service) Source() string { return s.source }

func (s *Service) HasQuotes(ctx context.Context) (bool, error) {
	return s.store.HasQuotes(ctx)
}

func OpenDemo(ctx context.Context) (*Service, error) {
	f, err := assets.Files.Open("data/demo/calendar.json")
	if err != nil {
		return nil, err
	}
	cal, err := calendar.Load(f)
	f.Close()
	if err != nil {
		return nil, err
	}
	f, err = assets.Files.Open("data/demo/prefixado.csv")
	if err != nil {
		return nil, err
	}
	quotes, err := tesouro.Parse(f, OfficialSource)
	f.Close()
	if err != nil {
		return nil, err
	}
	store, err := sqlite.OpenMemory(ctx, assets.Files)
	if err != nil {
		return nil, err
	}
	if err = store.Upsert(ctx, quotes); err != nil {
		store.Close()
		return nil, err
	}
	return &Service{store: store, calendar: cal, source: "demo"}, nil
}

func (s *Service) Close() error { return s.store.Close() }

type Request struct {
	BondID, Source, Basis, Date string
	Yield                       float64
	Amount                      *float64
}

var decimal = regexp.MustCompile(`^[+-]?[0-9]+(?:\.[0-9]+)?$`)

func number(text string) (float64, error) {
	if !decimal.MatchString(text) {
		return 0, bond.InvalidInput
	}
	n, err := strconv.ParseFloat(text, 64)
	if err != nil || !bond.Finite(n) {
		return 0, bond.InvalidInput
	}
	return n, nil
}

// ParseRequest consumes canonical English inputs; yields arrive as percentages.
func ParseRequest(values url.Values, defaultSource string) (Request, error) {
	r := Request{BondID: values.Get("bond"), Source: defaultSource, Basis: "purchase"}
	for key, v := range values {
		switch key {
		case "bond", "source", "basis", "date", "yield", "amount":
		default:
			return r, bond.InvalidInput
		}
		if len(v) != 1 {
			return r, bond.InvalidInput
		}
	}
	if v := values.Get("source"); v != "" {
		r.Source = v
	}
	if v := values.Get("basis"); v != "" {
		r.Basis = v
	}
	r.Date = values.Get("date")
	if r.Date != "" {
		if _, err := bond.ParseDate(r.Date); err != nil {
			return r, err
		}
	}
	n, err := number(values.Get("yield"))
	if err != nil {
		return r, err
	}
	r.Yield = n / 100
	if a := values.Get("amount"); a != "" {
		n, err := number(a)
		if err != nil || !bond.Positive(n) {
			return r, bond.InvalidInput
		}
		r.Amount = &n
	}
	if err := r.validate(); err != nil {
		return r, err
	}
	return r, nil
}

func (r Request) validate() error {
	if r.BondID == "" || !bond.ValidYield(r.Yield) || (r.Amount != nil && !bond.Positive(*r.Amount)) {
		return bond.InvalidInput
	}
	if r.Source != "demo" && r.Source != "synced" {
		return bond.InvalidInput
	}
	switch r.Basis {
	case "purchase", "mark_to_market", "early_exit":
	default:
		return bond.InvalidInput
	}
	if r.Date != "" {
		if _, err := bond.ParseDate(r.Date); err != nil {
			return err
		}
	}
	return nil
}

type Result struct {
	pricing.Result
	BondID             string `json:"bond_id"`
	Name               string `json:"name"`
	Source             string `json:"source"`
	Basis              string `json:"basis"`
	QuoteDate          string `json:"quote_date"`
	SettlementDate     string `json:"settlement_date"`
	Maturity           string `json:"maturity"`
	Provenance         string `json:"provenance"`
	CalendarVersion    string `json:"calendar_version"`
	FixtureVersion     string `json:"fixture_version"`
	CalculationVersion string `json:"calculation_version"`
	BuildVersion       string `json:"build_version"`
}

func (s *Service) DefaultRequest(ctx context.Context) (Request, error) {
	if s.source != "demo" {
		return Request{}, bond.SourceUnavailable
	}
	q, err := s.store.Quote(ctx, DemoBond, "")
	if err != nil {
		return Request{}, err
	}
	if q.BuyYield == nil {
		return Request{}, bond.MissingQuote
	}
	return Request{BondID: DemoBond, Source: "demo", Basis: "purchase", Date: q.Date.Format(time.DateOnly), Yield: *q.BuyYield}, nil
}

func (s *Service) Analyze(ctx context.Context, r Request) (Result, error) {
	if err := r.validate(); err != nil {
		return Result{}, err
	}
	if r.Source != s.source {
		return Result{}, bond.SourceMismatch
	}
	if s.source != "demo" {
		return Result{}, bond.SourceUnavailable
	}
	if !strings.HasPrefix(r.BondID, "prefixado:") {
		return Result{}, bond.Unsupported
	}
	q, err := s.store.Quote(ctx, r.BondID, r.Date)
	if err != nil {
		return Result{}, err
	}
	if q.Bond.Kind != "prefixado" {
		return Result{}, bond.Unsupported
	}
	settlement := q.Date
	var pu, yield *float64
	switch r.Basis {
	case "purchase":
		pu, yield = q.BuyPU, q.BuyYield
	case "mark_to_market":
		pu, yield = q.BasePU, q.SellYield
	case "early_exit":
		pu, yield = q.SellPU, q.SellYield
	}
	if pu == nil || yield == nil {
		return Result{}, bond.MissingQuote
	}
	if r.Basis != "mark_to_market" {
		settlement, err = s.calendar.Next(q.Date)
		if err != nil {
			return Result{}, err
		}
	}
	days, err := s.calendar.Count(settlement, q.Bond.Maturity)
	if err != nil {
		return Result{}, err
	}
	calculated, err := pricing.Scenario(pricing.Input{BasePU: *pu, BaseYield: *yield, ScenarioYield: r.Yield, BusinessDays: days, Amount: r.Amount})
	if err != nil {
		return Result{}, err
	}
	return Result{Result: calculated, BondID: q.Bond.ID, Name: q.Bond.Name, Source: r.Source, Basis: r.Basis,
		QuoteDate: q.Date.Format(time.DateOnly), SettlementDate: settlement.Format(time.DateOnly), Maturity: q.Bond.Maturity.Format(time.DateOnly),
		Provenance: q.Source, CalendarVersion: s.calendar.Version, FixtureVersion: FixtureVersion, CalculationVersion: pricing.Version, BuildVersion: BuildVersion}, nil
}

func Decimal(n float64) string { return strconv.FormatFloat(n, 'f', -1, 64) }

// Values encodes resolved inputs, not rounded display amounts.
func (r Result) Values() url.Values {
	v := url.Values{"bond": {r.BondID}, "source": {r.Source}, "basis": {r.Basis}, "date": {r.QuoteDate}, "yield": {Decimal(r.ScenarioYield * 100)}}
	if r.Amount != nil {
		v.Set("amount", Decimal(*r.Amount))
	}
	return v
}

func (r Result) Command() string {
	c := fmt.Sprintf("go run ./cmd/tesouro-lab analyze %s --source %s --basis %s --date %s --yield %s",
		r.BondID, r.Source, r.Basis, r.QuoteDate, Decimal(r.ScenarioYield*100))
	if r.Amount != nil {
		c += " --amount " + Decimal(*r.Amount)
	}
	return c
}

type Point struct {
	Yield       float64 `json:"yield"`
	PU          float64 `json:"pu"`
	Variation   float64 `json:"variation"`
	BasisPoints int     `json:"basis_points"`
}

// Shocks uses the same pricing function and baseline as the selected scenario.
func Shocks(r Result) ([]Point, error) {
	points := make([]Point, 0, 9)
	for bps := -400; bps <= 400; bps += 100 {
		in := r.Input
		in.ScenarioYield = r.BaseYield + float64(bps)/10000
		in.Amount = nil
		out, err := pricing.Scenario(in)
		if err != nil {
			return nil, err
		}
		points = append(points, Point{Yield: in.ScenarioYield, PU: out.ScenarioPU, Variation: out.Variation, BasisPoints: bps})
	}
	return points, nil
}
