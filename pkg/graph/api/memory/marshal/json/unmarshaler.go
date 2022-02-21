package json

import (
	"context"
	"encoding/json"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/api/memory/marshal"
)

// Unmarshaler implements graph.Unmarshaler.
type Unmarshaler struct{}

// NewUnmarshaler creates a new Unmarshaler and returns it.
func NewUnmarshaler() (*Unmarshaler, error) {
	return &Unmarshaler{}, nil
}

// Unmarshal unmarshals data into graph.
func (u *Unmarshaler) Unmarshal(data []byte, g graph.Graph) error {
	ga, ok := g.(graph.Adder)
	if !ok {
		return graph.Errorf(graph.EUNSUPPORTED, "Unable to update graph")
	}

	var mg marshal.Graph
	if err := json.Unmarshal(data, &mg); err != nil {
		return err
	}

	b, err := marshal.NewBuilder()
	if err != nil {
		return err
	}

	dataChan := make(chan interface{})
	errChan := make(chan error)
	go func() {
		errChan <- b.Build(context.Background(), ga, dataChan)
	}()

	dataChan <- mg
	close(dataChan)

	if err := <-errChan; err != nil {
		return graph.Errorf(graph.EINTERNAL, "Failed to build graph: %v", err)
	}

	return nil
}
