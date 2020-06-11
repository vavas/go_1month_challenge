package server

import (
	"context"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vavas/go_mc_gateway/config"
	"github.com/vavas/go_mc_gateway/middlewares/cors"
	"github.com/vavas/go_mc_gateway/middlewares/request_logger"
	"github.com/vavas/go_mc_gateway/proxy"
)

type Server struct {
	config     *config.Config
	proxy      *proxy.Proxy
	gin        *gin.Engine
	httpServer *http.Server
	logger     *zap.Logger
	tlsEnabled bool
}

// New creates an instance of Server. It requires a config, logger and proxy
// are given. These will be used by the server to determine configuration,
// log operational status, and dispatch requests to Nats services.
func New(config *config.Config, proxy *proxy.Proxy, logger *zap.Logger) (*Server, error) {
	server := &Server{
		config: config,
		gin:    gin.New(),
		proxy:  proxy,
		logger: logger,
	}

	if err := server.setupHTTPServer(); err != nil {
		return server, err
	}

	if err := server.setupGinMiddleware(); err != nil {
		return server, err
	}
	if err := server.setupGinRoutes(); err != nil {
		return server, err
	}

	return server, nil
}

func (s *Server) Listen() {

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			log.Fatal("s.httpServer.ListenAndServe")
		}
	}()
}

// Close prevents the HTTP server from accepting new incoming connections and
// attempts to allow existing connections a window of time to complete transit.
// Returns once all connections have been closed.
func (s *Server) Close() error {
	return s.httpServer.Shutdown(context.Background())
}

func (s *Server) setupHTTPServer() error {

	s.httpServer = &http.Server{
		Addr:              s.config.Server.Listen,
		Handler:           s.gin,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	s.logger.Debug("HTTP server created")

	return nil
}

// setupGinMiddleware is used by New to attach all gin middleware required by
// the server's request pipeline.
func (s *Server) setupGinMiddleware() error {
	s.logger.Debug("Setting up gin middleware")

	s.gin.Use(cors.Cors)
	s.gin.Use(requestlogger.RequestLogger(s.logger))
	s.gin.Use(func(c *gin.Context) { c.Set("config", s.config) })
	s.logger.Debug("Gin middleware set up")

	return nil
}

func (s *Server) setupGinRoutes() error {
	s.logger.Debug("Setting up gin routes")
	s.gin.Use(s.proxy.GetHandler)

	s.gin.Use(s.proxy.HandleRequest)

	s.logger.Debug("Gin routes set up")

	return nil
}
