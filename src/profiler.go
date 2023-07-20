package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
)

func Profile(port uint16) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", pprof.Profile)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
