package web

import (
	"context"
	"encoding/json"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
)

func TestTranslationCoverage(t *testing.T) {
	files, err := assets.Files.ReadDir("web/templates")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := assets.Files.ReadFile("web/templates/" + file.Name())
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range regexp.MustCompile(`\btr\s+("(?:[^"\\]|\\.)*")`).FindAllStringSubmatch(string(data), -1) {
			var key string
			if err := json.Unmarshal([]byte(match[1]), &key); err != nil {
				t.Fatal(err)
			}
			if portuguese[key] == "" {
				t.Errorf("%s: missing Portuguese copy for %q", file.Name(), key)
			}
		}
	}
	for _, code := range []bond.Error{bond.InvalidInput, bond.MissingQuote, bond.Unsupported, bond.SourceUnavailable, bond.SyncFailed, bond.SourceMismatch, bond.SchemaChanged, bond.CalendarOutOfRange, bond.NoRemainingTerm, bond.CalculationOutOfRange, bond.CalculationNotValidated} {
		if portuguese[bond.Message(code)] == "" {
			t.Errorf("missing domain error translation: %s", code)
		}
	}
}

func TestPortugueseOfficialMarketAndHistory(t *testing.T) {
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
	h, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/market?matured=1", "/history?bond=prefixado:2032-01-01"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path+"&lang=pt-BR", nil))
		if w.Code != 200 || !strings.Contains(w.Body.String(), "</html>") {
			t.Fatalf("%s: %d", path, w.Code)
		}
		for _, text := range []string{"2026-09-04", "BRL 493.41", "BRL 490.42", "PU Base oficial", "Importado em (UTC)", "Indisponível", "https://www.tesourotransparente.gov.br/test"} {
			if !strings.Contains(w.Body.String(), text) {
				t.Fatalf("%s: missing %s", path, text)
			}
		}
		if strings.HasPrefix(path, "/market") && !strings.Contains(w.Body.String(), "Vencido em 2015-01-01") {
			t.Fatal("matured state not translated")
		}
	}
}

