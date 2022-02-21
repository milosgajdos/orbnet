package dot

import (
	"github.com/milosgajdos/orbnet/pkg/graph"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

// Marshaler is used for marshaling graph to DOT format.
type Marshaler struct {
	name   string
	prefix string
	indent string
}

// NewMarshaler creates a new DOT graph marshaler and returns it.
func NewMarshaler(name, prefix, indent string) (*Marshaler, error) {
	return &Marshaler{
		name:   name,
		prefix: prefix,
		indent: indent,
	}, nil
}

// Marshal marshal g into DOT and returns it.
func (m *Marshaler) Marshal(g graph.Graph) ([]byte, error) {
	return dot.Marshal(g, m.name, m.prefix, m.indent)
}
