package main

import (
	"os"

	"github.com/milosgajdos/orbnet/pkg/fetcher"
	"github.com/milosgajdos/orbnet/pkg/fetcher/fs"
	"github.com/milosgajdos/orbnet/pkg/fetcher/stream"
)

// NewFetcher creates a new fetcher.Fetcher and returns it.
// If indir is empty string, it returns stream.Fetcher.
func NewFetcher(indir string) (fetcher.Fetcher, error) {
	if indir != "" {
		f, err := fs.NewFetcher(indir)
		if err != nil {
			return nil, err
		}
		return f, nil
	}

	return stream.NewFetcher(os.Stdin)
}
