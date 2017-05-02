package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/wantedly/webmock-proxy/webmock"
)

func main() {
	err, status := run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	os.Exit(status)
}

func run(args []string) (error, int) {
	f := flag.NewFlagSet("webmock-proxy", flag.ContinueOnError)

	dir := f.String("dir", ".webmock", "cache directory")
	record := f.Bool("record", false, "record http/https exchanges")
	addr := f.String("addr", ":8080", "listening address")

	if err := f.Parse(args[1:]); err != nil {
		return err, 1
	}

	var options []webmock.Option
	options = append(options, webmock.BaseDir(*dir))
	if *record {
		options = append(options, webmock.RecordMode)
	}
	options = append(options, webmock.Addr(*addr))

	px, err := webmock.NewProxy(options...)
	if err != nil {
		return err, 1
	}
	defer px.Close()
	fmt.Printf("Listening on %s\n", px.URL.String())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	return nil, 0
}
