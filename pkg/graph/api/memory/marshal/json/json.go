package json

import "github.com/milosgajdos/orbnet/pkg/graph"

const (
	DefaultName   = "marshaler"
	DefaultPrefix = ""
	DefaultIndent = "  "
)

// Marshal marshals g into JSON.
func Marshal(g graph.Graph) ([]byte, error) {
	m, err := NewMarshaler(DefaultName, DefaultPrefix, DefaultIndent)
	if err != nil {
		return nil, err
	}

	return m.Marshal(g)
}

// Unmarshal unmarshals JSON data into g.
func Unmarshal(data []byte, g graph.Graph) error {
	m, err := NewUnmarshaler()
	if err != nil {
		return err
	}

	return m.Unmarshal(data, g)
}
