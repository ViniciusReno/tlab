// Package tesouro normalizes Brazilian source formatting at the ingestion boundary.
package tesouro

import (
	"encoding/csv"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
)

var brazilianNumber = regexp.MustCompile(`^[+-]?(?:[0-9]+|[0-9]{1,3}(?:\.[0-9]{3})+),[0-9]+$`)

// Number returns nil for an explicitly missing optional numeric field.
func Number(s string) (*float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if !brazilianNumber.MatchString(s) {
		return nil, bond.InvalidInput
	}
	n, err := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), ",", "."), 64)
	if err != nil || !bond.Finite(n) {
		return nil, bond.InvalidInput
	}
	return &n, nil
}

func sourceDate(s string) (time.Time, error) {
	d, err := time.Parse("02/01/2006", s)
	if err != nil || d.Format("02/01/2006") != s {
		return time.Time{}, bond.InvalidInput
	}
	return d, nil
}

// Parse reads a bounded CSV. M1 accepts only the implemented no-coupon Prefixado.
func Parse(r io.Reader, source string) ([]bond.Quote, error) {
	const limit = 1 << 20
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit || source == "" {
		return nil, bond.InvalidInput
	}
	reader := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(data), "\ufeff")))
	reader.Comma = ';'
	headers, err := reader.Read()
	if err != nil {
		return nil, bond.SchemaChanged
	}
	indexes := map[string]int{}
	for i, name := range headers {
		if _, exists := indexes[name]; exists {
			return nil, bond.SchemaChanged
		}
		indexes[name] = i
	}
	required := []string{"Tipo Titulo", "Data Vencimento", "Data Base", "Taxa Compra Manha", "Taxa Venda Manha", "PU Compra Manha", "PU Venda Manha", "PU Base Manha"}
	for _, name := range required {
		if _, ok := indexes[name]; !ok {
			return nil, bond.SchemaChanged
		}
	}
	var quotes []bond.Quote
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, bond.InvalidInput
		}
		field := func(name string) string { return row[indexes[name]] }
		if field("Tipo Titulo") != "Tesouro Prefixado" {
			return nil, bond.Unsupported
		}
		maturity, err := sourceDate(field("Data Vencimento"))
		if err != nil {
			return nil, err
		}
		date, err := sourceDate(field("Data Base"))
		if err != nil {
			return nil, err
		}
		if !date.Before(maturity) {
			return nil, bond.InvalidInput
		}
		q := bond.Quote{Bond: bond.Bond{ID: "prefixado:" + maturity.Format(time.DateOnly), Kind: "prefixado", Name: field("Tipo Titulo"), Maturity: maturity}, Date: date, Source: source}
		targets := []**float64{&q.BuyYield, &q.SellYield, &q.BuyPU, &q.SellPU, &q.BasePU}
		for i, name := range required[3:] {
			n, err := Number(field(name))
			if err != nil {
				return nil, err
			}
			if n != nil {
				if i < 2 {
					*n /= 100
					if !bond.ValidYield(*n) {
						return nil, bond.InvalidInput
					}
				} else if !bond.Positive(*n) {
					return nil, bond.InvalidInput
				}
			}
			*targets[i] = n
		}
		quotes = append(quotes, q)
	}
	if len(quotes) == 0 {
		return nil, bond.MissingQuote
	}
	return quotes, nil
}
