package main

import (
	"fmt"

	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/api/memory/marshal/json"
	"github.com/milosgajdos/orbnet/pkg/graph/marshal/cytoscape"
	"github.com/milosgajdos/orbnet/pkg/graph/marshal/dot"
	"github.com/milosgajdos/orbnet/pkg/graph/marshal/gexf"
	"github.com/milosgajdos/orbnet/pkg/graph/marshal/networkx"
	"github.com/milosgajdos/orbnet/pkg/graph/marshal/sigma"
)

func NewMarshaler(format string, name, prefix, indent string) (graph.Marshaler, error) {
	switch format {
	case "dot":
		return dot.NewMarshaler(name, prefix, indent)
	case "gexf":
		return gexf.NewMarshaler(name, prefix, indent)
	case "cytoscape":
		return cytoscape.NewMarshaler(name, prefix, indent)
	case "sigma":
		return sigma.NewMarshaler(name, prefix, indent)
	case "networkx":
		return networkx.NewMarshaler(name, prefix, indent)
	case "jsonapi":
		return json.NewMarshaler(name, prefix, indent)
	}

	return nil, fmt.Errorf("unsupported format: %q", format)
}
