package http

import (
	"errors"

	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func validateEdge(e api.Edge) error {
	if len(e.Source) == 0 || len(e.Target) == 0 {
		return errors.New("invalid node ID")
	}

	if e.Source == e.Target {
		return errors.New("self-loop not allowed")
	}
	return nil
}

func StringPtr(s string) *string {
	return &s
}
