package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/wantedly/webmock-proxy/cache"
)

func main() {
	c, err := cache.New(".webmock")
	if err != nil {
		log.Fatal(err)
	}
	port := flag.Int("port", 8080, "listening port")
	flag.Parse()
	var s *server
	if len(flag.Args()) == 0 || flag.Arg(0) == "replay" {
		s = newReplayServer(c)
	} else if flag.Arg(0) == "record" {
		s = newRecordServer(c)
	} else {
		log.Fatalf("no such command: %s", flag.Arg(0))
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), s); err != nil {
		log.Fatal(err)
	}
}
