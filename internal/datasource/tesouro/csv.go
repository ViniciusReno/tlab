// Package tesouro normalizes Brazilian source formatting at the ingestion boundary.
package tesouro

import (
	"encoding/csv"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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

// MaxDatasetBytes bounds the official CSV (about 14 MiB at M2.2 verification).
const MaxDatasetBytes = 32 << 20

// Dataset reports excluded instrument rows explicitly; only Prefixado is imported in M2.
type Dataset struct {
	Quotes      []bond.Quote
	Unsupported map[string]int
}

// Parse is the strict, small-fixture entry point used by the offline demo.
func Parse(r io.Reader, source string) ([]bond.Quote, error) {
	result, err := parse(r, source, 1<<20, false)
	return result.Quotes, err
}

// ParseDataset validates a mixed official CSV without approximating other instruments.
// On failure it returns no partial results. It performs no network or storage I/O.
func ParseDataset(r io.Reader, source string) (Dataset, error) {
	return parse(r, source, MaxDatasetBytes, true)
}

func parse(r io.Reader, source string, limit int64, mixed bool) (Dataset, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return Dataset{}, err
	}
	if int64(len(data)) > limit || strings.TrimSpace(source) == "" || !utf8.Valid(data) {
		return Dataset{}, bond.InvalidInput
	}
	reader := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(data), "\ufeff")))
	reader.Comma = ';'
	headers, err := reader.Read()
	if err != nil {
		return Dataset{}, bond.SchemaChanged
	}
	indexes := map[string]int{}
	for i, name := range headers {
		if _, exists := indexes[name]; exists {
			return Dataset{}, bond.SchemaChanged
		}
		indexes[name] = i
	}
	required := []string{"Tipo Titulo", "Data Vencimento", "Data Base", "Taxa Compra Manha", "Taxa Venda Manha", "PU Compra Manha", "PU Venda Manha", "PU Base Manha"}
	for _, name := range required {
		if _, ok := indexes[name]; !ok {
			return Dataset{}, bond.SchemaChanged
		}
	}
	result := Dataset{Unsupported: make(map[string]int)}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Dataset{}, bond.InvalidInput
		}
		field := func(name string) string { return row[indexes[name]] }
		name := field("Tipo Titulo")
		if strings.TrimSpace(name) == "" {
			return Dataset{}, bond.InvalidInput
		}
		if name != "Tesouro Prefixado" && !mixed {
			return Dataset{}, bond.Unsupported
		}
		maturity, err := sourceDate(field("Data Vencimento"))
		if err != nil {
			return Dataset{}, err
		}
		date, err := sourceDate(field("Data Base"))
		if err != nil {
			return Dataset{}, err
		}
		if !date.Before(maturity) {
			return Dataset{}, bond.InvalidInput
		}
		if name != "Tesouro Prefixado" {
			// Validate source syntax even for excluded rows. Do not create domain
			// prices or apply Prefixado validity rules to unsupported instruments.
			for _, column := range required[3:] {
				if _, err := Number(field(column)); err != nil {
					return Dataset{}, err
				}
			}
			result.Unsupported[name]++
			continue
		}
		q := bond.Quote{Bond: bond.Bond{ID: "prefixado:" + maturity.Format(time.DateOnly), Kind: "prefixado", Name: field("Tipo Titulo"), Maturity: maturity}, Date: date, Source: source}
		targets := []**float64{&q.BuyYield, &q.SellYield, &q.BuyPU, &q.SellPU, &q.BasePU}
		for i, name := range required[3:] {
			n, err := Number(field(name))
			if err != nil {
				return Dataset{}, err
			}
			if n != nil {
				if i < 2 {
					*n /= 100
					if !bond.ValidYield(*n) {
						return Dataset{}, bond.InvalidInput
					}
				} else if !bond.Positive(*n) {
					return Dataset{}, bond.InvalidInput
				}
			}
			*targets[i] = n
		}
		result.Quotes = append(result.Quotes, q)
	}
	if len(result.Quotes) == 0 && len(result.Unsupported) == 0 {
		return Dataset{}, bond.MissingQuote
	}
	return result, nil
}
