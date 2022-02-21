package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
)

type Syncer struct {
	sync.RWMutex
	dst   string
	count int
}

func NewSyncer(dst string) (*Syncer, error) {
	return &Syncer{
		dst:   dst,
		count: 0,
	}, nil
}

func (s *Syncer) newEncoder(idx int) (*json.Encoder, error) {
	if s.dst != "" {
		path := path.Join(s.dst, fmt.Sprintf("%d.json", idx))
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		return json.NewEncoder(f), nil
	}
	return json.NewEncoder(os.Stdout), nil
}

func (s *Syncer) Sync(ctx context.Context, reposChan <-chan interface{}) error {
	for repos := range reposChan {
		s.RLock()
		idx := s.count
		s.RUnlock()

		e, err := s.newEncoder(idx)
		if err != nil {
			return err
		}

		if err := e.Encode(repos); err != nil {
			return err
		}

		s.Lock()
		s.count++
		s.Unlock()
	}
	return nil
}
