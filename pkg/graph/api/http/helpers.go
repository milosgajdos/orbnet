package http

import (
	"errors"
	"strconv"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func validateEdge(e api.Edge) error {
	if e.Source < 0 || e.Target < 0 {
		return errors.New("invalid node ID")
	}

	if e.Source == e.Target {
		return errors.New("self-loop not allowed")
	}
	return nil
}

func parseNodeID(id string) (int64, error) {
	nodeID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return -100, errors.New("invalid node ID")
	}

	if nodeID < 0 {
		return -100, errors.New("invalid node ID")
	}
	return nodeID, nil
}

func StringPtr(s string) *string {
	return &s
}
