package http

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/milosgajdos/orbnet/pkg/graph/api"
	_ "github.com/milosgajdos/orbnet/pkg/graph/api/http/docs" // blank import for swagger docs
)

// Server is an HTTP server used to provide REST API
// access for various Graph API endpoints.
type Server struct {
	// app is fiber app.
	app *fiber.App
	// ln is a network listener.
	ln net.Listener
	// Addr is bind address
	Addr string
	// GraphService provides access to Graph endpoints.
	GraphService api.GraphService
	// NodeService provides access to Node endpoints.
	NodeService api.NodeService
	// EdgeService provides access to Edge endpoints.
	EdgeService api.EdgeService
}

type Config struct {
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewServer creates a new Server and returns it.
func NewServer(config ...Config) (*Server, error) {
	var c fiber.Config

	if len(config) > 0 {
		c.IdleTimeout = config[0].IdleTimeout
		c.ReadTimeout = config[0].ReadTimeout
		c.WriteTimeout = config[0].WriteTimeout
	}

	s := &Server{
		app: fiber.New(c),
	}

	s.app.Use(recover.New())
	s.app.Use(logger.New())

	api := s.app.Group("/api")
	v1 := api.Group("/v1")
	v1.Get("/docs/*", swagger.New())

	s.registerGraphRoutes(v1)
	s.registerNodeRoutes(v1)
	s.registerEdgeRoutes(v1)

	return s, nil
}

// Listen validates the server options and starts listening on the bind address.
func (s *Server) Listen() error {
	if s.Addr == "" {
		return fmt.Errorf("empty bind address")
	}

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	s.ln = ln

	return s.app.Listener(ln)
}

// Close gracefully shuts down the server.
func (s *Server) Close(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		select {
		case <-ctx.Done():
		case errChan <- s.app.Shutdown():
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("server shut down: %v", ctx.Err())
	case err := <-errChan:
		return err
	}
}
