package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/portfolio"
)

type PositionView struct {
	portfolio.Position
	Quote                                              bond.Quote
	Status, Completeness, ComparisonError, ChangeError string
	Value, Change, ChangeRate                          *float64
	Comparison                                         *portfolio.Comparison
	Scenario                                           *Result
	ScenarioValue                                      *float64
}

type Portfolio struct {
	Positions                                              []PositionView
	Bonds                                                  []bond.Bond
	Total                                                  *float64
	Valued, Unavailable                                    int
	Allocations                                            []portfolio.Allocation
	FirstQuoteDate, LastQuoteDate, DatasetDate, ImportedAt string
	Freshness                                              DatasetFreshness
	TotalError                                             string
}

func ParsePosition(v url.Values) (portfolio.Position, error) {
	p := portfolio.Position{ID: v.Get("id"), BondID: v.Get("bond"), Note: v.Get("note")}
	for k, values := range v {
		switch k {
		case "id", "bond", "quantity", "purchase_date", "purchase_pu", "purchase_yield", "invested_brl", "note", "token", "action":
		default:
			return p, bond.InvalidInput
		}
		if len(values) != 1 {
			return p, bond.InvalidInput
		}
	}
	for key, dest := range map[string]**float64{"purchase_pu": &p.PurchasePU, "purchase_yield": &p.PurchaseYield, "invested_brl": &p.InvestedBRL} {
		if text := v.Get(key); text != "" {
			n, err := number(text)
			if err != nil {
				return p, err
			}
			if key == "purchase_yield" {
				n /= 100
			}
			*dest = &n
		}
	}
	if text := v.Get("quantity"); text != "" {
		n, err := number(text)
		if err != nil || !bond.Positive(n) {
			return p, bond.InvalidInput
		}
		p.Quantity = n
	}
	if text := v.Get("purchase_date"); text != "" {
		d, err := bond.ParseDate(text)
		if err != nil {
			return p, err
		}
		p.PurchaseDate = &d
	}
	return portfolio.Normalize(p)
}

func (s *Service) SavePosition(ctx context.Context, p portfolio.Position) error {
	p, err := portfolio.Normalize(p)
	if err != nil {
		return err
	}
	q, err := s.store.Quote(ctx, p.BondID, "")
	if err != nil {
		return err
	}
	if q.Bond.Kind != "prefixado" && q.Bond.Kind != "ipca" && q.Bond.Kind != "selic" {
		return bond.Unsupported
	}
	create := p.ID == ""
	if create {
		var id [16]byte
		if _, err := rand.Read(id[:]); err != nil {
			return err
		}
		p.ID = hex.EncodeToString(id[:])
	}
	return s.store.SavePosition(ctx, p, create)
}
func (s *Service) DeletePosition(ctx context.Context, id string) error {
	return s.store.DeletePosition(ctx, id)
}

func (s *Service) Portfolio(ctx context.Context, now time.Time) (Portfolio, error) {
	var out Portfolio
	data, err := s.store.Market(ctx)
	if err != nil {
		return out, err
	}
	positions, err := s.store.Positions(ctx)
	if err != nil {
		return out, err
	}
	out.DatasetDate, out.ImportedAt = data.DatasetMaxQuoteDate, data.ImportedAt
	out.Freshness = s.datasetFreshness(out.DatasetDate, now)
	quotes := make(map[string]bond.Quote)
	for _, q := range data.Quotes {
		quotes[q.Bond.ID] = q
		out.Bonds = append(out.Bonds, q.Bond)
	}
	var values []portfolio.Allocation
	for _, p := range positions {
		q := quotes[p.BondID]
		v := PositionView{Position: p, Quote: q, Status: "available", Completeness: "valuation_only"}
		if p.InvestedBRL != nil {
			v.Completeness = "cost_basis_known"
		}
		switch {
		case q.Bond.Kind != "prefixado" && q.Bond.Kind != "ipca" && q.Bond.Kind != "selic":
			v.Status = "unsupported"
		case now.Format(time.DateOnly) >= q.Bond.Maturity.Format(time.DateOnly):
			v.Status = "matured"
		case q.Date.IsZero() || q.BasePU == nil:
			v.Status = "missing_base_quote"
		case q.Date.Format(time.DateOnly) > now.Format(time.DateOnly):
			v.Status = "future_quote_date"
		default:
			n, e := portfolio.Value(p.Quantity, *q.BasePU)
			if e != nil {
				v.Status = string(bond.CalculationOutOfRange)
			} else {
				v.Value = &n
				date := q.Date.Format(time.DateOnly)
				if out.FirstQuoteDate == "" || date < out.FirstQuoteDate {
					out.FirstQuoteDate = date
				}
				if date > out.LastQuoteDate {
					out.LastQuoteDate = date
				}
				if out.DatasetDate != "" && date < out.DatasetDate {
					v.Status = "older_quote"
				} else if out.DatasetDate == "" {
					v.Status = "freshness_unknown"
				}
				if p.InvestedBRL != nil {
					delta, rate, e := portfolio.Change(n, *p.InvestedBRL)
					if e == nil {
						v.Change, v.ChangeRate = &delta, &rate
					} else {
						v.ChangeError = e.Error()
					}
				}
			}
		}
		if v.Value != nil {
			out.Valued++
			values = append(values, portfolio.Allocation{Kind: q.Bond.Kind, Value: *v.Value})
		} else {
			out.Unavailable++
		}
		if q.Bond.Kind != "prefixado" {
			v.ComparisonError = string(bond.Unsupported)
		} else if v.Status == "matured" {
			v.ComparisonError = string(bond.NoRemainingTerm)
		} else {
			comparison, e := portfolio.HoldExit(q, p.Quantity, s.calendar)
			if e != nil {
				v.ComparisonError = e.Error()
			} else {
				v.Comparison = &comparison
			}
		}
		out.Positions = append(out.Positions, v)
	}
	total, allocations, aggregateErr := portfolio.Aggregate(values)
	if aggregateErr != nil {
		out.TotalError = aggregateErr.Error()
	} else if len(values) > 0 {
		out.Total = &total
		out.Allocations = allocations
	}

	return out, nil
}

// PositionScenario uses the stored quantity directly, without reconstructing it
// from a displayed or rounded amount. Acquisition details never enter the URL.
func (s *Service) PositionScenario(ctx context.Context, id, date string, yield float64) (Result, float64, error) {
	if date == "" {
		return Result{}, 0, bond.InvalidInput
	}
	positions, err := s.store.Positions(ctx)
	if err != nil {
		return Result{}, 0, err
	}
	for _, p := range positions {
		if p.ID == id {
			result, err := s.Analyze(ctx, Request{BondID: p.BondID, Source: s.source, Basis: "mark_to_market", Date: date, Yield: yield})
			if err != nil {
				return Result{}, 0, err
			}
			amount, err := portfolio.Value(p.Quantity, result.ScenarioPU)
			return result, amount, err
		}
	}
	return Result{}, 0, bond.InvalidInput
}
