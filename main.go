package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

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

	c, err := store.New(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	var s *proxy.Server
	if *record {
		s = proxy.NewRecordServer(c)
	} else {
		s = proxy.NewReplayServer(c)
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), s); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
