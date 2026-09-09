package tesouro

import (
	"errors"
	"io"
	"math"
	"os"
	"strings"
	"testing"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/bond"
)

func demo(t *testing.T) string {
	t.Helper()
	b, e := assets.Files.ReadFile("data/demo/prefixado.csv")
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}

func TestOfficialDatasetFixture(t *testing.T) {
	f, err := os.Open("testdata/official-quotes.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	const source = "https://www.tesourotransparente.gov.br/ckan/dataset/df56aa42-484a-4a59-8184-7676580c81e3/resource/796d2059-14e9-44e3-80c9-2d9e30b405c1/download/precotaxatesourodireto.csv"
	result, err := ParseDataset(f, source)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Quotes) != 2 || len(result.Unsupported) != 7 || result.Unsupported["Tesouro Prefixado com Juros Semestrais"] != 2 {
		t.Fatalf("unexpected dataset report: %+v", result)
	}
	for _, name := range []string{"Tesouro Selic", "Tesouro IPCA+", "Tesouro IPCA+ com Juros Semestrais", "Tesouro IGPM+ com Juros Semestrais", "Tesouro Renda+ Aposentadoria Extra", "Tesouro Educa+"} {
		if result.Unsupported[name] != 1 {
			t.Fatalf("missing unsupported count for %s", name)
		}
	}
	for i, expected := range []struct {
		id, date string
		values   [5]float64
	}{
		{"prefixado:2032-01-01", "2026-09-04", [5]float64{0.1431, 0.1443, 493.41, 490.42, 490.42}},
		{"prefixado:2015-01-01", "2012-01-03", [5]float64{0.1083, 0.1089, 734.86, 733.67, 733.36}},
	} {
		q := result.Quotes[i]
		if q.Bond.ID != expected.id || q.Date.Format("2006-01-02") != expected.date || q.Source != source || q.Bond.Kind != "prefixado" {
			t.Fatalf("quote identity: %+v", q)
		}
		for j, got := range []*float64{q.BuyYield, q.SellYield, q.BuyPU, q.SellPU, q.BasePU} {
			// Parsing tolerance: 1e-12 for normalized yields, 1e-9 BRL for PUs.
			tolerance := 1e-9
			if j < 2 {
				tolerance = 1e-12
			}
			if got == nil || math.Abs(*got-expected.values[j]) > tolerance {
				t.Fatalf("quote %d field %d: %v", i, j, got)
			}
		}
	}
}

func TestDatasetSchemaAndMissingValues(t *testing.T) {
	// Synthetic variations of the existing fixture test parser rules, not market facts.
	raw := strings.Split(strings.TrimSpace(demo(t)), "\n")
	headers := strings.Split(raw[0], ";")
	row := strings.Split(raw[1], ";")
	for l, r := 0, len(headers)-1; l < r; l, r = l+1, r-1 {
		headers[l], headers[r] = headers[r], headers[l]
		row[l], row[r] = row[r], row[l]
	}
	input := "\ufeff" + strings.Join(headers, ";") + ";Extra\r\n" + strings.Join(row, ";") + ";ignored\r\n"
	result, err := ParseDataset(strings.NewReader(input), "test")
	if err != nil || len(result.Quotes) != 1 {
		t.Fatalf("reordered fields: %+v %v", result, err)
	}
	q := result.Quotes[0]
	if q.SellYield != nil || q.SellPU != nil || q.BasePU != nil {
		t.Fatal("missing values inferred")
	}
	for _, name := range []string{"Tesouro Selic", "Unknown future instrument"} {
		result, err := ParseDataset(strings.NewReader(strings.Replace(demo(t), "Tesouro Prefixado;", name+";", 1)), "test")
		if err != nil || len(result.Quotes) != 0 || result.Unsupported[name] != 1 {
			t.Fatalf("unsupported-only report: %+v %v", result, err)
		}
	}
}

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestDatasetFailureHasNoPartialResults(t *testing.T) {
	valid := demo(t)
	row := strings.SplitN(valid, "\n", 2)[1]
	for _, suffix := range []string{
		"broken;row\n",
		strings.Replace(row, "733,86", "NaN", 1),
		strings.Replace(row, "733,86", "0,00", 1),
		strings.Replace(row, "03/01/2012", "01/01/2015", 1),
		strings.Replace(row, "Tesouro Prefixado", "", 1),
		strings.Replace(strings.Replace(row, "Tesouro Prefixado", "Unsupported", 1), "733,86", "bad", 1),
		strings.Replace(row, "Tesouro Prefixado", "\xff", 1),
	} {
		result, err := ParseDataset(strings.NewReader(valid+suffix), "test")
		if err == nil || result.Quotes != nil || result.Unsupported != nil {
			t.Fatalf("partial/accepted malformed data: %+v %v", result, err)
		}
	}
	if _, err := ParseDataset(io.MultiReader(strings.NewReader(valid), failedReader{}), "test"); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal(err)
	}
	for _, input := range []string{strings.Replace(valid, "Tipo Titulo", "Unknown", 1), strings.Replace(valid, "PU Base Manha", "PU Compra Manha", 1)} {
		if _, err := ParseDataset(strings.NewReader(input), "test"); err != bond.SchemaChanged {
			t.Fatal(err)
		}
	}
	if _, err := ParseDataset(strings.NewReader(valid), " "); err != bond.InvalidInput {
		t.Fatal(err)
	}
	if _, err := ParseDataset(strings.NewReader(strings.SplitN(valid, "\n", 2)[0]+"\n"), "test"); err != bond.MissingQuote {
		t.Fatal(err)
	}
}

