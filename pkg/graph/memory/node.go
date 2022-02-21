package memory

import (
	"image/color"

	"github.com/google/uuid"
	"github.com/milosgajdos/orbnet/pkg/graph/attrs"
	"github.com/milosgajdos/orbnet/pkg/graph/style"
	"gonum.org/v1/gonum/graph/encoding"
)

// Node is a graph node.
type Node struct {
	id    int64
	uid   string
	dotid string
	label string
	attrs map[string]interface{}
	style style.Style
}

// NewNode creates a new Node and returns it.
func NewNode(id int64, opts ...Option) (*Node, error) {
	uid := uuid.New().String()

	nopts := Options{
		UID:   uid,
		DotID: uid,
		Attrs: make(map[string]interface{}),
		Style: style.DefaultNode(),
	}

	for _, apply := range opts {
		apply(&nopts)
	}

	return &Node{
		id:    id,
		uid:   nopts.UID,
		dotid: nopts.DotID,
		label: nopts.Label,
		attrs: nopts.Attrs,
		style: nopts.Style,
	}, nil
}

// ID returns node ID.
func (n Node) ID() int64 {
	return n.id
}

// UID returns node UID.
func (n Node) UID() string {
	return n.uid
}

// Label returns node label.
func (n Node) Label() string {
	return n.label
}

// SetLabel sets node label.
func (n *Node) SetLabel(l string) {
	n.label = l
}

// SetUID sets UID.
func (n *Node) SetUID(uid string) {
	n.uid = uid
}

// Attrs returns node attributes.
func (n *Node) Attrs() map[string]interface{} {
	return n.attrs
}

// Type returns the type of node style.
func (n Node) Type() string {
	return n.style.Type
}

// Shape returns node shape.
func (n Node) Shape() string {
	return n.style.Shape
}

// Color returns node color.
func (n Node) Color() color.RGBA {
	return n.style.Color
}

// DOTID returns GraphVIz DOT ID.
func (n Node) DOTID() string {
	return n.dotid
}

// SetDOTID sets GraphVIz DOT ID.
func (n *Node) SetDOTID(dotid string) {
	n.dotid = dotid
}

// Attributes returns node DOT attributes.
func (n Node) Attributes() []encoding.Attribute {
	a := attrs.ToStringMap(n.attrs)
	attributes := make([]encoding.Attribute, len(a))

	i := 0
	for k, v := range a {
		attributes[i] = encoding.Attribute{Key: k, Value: v}
		i++
	}
	return attributes
}
