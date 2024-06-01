package main

import (
	"os"

	"github.com/milosgajdos/orbnet/pkg/fetcher"
	"github.com/milosgajdos/orbnet/pkg/fetcher/fs"
	"github.com/milosgajdos/orbnet/pkg/fetcher/stream"
)

// NewFetcher creates a new fetcher.Fetcher and returns it.
// If input is empty string, it returns stream.Fetcher.
// Otherwise it considers input to be a feilsystem path.
func NewFetcher(input string) (fetcher.Fetcher, error) {
	if input != "" {
		f, err := fs.NewFetcher(input)
		if err != nil {
			return nil, err
		}
		return f, nil
	}

	return stream.NewFetcher(os.Stdin)
}
