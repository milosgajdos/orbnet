package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
)

// Syncer is a filesystem syncer.
type Syncer struct {
	sync.RWMutex
	dst   string
	count int
}

// NewSyncer creates a new filesystem syncer and returns it.
// If dst is an empty string syncer streams the data to stdout.
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

// Sync stores the data received via channel ch on the filesystem
// nolint:revive
func (s *Syncer) Sync(ctx context.Context, ch <-chan interface{}) error {
	for repos := range ch {
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
