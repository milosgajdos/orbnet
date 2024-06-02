package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func (s *Server) registerEdgeRoutes(r fiber.Router) {
	routes := fiber.New()
	// get all edges in the given graph
	routes.Get("/:guid/edges", s.GetAllEdges)
	// get an edge in the graph by uid
	routes.Get("/:guid/edges/:uid", s.GetEdgeByUID)
	// create a new edge in the graph
	routes.Post("/:guid/edges", s.CreateEdge)
	// update existing edge in the graph
	routes.Patch("/:guid/edges", s.UpdateEdgeBetween)
	// delete edge in the graph by UID
	routes.Delete("/:guid/edges/:uid", s.DeleteEdge)
	// delete edge between two nodes in the graph.
	routes.Delete("/:guid/edges", s.DeleteEdgeBetween)
	// mount edge routes to /graphs
	r.Mount("/graphs", routes)
}

// GetAllEdges returns all edges in the graph matching a query.
// @Summary Get graph edges matching a filter.
// @Description Get all edges matching a query.
// @Tags edges
// @Produce json
// @Param offset query int false "Result offset"
// @Param limit query int false "Result limit"
// @Param label query string false "Node label"
// @Param guid path string true "Graph UID"
// @Success 200 {object} EdgesResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/edges [get]
func (s *Server) GetAllEdges(c *fiber.Ctx) error {
	graphUID := c.Params("guid")

	var filter api.EdgeFilter
	filter.Limit = DefaultLimit

	// NOTE(milosgajdos): we don't care if the conversion fails
	offset, _ := strconv.Atoi(c.Query("offset"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	if offset > 0 {
		filter.Offset = offset
	}

	if limit > 0 {
		filter.Limit = limit
	}

	if label := c.Query("label"); label != "" {
		filter.Label = new(string)
		*filter.Label = c.Query("label")
	}

	edges, n, err := s.EdgeService.FindEdges(c.Context(), graphUID, filter)
	if err != nil {
		if code := api.ErrorCode(err); code == api.ENOTFOUND {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(EdgesResponse{
		Edges: edges,
		N:     n,
	})
}

// GetEdgeByUID returns graph edge with the given UID.
// @Summary Get graph edge by UID.
// @Description Get graph a single graph edge with the given UID.
// @Tags edges
// @Produce json
// @Param guid path string true "Graph UID"
// @Param uid path string true "Edge UID"
// @Success 200 {object} api.Edge
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/edges/{uid} [get]
func (s *Server) GetEdgeByUID(c *fiber.Ctx) error {
	graphUID := c.Params("guid")
	nodeUID := c.Params("uid")

	edge, err := s.EdgeService.FindEdgeByUID(c.Context(), graphUID, nodeUID)
	if err != nil {
		if code := api.ErrorCode(err); code == api.ENOTFOUND {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(edge)
}

// CreateEdge creates a new graph edge between two nodes.
// @Summary Create new graph edge.
// @Description Create a new graph edge and return it.
// @Tags edges
// @Accept json
// @Produce json
// @Param graph body api.Edge true "Create a edge"
// @Success 200 {object} api.Edge
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/edges [post]
func (s *Server) CreateEdge(c *fiber.Ctx) error {
	graphUID := c.Params("guid")

	edge := new(api.Edge)
	if err := c.BodyParser(edge); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	if err := validateEdge(*edge); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	if err := s.EdgeService.CreateEdge(c.Context(), graphUID, edge); err != nil {
		if code := api.ErrorCode(err); code == api.ENOTFOUND {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(edge)
}

// UpdateEdgeBetween updates an existing edge between two nodes.
// @Summary Update graph edge between two existing nodes.
// @Description Update existing graph edge between the nodes with the given IDs.
// @Tags edges
// @Accept json
// @Produce json
// @Param uid path string true "Graph UID"
// @Param graph body api.EdgeUpdate true "Update an edge"
// @Param source query int false "Source node ID"
// @Param target query int false "Target node ID"
// @Success 200 {object} api.Edge
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/edges [patch]
func (s *Server) UpdateEdgeBetween(c *fiber.Ctx) error {
	graphUID := c.Params("guid")

	source := c.Query("source")
	if len(source) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "empty source node",
		})
	}

	target := c.Query("target")
	if len(target) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "empty target node",
		})
	}

	update := new(api.EdgeUpdate)
	if err := c.BodyParser(update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	edge, err := s.EdgeService.UpdateEdgeBetween(c.Context(), graphUID, source, target, *update)
	if err != nil {
		if code := api.ErrorCode(err); code == api.ENOTFOUND {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(edge)
}

// DeleteEdge deletes graph edge with the given UID.
// @Summary Delete graph edge by UID.
// @Description Delete graph edge with the given UID.
// @Tags edges
// @Produce json
// @Param guid path string true "Graph UID"
// @Param uid path string true "Edge UID"
// @Success 204 {string} status "Edge was deleted successfully"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/edges/{uid} [delete]
func (s *Server) DeleteEdge(c *fiber.Ctx) error {
	graphUID := c.Params("guid")
	edgeUID := c.Params("uid")

	if err := s.EdgeService.DeleteEdge(c.Context(), graphUID, edgeUID); err != nil {
		if code := api.ErrorCode(err); code == api.ENOTFOUND {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteEdgeBetween deletes a graph edge between two nodes.
// @Summary Delete graph edge between two nodes.
// @Description Delete graph edge between nodes with given IDs.
// @Tags edges
// @Produce json
// @Param uid path string true "Graph UID"
// @Param source query int false "Source node ID"
// @Param target query int false "Target node ID"
// @Success 204 {string} status "Edge was deleted successfully"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/edges [delete]
func (s *Server) DeleteEdgeBetween(c *fiber.Ctx) error {
	graphUID := c.Params("guid")

	source := c.Query("source")
	if len(source) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "empty source node",
		})
	}

	target := c.Query("target")
	if len(target) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "empty target node",
		})
	}

	if err := s.EdgeService.DeleteEdgeBetween(c.Context(), graphUID, source, target); err != nil {
		if code := api.ErrorCode(err); code == api.ENOTFOUND {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