func TestLanguageSelectionAndCanonicalScenario(t *testing.T) {
	s, err := app.OpenDemo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	get := func(path string, cookie *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("GET", path, nil)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	query := "bond=ipca%3A2015-05-15&source=demo&basis=mark_to_market&date=2012-02-17&yield=4.51&amount=10000.50"
	en := get("/playground?"+query, nil)
	pt := get("/playground?"+query+"&lang=pt-BR", nil)
	if en.Code != 200 || pt.Code != 200 {
		t.Fatalf("scenario: %d / %d", en.Code, pt.Code)
	}
	if !strings.Contains(pt.Body.String(), `<html lang="pt-BR">`) || !strings.Contains(pt.Body.String(), "Preço no cenário hipotético") || pt.Header().Get("Content-Language") != "pt-BR" {
		t.Fatal("Portuguese scenario was not localized")
	}
	// Presentation retains exact inputs, money, CLI commands, and JSON links.
	for _, pattern := range []string{`<h2>(BRL [^<]+)</h2>`, `<pre><code>([^<]+)</code></pre>`, `href="(/api/scenario\?[^"]+)"`} {
		re := regexp.MustCompile(pattern)
		if strings.Join(re.FindAllString(en.Body.String(), -1), "\n") != strings.Join(re.FindAllString(pt.Body.String(), -1), "\n") {
			t.Fatalf("language changed canonical result: %s", pattern)
		}
	}
	cookies := pt.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != languageCookie || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode || cookies[0].Path != "/" {
		t.Fatal("invalid preference cookie")
	}
	for _, path := range []string{"/playground", "/learn", "/portfolio"} {
		w := get(path, cookies[0])
		if w.Code != 200 || !strings.Contains(w.Body.String(), `<html lang="pt-BR">`) || !strings.Contains(w.Body.String(), "Pular para o conteúdo") || !strings.Contains(w.Body.String(), "Idioma") {
			t.Fatalf("language not retained on %s: %d", path, w.Code)
		}
	}
	if a, b := get("/api/scenario?"+query, nil), get("/api/scenario?"+query, cookies[0]); a.Code != 200 || a.Body.String() != b.Body.String() {
		t.Fatal("cookie altered canonical API result")
	}
	ptError := get("/playground?bond=prefixado%3A2015-01-01&source=demo&basis=mark_to_market&date=2012-01-03&yield=10.88", cookies[0])
	if ptError.Code != 422 || !strings.Contains(ptError.Body.String(), "Não há cotação oficial") {
		t.Fatal("missing-context error not translated")
	}
	for _, query := range []string{"lang=fr", "lang=en&lang=pt-BR", "lang=%3Cscript%3E"} {
		w := get("/learn?"+query, nil)
		if w.Code != 400 || strings.Contains(w.Body.String(), "<script>") || len(w.Result().Cookies()) != 0 {
			t.Fatalf("invalid preference accepted: %s", query)
		}
	}
	if w := get("/playground?"+query+"&yield=4.51&lang=pt-BR", nil); w.Code != 422 {
		t.Fatal("language query hid duplicate financial input")
	}
	if w := get("/playground?lang=pt-BR", nil); w.Code != 200 {
		t.Fatal("presentation-only query stopped demo defaults")
	}
	if w := get("/learn?lang=en", cookies[0]); !strings.Contains(w.Body.String(), `<html lang="en">`) || w.Result().Cookies()[0].Value != "en" {
		t.Fatal("explicit English selection did not override preference")
	}
	// Switching language retains every resolved scenario input in the link.
	for _, match := range regexp.MustCompile(`href="([^"]*lang=en[^"]*)"`).FindAllStringSubmatch(pt.Body.String(), -1) {
		u, err := url.Parse(html.UnescapeString(match[1]))
		if err != nil || u.Query().Get("date") != "2012-02-17" || u.Query().Get("yield") != "4.51" || u.Query().Get("amount") != "10000.50" {
			t.Fatal("language link lost scenario inputs")
		}
	}
}

func TestPortuguesePortfolioFormsAndPersistentEmptyState(t *testing.T) {
	for _, demo := range []bool{true, false} {
		var s *app.Service
		var err error
		if demo {
			s, err = app.OpenDemo(context.Background())
		} else {
			s, err = app.OpenPersistent(context.Background(), t.TempDir())
		}
		if err != nil {
			t.Fatal(err)
		}
		h, err := newHandler(s, func() time.Time { return time.Date(2012, 2, 17, 12, 0, 0, 0, time.UTC) })
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", "/portfolio?lang=pt-BR", nil))
		if w.Code != 200 || !strings.Contains(w.Body.String(), "Carteira") || !strings.Contains(w.Body.String(), "Use ponto decimal") {
			t.Fatal("Portuguese portfolio not rendered")
		}
		if demo {
			match := regexp.MustCompile(`name="token" value="([a-f0-9]+)"`).FindStringSubmatch(w.Body.String())
			if len(match) != 2 {
				t.Fatal("missing demo form token")
			}
			token := match[1]
			// The token from Portuguese is valid after selecting English as well.
			values := url.Values{"token": {token}, "action": {"save"}, "bond": {app.DemoIPCABond}, "quantity": {"2"}, "note": {"<script>Português</script>"}}
			post := func() *httptest.ResponseRecorder {
				r := httptest.NewRequest("POST", "/portfolio", strings.NewReader(values.Encode()))
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				r.AddCookie(&http.Cookie{Name: languageCookie, Value: "pt-BR"})
				response := httptest.NewRecorder()
				h.ServeHTTP(response, r)
				return response
			}
			values.Set("quantity", "0")
			if response := post(); response.Code != 422 || !strings.Contains(response.Body.String(), "Entrada inválida") {
				t.Fatal("form validation was not translated")
			}
			values.Set("quantity", "2")
			if response := post(); response.Code != 303 {
				t.Fatal("Portuguese form did not save")
			}
			english := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/portfolio?lang=en", strings.NewReader(values.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			h.ServeHTTP(english, r)
			if english.Code != 303 {
				t.Fatal("switching language invalidated CSRF token")
			}
			w = httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", "/portfolio?lang=pt-BR", nil))
			if strings.Contains(w.Body.String(), "<script>") || !strings.Contains(w.Body.String(), "&lt;script&gt;Português&lt;/script&gt;") {
				t.Fatal("language processing altered or failed to escape private notes")
			}
		} else {
			for _, path := range []string{"/", "/market", "/playground", "/learn"} {
				w = httptest.NewRecorder()
				h.ServeHTTP(w, httptest.NewRequest("GET", path+"?lang=pt-BR", nil))
				if w.Code != 200 || !strings.Contains(w.Body.String(), `<html lang="pt-BR">`) {
					t.Fatalf("persistent Portuguese empty page: %s", path)
				}
			}
			if quotes, err := s.HasQuotes(context.Background()); err != nil || quotes {
				t.Fatal("language selection imported demo or official data")
			}
		}
		s.Close()
	}
}
