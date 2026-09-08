package tesouro

import (
	"math"
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
