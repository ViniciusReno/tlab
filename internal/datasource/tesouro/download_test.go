package tesouro

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fixture(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func response(r *http.Request, contentType, body string) *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body)), ContentLength: -1, Request: r}
}

func TestFetchOfficialFixtureAndProvenance(t *testing.T) {
	metadata, csv := fixture(t, "package-show.json"), fixture(t, "official-quotes.csv")
	var requested []string
	client := Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requested = append(requested, r.URL.String())
		if r.Method != "GET" || r.Header.Get("User-Agent") == "" {
			t.Fatal("missing request contract")
		}
		if r.URL.String() == PackageURL {
			return response(r, "application/json; charset=utf-8", metadata), nil
		}
		return response(r, "text/csv", csv), nil
	})}
	got, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(requested) != 2 || got.ResourceID != "796d2059-14e9-44e3-80c9-2d9e30b405c1" || got.SourceURL != requested[1] || len(got.Quotes) != 4 || got.RecordsRead != 10 || got.MaxQuoteDate.Format(time.DateOnly) != "2026-09-04" {
		t.Fatalf("unexpected download: %+v requests=%v", got, requested)
	}
	for _, q := range got.Quotes {
		if q.Source != got.SourceURL || q.ImportedAt != "" {
			t.Fatal("incorrect quote provenance")
		}
	}
}

