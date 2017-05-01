package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wantedly/webmock-proxy/webmock/proxy"
	"github.com/wantedly/webmock-proxy/webmock/store"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	f := flag.NewFlagSet("webmock-proxy", flag.ContinueOnError)

	dir := f.String("dir", ".webmock", "cache directory")
	record := f.Bool("record", false, "record http/https exchanges")
	port := f.Int("port", 8080, "listening port")

	if err := f.Parse(args[1:]); err != nil {
		return 1
	}

	if err := os.MkdirAll(*dir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var exs store.Exchanges
	yamlpath := filepath.Join(*dir, "default.yaml")
	if _, err := os.Stat(yamlpath); err == nil {
		y, err := os.Open(yamlpath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		exs, err = store.LoadExchanges(y)
		y.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	out, err := os.Create(yamlpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	store := store.New(out, exs)

	var s *proxy.Server
	if *record {
		s = proxy.NewRecordServer(store)
	} else {
		s = proxy.NewReplayServer(store)
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), s); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
