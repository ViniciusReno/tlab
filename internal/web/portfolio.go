package web

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"html/template"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/portfolio"
)

type portfolioPage struct {
	Data         app.Portfolio
	Demo         bool
	Token, Error string
	Values       url.Values
}

func portfolioHandler(service *app.Service, tmpl *template.Template, now func() time.Time) (http.HandlerFunc, error) {
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(random[:])
	return func(w http.ResponseWriter, r *http.Request) {
		p := portfolioPage{Demo: service.Source() == "demo", Token: token, Values: make(url.Values)}
		status := http.StatusOK
		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
			media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || media != "application/x-www-form-urlencoded" {
				http.Error(w, "Use a form submission.", http.StatusUnsupportedMediaType)
				return
			}
			if err = r.ParseForm(); err != nil {
				http.Error(w, "Invalid or oversized form.", http.StatusBadRequest)
				return
			}
			origin := r.Header.Get("Origin")
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			if subtle.ConstantTimeCompare([]byte(r.PostForm.Get("token")), []byte(token)) != 1 || (origin != "" && origin != scheme+"://"+r.Host) || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				http.Error(w, "This form is no longer valid. Reload the portfolio and try again.", http.StatusForbidden)
				return
			}
			p.Values = r.PostForm
			switch r.PostForm.Get("action") {
			case "save":
				var position portfolio.Position
				position, err = app.ParsePosition(r.PostForm)
				if err == nil {
					err = service.SavePosition(r.Context(), position)
				}
			case "remove":
				if r.PostForm.Get("confirm") != "yes" {
					err = bond.InvalidInput
				} else {
					err = service.DeletePosition(r.Context(), r.PostForm.Get("id"))
				}
			default:
				err = bond.InvalidInput
			}
			if err == nil {
				http.Redirect(w, r, "/portfolio", http.StatusSeeOther)
				return
			}
			status = http.StatusUnprocessableEntity
			p.Error = bond.Message(err)
			if err == portfolio.InconsistentCost {
				p.Error = "Acquisition inputs disagree. Choose the authoritative quantity, purchase unit price, or gross acquisition amount; clear the conflicting optional field and save again."
			}
		}
		var err error
		p.Data, err = service.Portfolio(r.Context(), now())
		if err != nil {
			http.Error(w, "Unable to read the local portfolio.", http.StatusInternalServerError)
			return
		}
		if r.Method == http.MethodGet {
			if len(r.URL.RawQuery) > 4096 {
				http.Error(w, "Query is too large.", http.StatusRequestURITooLong)
				return
			}
			v, err := url.ParseQuery(r.URL.RawQuery)
			if err != nil {
				http.Error(w, "Invalid query.", http.StatusBadRequest)
				return
			}
			if id := v.Get("edit"); id != "" {
				found := false
				for _, position := range p.Data.Positions {
					if position.ID == id {
						found = true
						p.Values = positionValues(position.Position)
					}
				}
				if !found {
					http.NotFound(w, r)
					return
				}
			}
			if id := v.Get("scenario"); id != "" {
				// Reuse the strict percentage/date parser and application scenario service.
				var selected *app.PositionView
				for i := range p.Data.Positions {
					if p.Data.Positions[i].ID == id {
						selected = &p.Data.Positions[i]
					}
				}
				if selected == nil {
					http.NotFound(w, r)
					return
				}
				request, parseErr := app.ParseRequest(url.Values{"bond": {selected.BondID}, "source": {service.Source()}, "basis": {"mark_to_market"}, "date": {v.Get("date")}, "yield": {v.Get("yield")}}, service.Source())
				if v.Get("date") == "" {
					parseErr = bond.InvalidInput
				}
				if parseErr == nil {
					result, amount, e := service.PositionScenario(r.Context(), id, request.Date, request.Yield)
					parseErr = e
					if e == nil {
						selected.Scenario = &result
						selected.ScenarioValue = &amount
					}
				}
				if parseErr != nil {
					status = http.StatusUnprocessableEntity
					p.Error = bond.Message(parseErr)
				}
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		tmpl.ExecuteTemplate(w, "portfolio.html", p)
	}, nil
}

func positionValues(p portfolio.Position) url.Values {
	v := url.Values{"id": {p.ID}, "bond": {p.BondID}, "quantity": {app.Decimal(p.Quantity)}, "note": {p.Note}}
	if p.PurchaseDate != nil {
		v.Set("purchase_date", p.PurchaseDate.Format(time.DateOnly))
	}
	if p.PurchasePU != nil {
		v.Set("purchase_pu", app.Decimal(*p.PurchasePU))
	}
	if p.PurchaseYield != nil {
		v.Set("purchase_yield", app.Decimal(*p.PurchaseYield*100))
	}
	if p.InvestedBRL != nil {
		v.Set("invested_brl", app.Decimal(*p.InvestedBRL))
	}
	return v
}
