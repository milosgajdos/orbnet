package neptune

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/milosgajdos/orbnet/pkg/graph"
)

func MarshalNodesCSV(g graph.Graph, propMap map[string]func(interface{}) string, w io.Writer) error {
	out := csv.NewWriter(w)
	header := []string{"~id", "~label"}

	// nolint:prealloc
	var properties []string
	for p := range propMap {
		properties = append(properties, p)
		header = append(header, p)
	}

	if err := out.Write(header); err != nil {
		return err
	}

	record := make([]string, len(header))

	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node().(graph.Node)

		record[0] = fmt.Sprintf("%d", n.ID())
		record[1] = n.Label()

		for i, p := range properties {
			px := strings.Split(p, ":")
			record[i+2] = propMap[p](n.Attrs()[px[0]])
		}

		if err := out.Write(record); err != nil {
			return err
		}
	}

	out.Flush()

	return out.Error()
}

func MarshalEdgesCSV(g graph.Graph, propMap map[string]func(interface{}) string, w io.Writer) error {
	out := csv.NewWriter(w)
	header := []string{"~id", "~from", "~to", "~label"}

	// nolint:prealloc
	var properties []string
	for p := range propMap {
		properties = append(properties, p)
		header = append(header, p)
	}

	if err := out.Write(header); err != nil {
		return err
	}

	record := make([]string, len(header))

	edges := g.Edges()
	for edges.Next() {
		e := edges.Edge().(graph.Edge)

		record[0] = e.UID()
		record[1] = fmt.Sprintf("%d", e.From().ID())
		record[2] = fmt.Sprintf("%d", e.To().ID())
		record[3] = e.Label()

		for i, p := range properties {
			px := strings.Split(p, ":")
			record[i+4] = propMap[p](e.Attrs()[px[0]])
		}

		if err := out.Write(record); err != nil {
			return err
		}
	}

	out.Flush()

	return out.Error()
}
