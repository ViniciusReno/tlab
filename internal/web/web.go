// Package web presents the shared application result using embedded HTML and assets.
package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/bond"
)

type page struct {
	Result            app.Result
	Points            []app.Point
	Curve             string
	Error             string
	Values            url.Values
	Persistent        bool
	HasQuotes         bool
	SelectionRequired bool
	ShockError        string
}

func money(n float64) string {
	parts := strings.Split(fmt.Sprintf("%.2f", n), ".")
	integer := parts[0]
	for i := len(integer) - 3; i > 0; i -= 3 {
		integer = integer[:i] + "," + integer[i:]
	}
	return "BRL " + integer + "." + parts[1]
}

func New(service *app.Service) (http.Handler, error) {
	return newHandler(service, time.Now)
}

func newHandler(service *app.Service, now func() time.Time) (http.Handler, error) {
	tmpl, err := template.New("page.html").Funcs(template.FuncMap{
		"money":   money,
		"percent": func(n float64) string { return fmt.Sprintf("%.2f%%", n*100) },
		"change":  func(n float64) string { return fmt.Sprintf("%+.2f%%", n*100) },
		"decimal": app.Decimal,
		"shockURL": func(r app.Result, yield float64) string {
			v := r.Values()
			v.Set("yield", app.Decimal(yield*100))
			return "/playground?" + v.Encode()
		},
		"yield": func(n float64) string { return app.Decimal(n * 100) },
		"value": func(v url.Values, key string) string { return v.Get(key) },
		"date": func(t time.Time) string {
			if t.IsZero() {
				return "Unavailable"
			}
			return t.Format(time.DateOnly)
		},
		"quoteMoney": func(n *float64) string {
			if n == nil {
				return "Unavailable"
			}
			return money(*n)
		},
		"quoteYield": func(n *float64) string {
			if n == nil {
				return "Unavailable"
			}
			return fmt.Sprintf("%.2f%%", *n*100)
		},
		"historyURL": func(id, before string) string {
			return "/history?" + url.Values{"bond": {id}, "before": {before}}.Encode()
		},
		"scenarioURL": func(q bond.Quote, basis string) string {
			pu, yield := q.BuyPU, q.BuyYield
			if basis == "mark_to_market" {
				pu, yield = q.BasePU, q.SellYield
			}
			if pu == nil || yield == nil || q.Date.IsZero() {
				return ""
			}
			v := url.Values{"bond": {q.Bond.ID}, "source": {"synced"}, "basis": {basis}, "date": {q.Date.Format(time.DateOnly)}, "yield": {app.Decimal(*yield * 100)}}
			return "/playground?" + v.Encode()
		},
	}).ParseFS(assets.Files, "web/templates/*.html")
	if err != nil {
		return nil, err
	}
	static, err := fs.Sub(assets.Files, "web/static")
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	render := func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.RawQuery) > 4096 {
			http.Error(w, "Request query is too large.", http.StatusRequestURITooLong)
			return
		}
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err == nil && service.Source() == "synced" && len(values) == 0 && r.URL.Path != "/api/scenario" {
			hasQuotes, readErr := service.HasQuotes(r.Context())
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			p := page{Persistent: true, HasQuotes: hasQuotes, SelectionRequired: true}
			if readErr != nil {
				p.Error = "Unable to read local data. Restart the application and check the database file."
				w.WriteHeader(http.StatusInternalServerError)
			}
			tmpl.ExecuteTemplate(w, "page.html", p)
			return
		}
		var request app.Request
		if err == nil && len(values) == 0 && r.URL.Path != "/api/scenario" {
			request, err = service.DefaultRequest(r.Context())
		} else if err == nil {
			if values.Get("source") != "" && values.Get("source") != service.Source() {
				err = bond.SourceMismatch
			} else {
				request, err = app.ParseRequest(values, service.Source())
			}
		}
		var result app.Result
		if err == nil {
			result, err = service.Analyze(r.Context(), request)
		}
		if r.URL.Path == "/api/scenario" {
			w.Header().Set("Content-Type", "application/json")
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				code := "internal_error"
				if e, ok := err.(bond.Error); ok {
					code = string(e)
				}
				json.NewEncoder(w).Encode(map[string]string{"error": code, "message": bond.Message(err)})
				return
			}
			json.NewEncoder(w).Encode(result)
			return
		}
		p := page{Result: result, Values: values, Persistent: service.Source() == "synced"}
		if err == nil {
			p.Values = result.Values()
			var shockErr error
			p.Points, shockErr = app.Shocks(result)
			if shockErr != nil {
				p.ShockError = "The default shock range is unavailable for this baseline. The selected scenario above remains valid."
			} else {
				// SVG coordinates only; financial values are calculated in app/pricing.
				min, max := p.Points[len(p.Points)-1].PU, p.Points[0].PU
				for i, point := range p.Points {
					x := 40 + float64(i)*60
					y := 100.0
					if max > min {
						y = 20 + (max-point.PU)/(max-min)*160
					}
					p.Curve += fmt.Sprintf("%.2f,%.2f ", x, y)
				}
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err != nil {
			p.Error = bond.Message(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
		// The template is parsed once at startup and has no user-supplied HTML.
		if err := tmpl.ExecuteTemplate(w, "page.html", p); err != nil {
			return
		}
	}
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		if service.Source() == "synced" {
			renderMarket(w, r, service, tmpl, now())
			return
		}
		render(w, r)
	})
	mux.HandleFunc("GET /market", func(w http.ResponseWriter, r *http.Request) { renderMarket(w, r, service, tmpl, now()) })
	mux.HandleFunc("GET /history", func(w http.ResponseWriter, r *http.Request) { renderMarket(w, r, service, tmpl, now()) })
	mux.HandleFunc("GET /playground", render)
	mux.HandleFunc("GET /api/scenario", render)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	}), nil
}
