package app

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
)

type syncTransport func(*http.Request) (*http.Response, error)

func (f syncTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSyncWorkflowFailurePreservationAndDemoIsolation(t *testing.T) {
	ctx := context.Background()
	metadata, err := os.ReadFile("../datasource/tesouro/testdata/package-show.json")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("../datasource/tesouro/testdata/official-quotes.csv")
	if err != nil {
		t.Fatal(err)
	}
	csv, mode, requests := string(fixture), "success", 0
	client := tesouro.Client{Transport: syncTransport(func(r *http.Request) (*http.Response, error) {
		requests++
		if mode == "network" {
			return nil, errors.New("network failure")
		}
		if mode == "cancel" {
			<-r.Context().Done()
			return nil, r.Context().Err()
		}
		body, contentType := csv, "text/csv"
		if r.URL.String() == tesouro.PackageURL {
			body, contentType = string(metadata), "application/json"
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body)), ContentLength: -1, Request: r}, nil
	})}
	s, err := OpenPersistent(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	first, err := s.Sync(ctx, client)
	if err != nil || first.Status != "success" || first.RecordsRead != 10 || first.RecordsWritten != 2 || first.DatasetMaxQuoteDate != "2026-09-04" || len(first.Unsupported) != 7 {
		t.Fatalf("sync: %+v %v", first, err)
	}
	if _, err := time.Parse(time.RFC3339Nano, first.ImportedAt); err != nil {
		t.Fatal(err)
	}
	quote, err := s.store.Quote(ctx, "prefixado:2032-01-01", "2026-09-04")
	if err != nil || math.Abs(*quote.BuyPU-493.41) > 1e-9 || math.Abs(*quote.BasePU-490.42) > 1e-9 || quote.ImportedAt != first.ImportedAt {
		t.Fatalf("quote: %+v %v", quote, err)
	}
	for _, failureMode := range []string{"network", "schema", "malformed", "unsupported", "cancel"} {
		t.Run(failureMode, func(t *testing.T) {
			mode, csv = failureMode, string(fixture)
			switch mode {
			case "schema":
				csv = strings.Replace(csv, "PU Base Manha", "Unknown field", 1)
			case "malformed":
				csv += "malformed;row\n"
			case "unsupported":
				csv = strings.ReplaceAll(csv, "Tesouro Prefixado;", "Unknown instrument;")
			}
			requestCtx := ctx
			if mode == "cancel" {
				var cancel context.CancelFunc
				requestCtx, cancel = context.WithTimeout(ctx, 25*time.Millisecond)
				defer cancel()
			}
			report, err := s.Sync(requestCtx, client)
			if !errors.Is(err, bond.SyncFailed) || report.Status != "failed" || report.RecordsWritten != 0 || report.ImportedAt != "" || report.ErrorMessage == "" {
				t.Fatalf("failure report: %+v %v", report, err)
			}
			stored, err := s.store.SyncRun(ctx, report.ID)
			if err != nil || stored.Status != "failed" || stored.ErrorMessage != report.ErrorMessage {
				t.Fatalf("failure not recorded: %+v %v", stored, err)
			}
			q, err := s.store.Quote(ctx, quote.Bond.ID, "2026-09-04")
			if err != nil || math.Abs(*q.BuyPU-*quote.BuyPU) > 1e-9 || q.ImportedAt != first.ImportedAt {
				t.Fatalf("failure replaced valid data: %+v %v", q, err)
			}
		})
	}
	mode, csv = "success", string(fixture)
	if report, err := s.Sync(ctx, client); err != nil || report.RecordsWritten != 2 {
		t.Fatalf("retry failed: %+v %v", report, err)
	}
	demo, err := OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer demo.Close()
	before := requests
	if _, err := demo.Sync(ctx, client); !errors.Is(err, bond.SourceMismatch) {
		t.Fatalf("demo sync accepted: %v", err)
	}
	if requests != before {
		t.Fatal("demo accessed network")
	}
	q, err := demo.store.Quote(ctx, DemoBond, "2012-01-03")
	if err != nil || math.Abs(*q.BuyPU-733.86) > 1e-9 || q.ImportedAt != "" {
		t.Fatalf("demo changed: %+v %v", q, err)
	}
}
