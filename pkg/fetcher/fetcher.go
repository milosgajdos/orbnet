package fetcher

import (
	"context"
)

// Fetcher is used for fetching GitHub repos.
type Fetcher interface {
	// Fetch fetches GitHub repos into the provided channel.
	Fetch(context.Context, chan<- interface{}) error
}
