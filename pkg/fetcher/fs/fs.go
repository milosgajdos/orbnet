package fs

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/go-github/v34/github"
)

type Fetcher struct {
	files []os.FileInfo
	dir   string
}

func NewFetcher(dir string) (*Fetcher, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	return &Fetcher{
		files: files,
		dir:   dir,
	}, nil
}

func (f *Fetcher) Fetch(ctx context.Context, reposChan chan<- interface{}) error {
	defer close(reposChan)

	for _, file := range f.files {
		r, err := os.Open(filepath.Join(f.dir, file.Name()))
		if err != nil {
			return err
		}

		var repos []*github.StarredRepository

		if err := json.NewDecoder(r).Decode(&repos); err != nil {
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
