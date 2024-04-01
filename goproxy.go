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

	err := http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", proxy)
	if err != nil && !errors.Is(http.ErrServerClosed, err) {
		log.Fatalf(err.Error())
	}
}