func TestDatasetSizeBound(t *testing.T) {
	// Blank lines are valid CSV padding: exercise the actual byte limit without extra rows.
	valid := demo(t)
	input := valid + strings.Repeat("\n", MaxDatasetBytes-len(valid))
	if _, err := ParseDataset(strings.NewReader(input), "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseDataset(io.MultiReader(strings.NewReader(input), strings.NewReader("\n")), "test"); err != bond.InvalidInput {
		t.Fatalf("oversized dataset: %v", err)
	}
}

func TestBrazilianNumbers(t *testing.T) {
	for _, tc := range []struct {
		text string
		want float64
	}{{"733,86", 733.86}, {"1.234,56", 1234.56}, {"-0,50", -0.5}, {"+10,88", 10.88}} {
		n, err := Number(tc.text)
		if err != nil || n == nil || math.Abs(*n-tc.want) > 1e-10 {
			t.Fatalf("%s: %v %v", tc.text, n, err)
		}
	}
	if n, err := Number(""); n != nil || err != nil {
		t.Fatal("missing must remain missing")
	}
	for _, s := range []string{"NaN", "Inf", "1,234.56", "10.88", "12.34,56", "1e3", "abc"} {
		if _, err := Number(s); err == nil {
			t.Fatalf("accepted %q", s)
		}
	}
}
func TestNormalizeOfficialExample(t *testing.T) {
	quotes, err := Parse(strings.NewReader(demo(t)), "official-methodology")
	if err != nil {
		t.Fatal(err)
	}
	q := quotes[0]
	if len(quotes) != 1 || q.Bond.ID != "prefixado:2015-01-01" || q.Date.Format("2006-01-02") != "2012-01-03" {
		t.Fatalf("%+v", quotes)
	}
	if math.Abs(*q.BuyYield-0.1088) > 1e-12 || math.Abs(*q.BuyPU-733.86) > 1e-9 {
		t.Fatal(q)
	}
	if q.SellYield != nil || q.SellPU != nil || q.BasePU != nil {
		t.Fatal("unknown quote fields must not be inferred")
	}
}
func TestRejectMalformedSource(t *testing.T) {
	source := demo(t)
	for _, tc := range []struct {
		from, to string
		want     error
	}{
		{"Tipo Titulo", "Unknown", bond.SchemaChanged},
		{"PU Base Manha", "PU Compra Manha", bond.SchemaChanged},
		{"03/01/2012", "31/02/2012", bond.InvalidInput},
		{"733,86", "bad", bond.InvalidInput},
		{"733,86", "0,00", bond.InvalidInput},
		{"10,88", "-100,00", bond.InvalidInput},
		{"Tesouro Prefixado;", "Tesouro Prefixado com Juros Semestrais;", bond.Unsupported},
		{"Tesouro Prefixado;", "Tesouro IPCA+;", bond.Unsupported},
		{"Tesouro Prefixado;", "Tesouro Selic;", bond.Unsupported},
	} {
		_, err := Parse(strings.NewReader(strings.Replace(source, tc.from, tc.to, 1)), "test")
		if err != tc.want {
			t.Fatalf("%s: %v", tc.to, err)
		}
	}
	if _, err := Parse(strings.NewReader(strings.Repeat("x", (1<<20)+1)), "test"); err != bond.InvalidInput {
		t.Fatal(err)
	}
	if _, err := Parse(strings.NewReader(source+"broken;row\n"), "test"); err != bond.InvalidInput {
		t.Fatal(err)
	}
}
