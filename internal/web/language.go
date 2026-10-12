package web

import (
	"context"
	"net/http"
	"net/url"
)

const languageCookie = "tlab_language"

type languageContextKey struct{}

type languageLinks struct {
	English, Portuguese string
}

func languageURLs(r *http.Request) languageLinks {
	v := r.URL.Query()
	v.Set("lang", "en")
	en := r.URL.Path + "?" + v.Encode()
	v.Set("lang", "pt-BR")
	return languageLinks{en, r.URL.Path + "?" + v.Encode()}
}

func requestLanguage(r *http.Request) string {
	if locale, ok := r.Context().Value(languageContextKey{}).(string); ok {
		return locale
	}
	if cookie, err := r.Cookie(languageCookie); err == nil && cookie.Value == "pt-BR" {
		return "pt-BR"
	}
	return "en"
}

func translator(locale string) func(string) string {
	return func(text string) string {
		if locale == "pt-BR" {
			if translated, ok := portuguese[text]; ok {
				return translated
			}
		}
		return text
	}
}

func translateRequest(r *http.Request, text string) string {
	return translator(requestLanguage(r))(text)
}

// Language belongs to browser presentation, never to the financial request.
// The two handlers share one CSRF token, so switching languages does not
// invalidate an already open portfolio form.
func languageHandler(handlers map[string]http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/scenario" || len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/static/" {
			handlers["en"].ServeHTTP(w, r)
			return
		}
		locale := requestLanguage(r)
		if len(r.URL.RawQuery) > 4096 {
			http.Error(w, translator(locale)("Request query is too large."), http.StatusRequestURITooLong)
			return
		}
		values, err := url.ParseQuery(r.URL.RawQuery)
		if languages, ok := values["lang"]; err == nil && ok {
			if len(languages) != 1 || (languages[0] != "en" && languages[0] != "pt-BR") {
				http.Error(w, translator(locale)("Choose English or Brazilian Portuguese."), http.StatusBadRequest)
				return
			}
			locale = languages[0]
			http.SetCookie(w, &http.Cookie{Name: languageCookie, Value: locale, Path: "/", MaxAge: 365 * 24 * 60 * 60, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil})
			// Clone before removing only the presentation parameter. Domain parsers
			// still reject missing, repeated, or malformed financial inputs.
			r = r.Clone(r.Context())
			values.Del("lang")
			r.URL.RawQuery = values.Encode()
		}
		r = r.WithContext(context.WithValue(r.Context(), languageContextKey{}, locale))
		w.Header().Set("Content-Language", locale)
		w.Header().Set("Vary", "Cookie")
		w.Header().Set("Cache-Control", "no-store")
		handlers[locale].ServeHTTP(w, r)
	})
}
