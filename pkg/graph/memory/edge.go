package memory

import (
	"image/color"

	"github.com/google/uuid"
	"github.com/milosgajdos/orbnet/pkg/graph/attrs"
	"github.com/milosgajdos/orbnet/pkg/graph/style"
	gonum "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
)

// Edge is a weighted graph edge.
type Edge struct {
	uid    string
	from   gonum.Node
	to     gonum.Node
	weight float64
	label  string
	attrs  map[string]interface{}
	style  style.Style
}

// NewEdge creates a new edge and returns it.
func NewEdge(from, to gonum.Node, opts ...Option) (*Edge, error) {
	uid := uuid.New().String()

	eopts := Options{
		UID:    uid,
		DotID:  uid,
		Weight: DefaultWeight,
		Attrs:  make(map[string]interface{}),
		Style:  style.DefaultEdge(),
	}

	for _, apply := range opts {
		apply(&eopts)
	}

	return &Edge{
		uid:    eopts.UID,
		from:   from,
		to:     to,
		weight: eopts.Weight,
		label:  eopts.Label,
		attrs:  eopts.Attrs,
		style:  eopts.Style,
	}, nil
}

// UID returns edge UID.
func (e Edge) UID() string {
	return e.uid
}

// From returns the from node of the first non-nil edge, or nil.
func (e *Edge) From() gonum.Node {
	return e.from
}

// To returns the to node of the first non-nil edge, or nil.
func (e *Edge) To() gonum.Node {
	return e.to
}

// Weight returns edge weight
func (e Edge) Weight() float64 {
	return e.weight
}

// SetWeight sets edge weight.
func (e *Edge) SetWeight(w float64) {
	e.weight = w
}

// ReversedEdge returns a new edge with end points of the pair swapped.
func (e *Edge) ReversedEdge() gonum.Edge {
	return &Edge{
		uid:   e.uid,
		from:  e.to,
		to:    e.from,
		label: e.label,
		attrs: e.attrs,
		style: e.style,
	}
}

// Label returns edge label.
func (e Edge) Label() string {
	return e.label
}

// SetLabel sets edge label.
func (e *Edge) SetLabel(l string) {
	e.label = l
}

// SetUID sets UID.
func (e *Edge) SetUID(uid string) {
	e.uid = uid
}

// Attrs returns node attributes.
func (e *Edge) Attrs() map[string]interface{} {
	return e.attrs
}

// Style returns edge style.
func (e Edge) Style() string {
	return e.style.Type
}

// Shape returns edge shape.
func (e Edge) Shape() string {
	return e.style.Shape
}

// Color returns edge color.
func (e Edge) Color() color.RGBA {
	return e.style.Color
}

// Attributes returns node DOT attributes.
func (e Edge) Attributes() []encoding.Attribute {
	a := attrs.ToStringMap(e.attrs)
	attributes := make([]encoding.Attribute, len(a))

	i := 0
	for k, v := range a {
		attributes[i] = encoding.Attribute{Key: k, Value: v}
		i++
	}
	return attributes
}
