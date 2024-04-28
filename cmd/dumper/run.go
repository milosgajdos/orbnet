package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/milosgajdos/orbnet/pkg/fetcher/stars"
	sig "github.com/milosgajdos/orbnet/pkg/signal"
	"github.com/milosgajdos/orbnet/pkg/syncer/fs"
)

const (
	// CliName is the name of the command line.
	CliName = "dumper"
	// SyncerPool defines the default size of syncer pool.
	SyncerPool = 2
	// Paging is the default GitHub API paging size.
	Paging = 50
	// ENV_GITHUB_TOKEN stores the name of the env var that store GitHub API token.
	ENV_GITHUB_TOKEN = "GITHUB_TOKEN"
)

func run(args []string) error {
	flags := flag.NewFlagSet(CliName, flag.ExitOnError)

	var (
		user    = flags.String("user", "", "GitHub username")
		token   = flags.String("token", "", "GitHub API token (GITHUB_TOKEN)")
		outdir  = flags.String("outdir", "", "Output directory")
		paging  = flags.Int("paging", Paging, "GitHub API results paging limit")
		syncers = flags.Int("syncers", SyncerPool, "syncer pool size")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if *token == "" {
		*token = os.Getenv(ENV_GITHUB_TOKEN)
		if *token == "" {
			return errors.New("missing GitHub token")
		}
	}

	sigChan := sig.Register(os.Interrupt)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		signal.Stop(sigChan)
		cancel()
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case s := <-sigChan:
			log.Printf("iinterrupting: got signal %s", s)
			cancel()
		case <-ctx.Done():
		}
	}()

	f, err := stars.NewFetcher(*token, *user, *paging)
	if err != nil {
		return fmt.Errorf("create fetcher: %w", err)
	}

	s, err := fs.NewSyncer(*outdir)
	if err != nil {
		return fmt.Errorf("create syncer: %w", err)
	}

	reposChan := make(chan interface{}, *syncers)
	errChan := make(chan error)

	for i := 0; i < *syncers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case errChan <- s.Sync(ctx, reposChan):
			case <-ctx.Done():
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case errChan <- f.Fetch(ctx, reposChan):
		case <-ctx.Done():
		}
	}()

	select {
	case err = <-errChan:
		cancel()
	case <-ctx.Done():
	}

	// notify signal handler
	cancel()
	// wait for all goroutine to stop
	wg.Wait()

	if err != nil {
		return fmt.Errorf("error dumping repos: %v", err)
	}

	return nil
}
