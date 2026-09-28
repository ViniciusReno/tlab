package web

import (
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/bond"
)

type marketPage struct {
	Market                    app.Market
	History                   app.History
	IsHistory, IncludeMatured bool
	BondID, Error             string
}

func renderMarket(w http.ResponseWriter, r *http.Request, service *app.Service, tmpl *template.Template, now time.Time) {
	if len(r.URL.RawQuery) > 4096 {
		http.Error(w, "Request query is too large.", http.StatusRequestURITooLong)
		return
	}
	p := marketPage{IsHistory: r.URL.Path == "/history"}
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		err = bond.InvalidInput
	}
	for key, v := range values {
		if len(v) != 1 || (p.IsHistory && key != "bond" && key != "before") || (!p.IsHistory && key != "matured") {
			err = bond.InvalidInput
		}
	}
	if err == nil && !p.IsHistory {
		if v := values.Get("matured"); v != "" && v != "1" {
			err = bond.InvalidInput
		}
	}
	if err == nil {
		if p.IsHistory {
			p.BondID = values.Get("bond")
			p.History, err = service.History(r.Context(), p.BondID, values.Get("before"))
		} else {
			p.IncludeMatured = values.Get("matured") == "1"
			p.Market, err = service.Market(r.Context(), now, p.IncludeMatured)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		p.Error = "Unable to read local data. Restart the application and check the database file."
		status := http.StatusInternalServerError
		if _, ok := err.(bond.Error); ok {
			p.Error, status = bond.Message(err), http.StatusUnprocessableEntity
		}
		w.WriteHeader(status)
	}
	tmpl.ExecuteTemplate(w, "market.html", p)
}
