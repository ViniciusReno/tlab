package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"testing"

	assets "github.com/ViniciusReno/tlab"
	"github.com/ViniciusReno/tlab/internal/datasource/tesouro"
	"github.com/ViniciusReno/tlab/internal/storage/sqlite"
)

type syncTransport func(*http.Request) (*http.Response, error)

func (f syncTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSyncCLIReportsPersistsAndRetriesOffline(t *testing.T) {
	ctx, dir := context.Background(), t.TempDir()
	metadata, err := os.ReadFile("../datasource/tesouro/testdata/package-show.json")
	if err != nil {
		t.Fatal(err)
	}
	csv, err := os.ReadFile("../datasource/tesouro/testdata/official-quotes.csv")
	if err != nil {
		t.Fatal(err)
	}
	fail := false
	requests := 0
	client := tesouro.Client{Transport: syncTransport(func(r *http.Request) (*http.Response, error) {
		requests++
		body, contentType := string(csv), "text/csv"
		if r.URL.String() == tesouro.PackageURL {
			body, contentType = string(metadata), "application/json"
		}
		res := &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body)), ContentLength: -1, Request: r}
		if fail {
			res.StatusCode = 503
		}
		return res, nil
	})}
	for _, failed := range []bool{false, true, false} {
		fail = failed
		var out, errOut bytes.Buffer
		code := run(ctx, []string{"sync", "--data-dir", dir}, &out, &errOut, client)
		if (code != 0) != failed {
			t.Fatalf("exit code=%d stderr=%s", code, &errOut)
		}
		var report sqlite.SyncRun
		if err := json.Unmarshal(out.Bytes(), &report); err != nil {
			t.Fatalf("invalid report %q: %v", &out, err)
		}
		if failed {
			if report.Status != "failed" || report.ErrorMessage == "" || !strings.Contains(errOut.String(), "sync_failed") {
				t.Fatalf("failure output: %+v %s", report, &errOut)
			}
		} else if report.Status != "success" || report.RecordsWritten != 2 || report.RecordsRead != 10 || report.DatasetMaxQuoteDate != "2026-09-04" || errOut.Len() != 0 {
			t.Fatalf("success output: %+v %s", report, &errOut)
		}
		s, err := sqlite.OpenPersistent(ctx, dir, assets.Files)
		if err != nil {
			t.Fatal(err)
		}
		stored, err := s.SyncRun(ctx, report.ID)
		if err != nil || stored.Status != report.Status {
			t.Fatalf("report not persisted: %+v %v", stored, err)
		}
		q, err := s.Quote(ctx, "prefixado:2032-01-01", "2026-09-04")
		if err != nil || math.Abs(*q.BuyPU-493.41) > 1e-9 {
			t.Fatalf("quote not persisted: %+v %v", q, err)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
	before := requests
	for _, args := range [][]string{{"sync", "--help"}, {"sync", "--source", "demo"}, {"sync", "--url", "https://example.com"}, {"sync", "extra"}, {"demo", "sync"}} {
		var out, errOut bytes.Buffer
		run(ctx, args, &out, &errOut, client)
	}
	if requests != before {
		t.Fatal("help or invalid flags accessed network")
	}
}
