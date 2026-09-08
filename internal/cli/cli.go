// Package cli is a thin delivery layer over the shared application workflow.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/web"
)

func Run(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, "Persistent mode is planned for M2. Run: tesouro-lab demo")
		return 1
	}
	switch args[0] {
	case "help", "--help", "-h":
		fmt.Fprintln(out, "Tesouro Lab — M1 offline Prefixado demo\n\nCommands:\n  demo [--port 8080]\n  analyze <bond-id> --yield <percent> [--source demo|synced] [--basis purchase|mark_to_market|early_exit] [--date YYYY-MM-DD] [--amount BRL]\n  version\n\nExample:\n  tesouro-lab analyze prefixado:2015-01-01 --source demo --yield 8.88\n\nLive synchronization and persistent portfolios are not implemented in M1.")
		return 0
	case "version":
		if len(args) != 1 {
			return failure(errOut, bond.InvalidInput)
		}
		fmt.Fprintln(out, app.BuildVersion)
		return 0
	case "sync":
		return failure(errOut, bond.SourceUnavailable)
	case "analyze":
		return analyze(ctx, args[1:], out, errOut)
	case "demo":
		return demo(ctx, args[1:], out, errOut)
	default:
		return failure(errOut, bond.Unsupported)
	}
}

func failure(w io.Writer, err error) int {
	code := "internal_error"
	if e, ok := err.(bond.Error); ok {
		code = string(e)
	}
	fmt.Fprintf(w, "%s: %s\n", code, bond.Message(err))
	return 1
}

func analyze(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		return failure(errOut, bond.InvalidInput)
	}
	id := args[0]
	flags := flag.NewFlagSet("analyze", flag.ContinueOnError)
	flags.SetOutput(errOut)
	values := url.Values{"bond": {id}}
	pointers := map[string]*string{}
	for _, key := range []string{"source", "basis", "date", "yield", "amount"} {
		pointers[key] = flags.String(key, "", "Scenario "+key)
	}
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if flags.NArg() != 0 {
		return failure(errOut, bond.InvalidInput)
	}
	flags.Visit(func(f *flag.Flag) { values.Set(f.Name, *pointers[f.Name]) })
	request, err := app.ParseRequest(values, "synced")
	if err != nil {
		return failure(errOut, err)
	}
	if request.Source != "demo" {
		return failure(errOut, bond.SourceUnavailable)
	}
	service, err := app.OpenDemo(ctx)
	if err != nil {
		return failure(errOut, err)
	}
	defer service.Close()
	result, err := service.Analyze(ctx, request)
	if err != nil {
		return failure(errOut, err)
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return failure(errOut, err)
	}
	return 0
}

func demo(ctx context.Context, args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("demo", flag.ContinueOnError)
	flags.SetOutput(errOut)
	port := flags.Int("port", 8080, "Loopback TCP port (1–65535)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if flags.NArg() != 0 || *port < 1 || *port > 65535 {
		return failure(errOut, bond.InvalidInput)
	}
	service, err := app.OpenDemo(ctx)
	if err != nil {
		return failure(errOut, err)
	}
	defer service.Close()
	handler, err := web.New(service)
	if err != nil {
		return failure(errOut, err)
	}
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(*port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Fprintln(errOut, "Unable to listen on "+address+". Check whether the port is already in use; try demo --port 8081.")
		return 1
	}
	server := &http.Server{Handler: handler, ReadTimeout: 5 * time.Second, ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 8192}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	fmt.Fprintf(out, "Tesouro Lab — historical demo data, temporary in-memory database\nOpen http://%s\nPress Ctrl+C to stop. No persistent portfolio is opened.\n", address)
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return failure(errOut, err)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			server.Close()
			return failure(errOut, err)
		}
		if err := <-done; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return failure(errOut, err)
		}
	}
	return 0
}
