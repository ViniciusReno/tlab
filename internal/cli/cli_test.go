package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/cli"
	"github.com/ViniciusReno/tlab/internal/web"
)

func TestCLIAndHTTPParity(t *testing.T) {
	ctx := context.Background()
	var out, errOut bytes.Buffer
	args := []string{"analyze", app.DemoBond, "--source", "demo", "--basis", "purchase", "--date", "2012-01-03", "--yield", "8.88", "--amount", "10000"}
	if code := cli.Run(ctx, args, &out, &errOut); code != 0 {
		t.Fatalf("%d %s", code, errOut.String())
	}
	var fromCLI app.Result
	if err := json.Unmarshal(out.Bytes(), &fromCLI); err != nil {
		t.Fatal(err)
	}
	s, err := app.OpenDemo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h, err := web.New(s)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	h.ServeHTTP(response, httptest.NewRequest("GET", "/api/scenario?"+fromCLI.Values().Encode(), nil))
	if response.Code != 200 {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
	var fromHTTP app.Result
	if err := json.Unmarshal(response.Body.Bytes(), &fromHTTP); err != nil {
		t.Fatal(err)
	}
	if math.Abs(fromCLI.ScenarioPU-fromHTTP.ScenarioPU) > 1e-9 || fromCLI.QuoteDate != fromHTTP.QuoteDate || fromCLI.Basis != fromHTTP.Basis || fromCLI.BusinessDays != fromHTTP.BusinessDays {
		t.Fatal("CLI/API mismatch")
	}
	page := httptest.NewRecorder()
	h.ServeHTTP(page, httptest.NewRequest("GET", "/playground?"+fromCLI.Values().Encode(), nil))
	for _, text := range []string{"BRL 774.99", "BRL 10,560.49", "2012-01-03", "755", "--yield 8.88", "Historical demo", "Advanced", `id="yield" name="yield" type="text" inputmode="decimal"`, `id="amount" name="amount" type="text" inputmode="decimal"`} {
		if !strings.Contains(page.Body.String(), text) {
			t.Fatalf("missing %q in %s", text, page.Body.String())
		}
	}
}

type startupWriter struct {
	bytes.Buffer
	cancel context.CancelFunc
}

func (w *startupWriter) Write(p []byte) (int, error) {
	n, err := w.Buffer.Write(p)
	if strings.Contains(w.Buffer.String(), "Press Ctrl+C") {
		w.cancel()
	}
	return n, err
}

func TestPersistentCLIStartupAndShutdown(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 2; i++ {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		_, port, err := net.SplitHostPort(listener.Addr().String())
		listener.Close()
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		out := &startupWriter{cancel: cancel}
		var errOut bytes.Buffer
		code := cli.Run(ctx, []string{"--data-dir", dir, "--port", port}, out, &errOut)
		cancel()
		if code != 0 || !strings.Contains(out.String(), "http://127.0.0.1:"+port) || !strings.Contains(out.String(), "Local data is preserved") {
			t.Fatalf("startup/shutdown: %d %s %s", code, out.String(), errOut.String())
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "tesouro-lab.db")); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := cli.Run(context.Background(), []string{"--data-dir", filepath.Join(dir, "tesouro-lab.db")}, &out, &errOut); code == 0 || !strings.Contains(errOut.String(), "Unable to initialize local storage") {
		t.Fatalf("missing storage error: %d %s", code, errOut.String())
	}
}

func TestHTTPErrorsAndAssets(t *testing.T) {
	s, err := app.OpenDemo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h, err := web.New(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, target string
		status         int
	}{
		{"GET", "/", 200}, {"GET", "/static/style.css", 200}, {"GET", "/static/app.js", 200},
		{"POST", "/playground", 405}, {"GET", "/sync", 404}, {"GET", "/static/../../go.mod", 404},
		{"GET", "/api/scenario?bond=" + app.DemoBond + "&yield=12&source=synced", 422},
		{"GET", "/api/scenario?bond=" + app.DemoBond + "&yield=NaN", 422},
		{"GET", "/api/scenario?bond=" + app.DemoBond + "&yield=12&basis=early_exit", 422},
		{"GET", "/api/scenario?bond=ipca:2015-01-01&yield=12", 422},
		{"GET", "/playground?bad=%ZZ", 422},
		{"GET", "/playground?" + strings.Repeat("a", 4097), 414},
	} {
		recorder := httptest.NewRecorder()
		h.ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.target, nil))
		// ServeMux canonicalizes traversal paths before routing.
		if tc.target == "/static/../../go.mod" && recorder.Code == http.StatusMovedPermanently {
			continue
		}
		if recorder.Code != tc.status {
			t.Fatalf("%s: %d want %d", tc.target, recorder.Code, tc.status)
		}
		if recorder.Header().Get("Content-Security-Policy") == "" {
			t.Fatal("missing CSP")
		}
		if strings.Contains(recorder.Body.String(), "SQLITE") {
			t.Fatal("internal storage error leaked")
		}
	}
}

func TestCLIRejectsUnavailableAndInvalidCommands(t *testing.T) {
	for _, args := range [][]string{
		{"sync"}, {"unknown"}, {"analyze", app.DemoBond, "--yield", "8.88"},
		{"analyze", app.DemoBond, "--source", "demo", "--yield", "NaN"},
		{"analyze", app.DemoBond, "--source", "demo", "--yield", "8.88", "extra"},
		{"demo", "--port", "0"}, {"demo", "--port", "65536"},
		{"--port", "0"}, {"demo", "--data-dir", "unused"},
	} {
		var out, errOut bytes.Buffer
		if cli.Run(context.Background(), args, &out, &errOut) == 0 {
			t.Fatalf("accepted %v", args)
		}
		if errOut.Len() == 0 {
			t.Fatalf("no explanation for %v", args)
		}
	}
}
