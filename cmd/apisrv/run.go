package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/milosgajdos/orbnet/pkg/graph/api/http"
	"github.com/milosgajdos/orbnet/pkg/graph/api/memory"
)

const (
	// CliName is command line name.
	CliName = "apisrv"

	// http.Server timeouts
	IdleTimeout  = 5 * time.Second
	ReadTimeout  = 5 * time.Second
	WriteTimeout = 10 * time.Second

	// ShutdownTimeout defines time when we forcefully shutdown the server
	ShutdownTimeout = 10 * time.Second
)

func run(args []string) error {
	flags := flag.NewFlagSet(CliName, flag.ExitOnError)

	var (
		addr = flags.String("addr", ":5050", "API server bind address")
		dsn  = flags.String("dsn", ":memory:", "Database connection string")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	config := http.Config{
		IdleTimeout:  IdleTimeout,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
	}

	s, err := http.NewServer(config)
	if err != nil {
		return err
	}

	db, err := memory.NewDB(*dsn)
	if err != nil {
		return fmt.Errorf("failed creating new DB: %v", err)
	}
	if err := db.Open(); err != nil {
		return fmt.Errorf("failed opening DB: %v", err)
	}

	gs, err := memory.NewGraphService(db)
	if err != nil {
		return fmt.Errorf("failed creating graph service: %v", err)
	}

	ns, err := memory.NewNodeService(db)
	if err != nil {
		return fmt.Errorf("failed creating node service: %v", err)
	}

	es, err := memory.NewEdgeService(db)
	if err != nil {
		return fmt.Errorf("failed creating graph service: %v", err)
	}

	s.Addr = *addr
	s.GraphService = gs
	s.NodeService = ns
	s.EdgeService = es

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		errChan <- s.Listen()
	}()

	// Listen for the interrupt signal.
	select {
	case <-ctx.Done():
	case err := <-errChan:
		return err
	}

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// Perform application shutdown with a maximum timeout of ShutdownTimeout seconds.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	return s.Close(timeoutCtx)
}
