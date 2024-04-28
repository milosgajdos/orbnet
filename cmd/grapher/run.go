package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/builder/stars"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
	sig "github.com/milosgajdos/orbnet/pkg/signal"
)

const (
	// CliName is command line name.
	CliName = "grapher"
	// BuilderPool is the default size of graph builder pool.
	BuilderPool = 2
)

func run(args []string) error {
	flags := flag.NewFlagSet(CliName, flag.ExitOnError)

	var (
		indir    = flags.String("indir", "", "Input directory")
		marshal  = flags.Bool("marshal", false, "marshal graph to stdout")
		format   = flags.String("format", "dot", "Encoding format (dot, gexf, cytoscape, sigma, networkx, jsonapi)")
		builders = flags.Int("builders", BuilderPool, "number of graph builders")
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
			log.Printf("interrupting: got signal %s", s)
			cancel()
		case <-ctx.Done():
			return
		}
		// hard exit
		<-sigChan // second signal, hard exit
		os.Exit(1)
	}()

	g, err := memory.NewGraph(memory.WithLabel("orbnet"))
	if err != nil {
		return err
	}

	b, err := stars.NewBuilder(g)
	if err != nil {
		return err
	}

	f, err := NewFetcher(*indir)
	if err != nil {
		return err
	}

	reposChan := make(chan interface{}, *builders)
	errChan := make(chan error)

	for i := 0; i < *builders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case errChan <- b.Build(ctx, reposChan):
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
		return fmt.Errorf("error building graph: %v", err)
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
