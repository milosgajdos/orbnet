package stars

import "image/color"

var (
	// OwnerColor is default owner node color.
	OwnerColor = color.RGBA{R: 245, G: 154, B: 240}
	// RepoColor is default repo node color.
	RepoColor = color.RGBA{R: 0, G: 255, B: 153}
	// TopicColor is default topic node color.
	TopicColor = color.RGBA{R: 153, G: 153, B: 255}
	// LangColor is default lang node color.
	LangColor = color.RGBA{R: 255, G: 133, B: 102}
	// LinkColor is default link color.
	LinkColor = color.RGBA{R: 0, G: 0, B: 0}
	// UnknownColor is default color for unknown entity.
	UnknownColor = color.RGBA{R: 255, G: 255, B: 255}
)

const (
	// DefaultStyleType is default style type.
	DefaultStyleType = "filling"
	// OwnerShape is default owner node shape.
	OwnerShape = "hexagon"
	// RepoShape is default repo node shape.
	RepoShape = "diamond"
	// TopicShape is default topic node shape.
	TopicShape = "ellipse"
	// LangShape is default lang node shape.
	LangShape = "square"
	// LinkShape is default link shape.
	LinkShape = "normal"
	// UnknownShape is unknown shape.
	UnknownShape = "unknown"
)
