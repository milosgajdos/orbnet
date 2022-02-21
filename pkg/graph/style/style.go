package style

import (
	"image/color"
)

var (
	// DefaultEdgeColor is default edge color.
	DefaultEdgeColor = color.RGBA{R: 0, G: 0, B: 0}
	// DefaultNodeColor is default color for unknown entity.
	DefaultNodeColor = color.RGBA{R: 255, G: 255, B: 255}
)

const (
	// DefaultStyleType is default style type.
	DefaultStyleType = "filling"
	// DefaultNodeShape is default node shape.
	DefaultNodeShape = "hexagon"
	// EdgeShape is default edge shape.
	DefaultEdgeShape = "normal"
	// UnknownShape is unknown shape.
	UnknownShape = "unknown"
)

// Style defines styling.
type Style struct {
	// Type is style type.
	Type string
	// Shape is style shape.
	Shape string
	// Color is style color.
	Color color.RGBA
}

// DefaultNode returns default node style
func DefaultNode() Style {
	return Style{
		Type:  DefaultStyleType,
		Shape: DefaultNodeShape,
		Color: DefaultNodeColor,
	}
}

// DefaultEdge returns default edge style
func DefaultEdge() Style {
	return Style{
		Type:  DefaultStyleType,
		Shape: DefaultEdgeShape,
		Color: DefaultEdgeColor,
	}
}
