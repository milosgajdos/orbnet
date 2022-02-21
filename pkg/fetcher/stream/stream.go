package stream

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"github.com/google/go-github/v34/github"
)

type Fetcher struct {
	d *json.Decoder
}

func NewFetcher(r io.Reader) (*Fetcher, error) {
	return &Fetcher{
		d: json.NewDecoder(bufio.NewReader(r)),
	}, nil
}

func (s *Fetcher) Fetch(ctx context.Context, reposChan chan<- interface{}) error {
	defer close(reposChan)

	for {
		var repos []*github.StarredRepository

		if err := s.d.Decode(&repos); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		select {
		case reposChan <- repos:
		case <-ctx.Done():
			return nil
		}
	}

	return nil
}