func TestFetchRejectsInvalidMetadataBeforeCSV(t *testing.T) {
	metadata := fixture(t, "package-show.json")
	for name, body := range map[string]string{
		"invalid JSON":           "<html>unavailable</html>",
		"API failure":            strings.Replace(metadata, `"success": true`, `"success": false`, 1),
		"wrong package":          strings.Replace(metadata, packageID, "wrong", 1),
		"private package":        strings.Replace(metadata, `"private": false`, `"private": true`, 1),
		"missing privacy":        strings.Replace(metadata, `"private": false,`, "", 1),
		"inactive package":       strings.Replace(metadata, `"state": "active"`, `"state": "deleted"`, 1),
		"missing CSV":            strings.ReplaceAll(metadata, `"format": "CSV"`, `"format": "PDF"`),
		"ambiguous CSV":          strings.ReplaceAll(metadata, `"format": "PDF"`, `"format": "CSV"`),
		"wrong resource package": strings.ReplaceAll(metadata, `"package_id": "`+packageID+`"`, `"package_id": "other"`),
		"external URL":           strings.ReplaceAll(metadata, "https://www.tesourotransparente.gov.br/ckan/dataset/", "https://example.com/"),
		"trailing JSON":          metadata + `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			client := Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return response(r, "application/json", body), nil
			})}
			_, err := client.Fetch(context.Background())
			if !errors.Is(err, bond.SchemaChanged) || calls != 1 {
				t.Fatalf("err=%v calls=%d", err, calls)
			}
		})
	}
}

func TestFetchRedirectPolicy(t *testing.T) {
	metadata, csv := fixture(t, "package-show.json"), fixture(t, "official-quotes.csv")
	for _, location := range []string{
		"https://example.com/data", "http://www.tesourotransparente.gov.br/data", "https://127.0.0.1/data",
		"https://www.tesourotransparente.gov.br.evil.example/data", "https://user@www.tesourotransparente.gov.br/data",
		"https://www.tesourotransparente.gov.br:8443/data", "https://www.tesourotransparente.gov.br/data#fragment",
		"/redirected.csv", "/loop",
	} {
		t.Run(location, func(t *testing.T) {
			calls := 0
			client := Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if !officialURL(r.URL.String()) {
					t.Fatalf("unsafe request reached transport: %s", r.URL)
				}
				if r.URL.String() == PackageURL {
					return response(r, "application/json", metadata), nil
				}
				if r.URL.Path == "/redirected.csv" {
					return response(r, "text/csv", csv), nil
				}
				res := response(r, "text/plain", "")
				res.StatusCode = 302
				res.Header.Set("Location", location)
				return res, nil
			})}
			got, err := client.Fetch(context.Background())
			if location == "/redirected.csv" {
				if err != nil || !strings.HasSuffix(got.SourceURL, location) || got.Quotes[0].Source != got.SourceURL {
					t.Fatalf("redirect provenance: %+v %v", got, err)
				}
			} else if !errors.Is(err, bond.SourceUnavailable) || calls > 6 {
				t.Fatalf("redirect accepted or unbounded: %v calls=%d", err, calls)
			}
		})
	}
}

type brokenBody struct{}

func (brokenBody) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (brokenBody) Close() error             { return nil }

func TestDownloadBoundsContentTypeAndReadFailures(t *testing.T) {
	for _, tc := range []struct {
		name, body, contentType string
		size                    int64
		status                  int
		broken                  bool
		want                    error
	}{
		{"at limit", "12345678", "text/csv", -1, 200, false, nil},
		{"stream exceeds limit", "123456789", "text/csv", -1, 200, false, bond.InvalidInput},
		{"declared too large", "x", "text/csv", 9, 200, false, bond.InvalidInput},
		{"HTML", "x", "text/html", -1, 200, false, bond.SchemaChanged},
		{"missing type", "x", "", -1, 200, false, bond.SchemaChanged},
		{"empty", "", "text/csv", -1, 200, false, bond.InvalidInput},
		{"truncated transfer", "", "text/csv", -1, 200, true, bond.SourceUnavailable},
		{"HTTP failure", "x", "text/csv", -1, 503, false, bond.SourceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				res := response(r, tc.contentType, tc.body)
				res.ContentLength, res.StatusCode = tc.size, tc.status
				if tc.broken {
					res.Body = brokenBody{}
				}
				return res, nil
			})}
			_, _, err := download(context.Background(), client, PackageURL, "text/csv", 8)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestFetchCancellationAndUnsupportedOnly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	client := Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	if _, err := client.Fetch(ctx); !errors.Is(err, bond.SourceUnavailable) {
		t.Fatalf("cancellation: %v", err)
	}
	metadata := fixture(t, "package-show.json")
	csv := fixture(t, "official-quotes.csv")
	for _, name := range []string{"Tesouro Prefixado;", "Tesouro IPCA+;", "Tesouro Selic;"} {
		csv = strings.ReplaceAll(csv, name, "Unknown instrument;")
	}
	client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() == PackageURL {
			return response(r, "application/json", metadata), nil
		}
		return response(r, "text/csv", csv), nil
	})
	got, err := client.Fetch(context.Background())
	if !errors.Is(err, bond.MissingQuote) || got.RecordsRead != 10 || got.Unsupported["Unknown instrument"] != 4 {
		t.Fatalf("zero supported rows: %+v %v", got, err)
	}
}

func TestDatasetDuplicateRejectionAndFullDatasetDate(t *testing.T) {
	rows := strings.Split(strings.TrimSpace(fixture(t, "official-quotes.csv")), "\n")
	var prefixado, other string
	for _, row := range rows[1:] {
		if strings.HasPrefix(row, "Tesouro Prefixado;") {
			prefixado = row
		}
		if strings.HasPrefix(row, "Tesouro Selic;") {
			other = row
		}
	}
	if _, err := ParseDataset(strings.NewReader(rows[0]+"\n"+prefixado+"\n"+prefixado), "test"); !errors.Is(err, bond.InvalidInput) {
		t.Fatalf("duplicate accepted: %v", err)
	}
	fields := strings.Split(prefixado, ";")
	fields[2] = "03/01/2012"
	prefixado = strings.Join(fields, ";")
	fields = strings.Split(other, ";")
	fields[1], fields[2] = "01/03/2030", "04/09/2026"
	other = strings.Join(fields, ";")
	got, err := ParseDataset(strings.NewReader(rows[0]+"\n"+prefixado+"\n"+other), "test")
	if err != nil || got.MaxQuoteDate.Format(time.DateOnly) != "2026-09-04" || got.RecordsRead != 2 {
		t.Fatalf("full dataset date: %+v %v", got, err)
	}
}
