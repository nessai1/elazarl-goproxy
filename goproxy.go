package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elazarl/goproxy"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
)

type Config struct {
	Address string `json:"address,omitempty"`

	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`

	IPWhitelist []string `json:"ip_whitelist,omitempty"`
}

func main() {
	config, err := fetchConfig()
	if err != nil {
		log.Fatalf("Cannot fetch config: %s", err.Error())
	}

	proxy := wrapProxy(goproxy.NewProxyHttpServer(), config)

	err = http.ListenAndServe(config.Address, proxy)
	if err != nil && !errors.Is(http.ErrServerClosed, err) {
		log.Fatalf(err.Error())
	}
}

func fetchConfig() (Config, error) {
	f, err := os.Open("config.json")
	if err != nil {
		return Config{}, fmt.Errorf("cannot open config file: %w", err)
	}

	b := bytes.Buffer{}
	n, err := b.ReadFrom(f)
	if err != nil {
		return Config{}, fmt.Errorf("cannot read file: %w", err)
	}

	if n == 0 {
		return Config{}, fmt.Errorf("empty file")
	}

	config := Config{}
	err = json.Unmarshal(b.Bytes(), &config)
	if err != nil {
		return Config{}, fmt.Errorf("cannot unmarshal file: %w", err)
	}

	return config, nil
}

func createLogger(level zapcore.Level) *zap.Logger {
	atom := zap.NewAtomicLevel()

	atom.SetLevel(level)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	return zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))
}

type ProxyWrapper struct {
	proxy  *goproxy.ProxyHttpServer
	logger *zap.Logger
	config Config
}

func wrapProxy(proxy *goproxy.ProxyHttpServer, config Config) *ProxyWrapper {
	wrapper := ProxyWrapper{
		proxy:  proxy,
		logger: createLogger(zapcore.DebugLevel),
		config: config,
	}

	proxy.Verbose = true

	return &wrapper
}

func (p *ProxyWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientIP := strings.Split(r.RemoteAddr, ":")[0]
	p.logger.Info("Got HTTP request", zap.String("method", r.Method), zap.String("url", r.URL.String()), zap.String("ip", clientIP), zap.String("remote_addr", r.RemoteAddr))

	if p.config.IPWhitelist != nil {

		if !slices.Contains(p.config.IPWhitelist, clientIP) {
			p.logger.Info("Client doesn't contains in IP whitelist", zap.String("ip", clientIP))
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	p.logger.Info("Client was verified", zap.String("ip", clientIP))
	p.proxy.ServeHTTP(w, r)
}
