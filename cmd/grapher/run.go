package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/builder/stars"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
	"golang.org/x/sync/errgroup"
)

const (
	// CliName is command line name.
	CliName = "grapher"
	// GraphLabel is the default graph label
	GraphLabel = "orbnet"
	// BuilderPool is the default size of graph builder pool.
	BuilderPool = 1
)

func run(args []string) error {
	flags := flag.NewFlagSet(CliName, flag.ExitOnError)

	var (
		input    = flags.String("input", "", "input source (default: stdin)")
		marshal  = flags.Bool("marshal", false, "marshal graph to stdout")
		format   = flags.String("format", "dot", "encoding format (dot, gexf, cytoscape, sigma, networkx, jsonapi)")
		builders = flags.Int("builders", BuilderPool, "number of graph builders")
		label    = flags.String("flabel", GraphLabel, "graph label")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	var m graph.Marshaler
	if *marshal {
		var err error
		m, err = NewMarshaler(*format, "GitHub Stars", "", "\t")
		if err != nil {
			return err
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

	g, err := memory.NewGraph(memory.WithLabel(*label))
	if err != nil {
		return err
	}

	b, err := stars.NewBuilder(g)
	if err != nil {
		return err
	}

	f, err := NewFetcher(*input)
	if err != nil {
		return err
	}

	reposChan := make(chan interface{}, *builders)

	eg, ctx := errgroup.WithContext(ctx)

	for i := 0; i < *builders; i++ {
		eg.Go(func() error {
			return b.Build(ctx, reposChan)
		})
	}

	eg.Go(func() error {
		return f.Fetch(ctx, reposChan)
	})

	if err := eg.Wait(); err != nil {
		if err != context.Canceled {
			return fmt.Errorf("encountered error: %v", err)
		}
	}

	if *marshal {
		out, err := m.Marshal(g)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", out)
	}
	return nil
}
