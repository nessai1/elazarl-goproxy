package main

import (
	"errors"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	err := http.ListenAndServe(":8080", proxy)
	if err != nil && errors.Is(http.ErrServerClosed, err) {
		log.Fatalf(err.Error())
	}
}
