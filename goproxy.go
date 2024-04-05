package main

import (
	"errors"
	"github.com/elazarl/goproxy"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type Config struct {
	Address string

	Login    string
	Password string
}

func main() {
	proxy := wrapProxy(goproxy.NewProxyHttpServer())

	err := http.ListenAndServe(":8080", proxy)
	if err != nil && !errors.Is(http.ErrServerClosed, err) {
		log.Fatalf(err.Error())
	}
}

type ProxyWrapper struct {
	proxy  *goproxy.ProxyHttpServer
	logger *zap.Logger
}

func wrapProxy(proxy *goproxy.ProxyHttpServer) *ProxyWrapper {
	wrapper := ProxyWrapper{
		proxy:  proxy,
		logger: nil,
	}

	proxy.Verbose = true

	return &wrapper
}

func (p *ProxyWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
