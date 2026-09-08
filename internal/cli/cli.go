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
	"strings"
	"time"

	"github.com/ViniciusReno/tlab/internal/app"
	"github.com/ViniciusReno/tlab/internal/bond"
	"github.com/ViniciusReno/tlab/internal/web"
)

func Run(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		return serve(ctx, args, out, errOut, false)
	}
	if strings.HasPrefix(args[0], "--") && args[0] != "--help" {
		return serve(ctx, args, out, errOut, false)
	}
	switch args[0] {
	case "help", "--help", "-h":
		fmt.Fprintln(out, "Tesouro Lab — offline demo and local persistent storage\n\nCommands:\n  [--port 8080] [--data-dir PATH]   Start persistent mode (empty until synchronization is implemented)\n  demo [--port 8080]\n  analyze <bond-id> --yield <percent> [--source demo|synced] [--basis purchase|mark_to_market|early_exit] [--date YYYY-MM-DD] [--amount BRL]\n  version\n\nExample:\n  tesouro-lab analyze prefixado:2015-01-01 --source demo --yield 8.88\n\nDefault data directory: tesouro-lab under the OS user configuration directory.\nOfficial synchronization, synchronized analysis, and portfolios are not implemented yet.")
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
		return serve(ctx, args[1:], out, errOut, true)
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

func serve(ctx context.Context, args []string, out, errOut io.Writer, demo bool) int {
	flags := flag.NewFlagSet("tesouro-lab", flag.ContinueOnError)
	flags.SetOutput(errOut)
	port := flags.Int("port", 8080, "Loopback TCP port (1–65535)")
	var directory string
	if !demo {
		flags.StringVar(&directory, "data-dir", "", "Local data directory (default: OS user configuration directory/tesouro-lab)")
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if flags.NArg() != 0 || *port < 1 || *port > 65535 {
		return failure(errOut, bond.InvalidInput)
	}
	var service *app.Service
	var err error
	if demo {
		service, err = app.OpenDemo(ctx)
	} else {
		if directory == "" {
			directory, err = app.DefaultDataDir()
		}
		if err == nil {
			service, err = app.OpenPersistent(ctx, directory)
		}
	}
	if err != nil {
		fmt.Fprintln(errOut, "Unable to initialize local storage. Check the data directory and file permissions.")
		return 1
	}
	defer service.Close()
	handler, err := web.New(service)
	if err != nil {
		return failure(errOut, err)
	}
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(*port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Fprintln(errOut, "Unable to listen on "+address+". Check whether the port is already in use; use --port 8081.")
		return 1
	}
	server := &http.Server{Handler: handler, ReadTimeout: 5 * time.Second, ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 8192}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	if demo {
		fmt.Fprintf(out, "Tesouro Lab — historical demo data, temporary in-memory database\nOpen http://%s\nPress Ctrl+C to stop. No persistent portfolio is opened.\n", address)
	} else {
		fmt.Fprintf(out, "Tesouro Lab — persistent local storage\nData directory: %s\nOpen http://%s\nPress Ctrl+C to stop. Local data is preserved.\n", directory, address)
	}
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
