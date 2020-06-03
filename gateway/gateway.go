package gateway

import (
	"github.com/go-errors/errors"
	"github.com/vavas/go_1month_challenge/config"
	"github.com/vavas/go_1month_challenge/proxy"
	"github.com/vavas/go_1month_challenge/server"
	"go.uber.org/zap"

	"log"
)

type Gateway struct {
	config *config.Config
	server *server.Server
	proxy  *proxy.Proxy
	logger *zap.Logger
}

// New creates a new Gateway instance
func New(config *config.Config, logger *zap.Logger) (*Gateway, error) {

	gtw := &Gateway{
		config: config,
		logger: logger,
	}

	proxy, err := proxy.New(config)
	if err != nil {
		return gtw, errors.Errorf("proxy.New")
	}
	gtw.proxy = proxy

	srv, err := server.New(config, proxy, logger)
	if err != nil {
		return gtw, errors.Errorf("server.New")
	}

	log.Println("Server created")
	gtw.server = srv

	return gtw, nil
}

func (g *Gateway) Start() {
	g.server.Listen()
}

func (g *Gateway) Stop() error {
	return g.server.Close()
}
