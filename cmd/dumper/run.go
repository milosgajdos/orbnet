package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/milosgajdos/orbnet/pkg/fetcher/stars"
	"github.com/milosgajdos/orbnet/pkg/syncer/fs"
	"golang.org/x/sync/errgroup"
)

const (
	// CliName is the name of the command line.
	CliName = "dumper"
	// Paging is the default GitHub API paging size.
	Paging = 50
	// SyncerPool defines the default size of syncer pool.
	SyncerPool = 1
	// FetcherPool defines the default size of fetcher pool.
	FetcherPool = 1
	// MaxFetchers is the upper bound of fetch workers.
	MaxFetchers = 100
	// EnvGithubToken stores the name of the env var that store GitHub API token.
	EnvGithubToken = "GITHUB_TOKEN"
)

func run(args []string) error {
	flags := flag.NewFlagSet(CliName, flag.ExitOnError)

	var (
		user     = flags.String("user", "", "GitHub username")
		token    = flags.String("token", "", "GitHub API token (GITHUB_TOKEN)")
		outdir   = flags.String("outdir", "", "Output directory")
		paging   = flags.Int("paging", Paging, "GitHub API results paging limit")
		syncers  = flags.Int("syncers", SyncerPool, "syncer pool size")
		fetchers = flags.Int("fetchers", FetcherPool, "fetcher pool size")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if *token == "" {
		*token = os.Getenv(EnvGithubToken)
		if *token == "" {
			return errors.New("missing GitHub token")
		}
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer func() {
		signal.Stop(sigChan)
		cancel()
	}()
	go func() {
		select {
		case <-sigChan:
			fmt.Println("shutting down: received SIGINT...")
			cancel()
		case <-ctx.Done():
		}
	}()

	f, err := stars.NewFetcher(*token, *user, *paging)
	if err != nil {
		return fmt.Errorf("create fetcher: %w", err)
	}

	totalPages, err := f.GetTotalPages(ctx, *paging)
	if err != nil {
		return fmt.Errorf("get total pages: %w", err)
	}

	s, err := fs.NewSyncer(*outdir)
	if err != nil {
		return fmt.Errorf("create syncer: %w", err)
	}

	// TODO: move this to a dedicated func
	numWorkers := *fetchers
	// upper bound on concurrent requests
	if numWorkers > MaxFetchers {
		numWorkers = MaxFetchers
	}
	if numWorkers > totalPages {
		numWorkers = totalPages
		// NOTE: this is a silly heuristic,
		// but we dont want to run unnecessarily
		// large number of fetchers if not needed.
		if totalPages > 1 {
			numWorkers = totalPages / 2
		}
	}
	batchSize := totalPages / numWorkers
	remainder := totalPages % numWorkers

	reposChan := make(chan interface{}, *syncers)

	g, ctx := errgroup.WithContext(ctx)

	// launch syncers
	g.Go(func() error {
		sg, sctx := errgroup.WithContext(ctx)
		for i := 0; i < *syncers; i++ {
			sg.Go(func() error {
				return s.Sync(sctx, reposChan)
			})
		}
		return sg.Wait()
	})
	// launch fetchers
	g.Go(func() error {
		defer close(reposChan)
		fg, fctx := errgroup.WithContext(ctx)
		startPage := 1
		for i := 0; i < numWorkers; i++ {
			endPage := startPage + batchSize - 1
			if i < remainder {
				endPage++
			}
			sp, ep := startPage, endPage
			fg.Go(func() error {
				return f.Fetch(fctx, sp, ep, reposChan)
			})
			startPage = endPage + 1
		}
		return fg.Wait()
	})

	if err := g.Wait(); err != nil {
		if err != context.Canceled {
			return fmt.Errorf("encountered error: %v", err)
		}
	}
	return nil
}
