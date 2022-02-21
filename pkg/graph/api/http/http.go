package http

import "github.com/milosgajdos/orbnet/pkg/graph/api"

const (
	// DefaultLimit defines default results limit
	DefaultLimit = 20
)

// GraphsResponse is returned when querying graphs.
type GraphsResponse struct {
	Graphs []*api.Graph `json:"graphs"`
	N      int          `json:"n"`
}

// NodesResponse is returned when querying nodes.
type NodesResponse struct {
	Nodes []*api.Node `json:"nodes"`
	N     int         `json:"n"`
}

// EdgesReponse is returned when querying edges.
type EdgesResponse struct {
	Edges []*api.Edge `json:"edges"`
	N     int         `json:"n"`
}

// ErrorResponse represents a JSON structure for error output.
type ErrorResponse struct {
	Error string `json:"error"`
}
