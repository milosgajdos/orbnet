package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func (s *Server) registerGraphRoutes(r fiber.Router) {
	routes := fiber.New()
	// get all graphs stored in the database
	routes.Get("/graphs", s.GetAllGraphs)
	// get a graph by UID
	routes.Get("/graphs/:uid", s.GetGraphByUID)
	// create a new graph
	routes.Post("/graphs", s.CreateGraph)
	// update existing graph
	routes.Patch("/graphs/:uid", s.UpdateGraph)
	// delete graph by UID
	routes.Delete("/graphs/:uid", s.DeleteGraph)
	// mount graph routes at the root of r
	r.Mount("/", routes)
}

// GetAllGraphs returns all available graphs.
// @Summary Get all graphs
// @Description Get all available graphs.
// @Tags graphs
// @Produce json
// @Param offset query int false "Result offset"
// @Param limit query int false "Result limit"
// @Success 200 {object} GraphsResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs [get]
func (s *Server) GetAllGraphs(c *fiber.Ctx) error {
	var filter api.GraphFilter
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

	graphs, n, err := s.GraphService.FindGraphs(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(GraphsResponse{
		Graphs: graphs,
		N:      n,
	})
}

// GetGraphByUID returns graph with the given UID.
// @Summary Get graph by UID.
// @Description Get graph returns graph with the given UID.
// @Tags graphs
// @Produce json
// @Param uid path string true "Graph UID"
// @Success 200 {object} api.Graph
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{uid} [get]
func (s *Server) GetGraphByUID(c *fiber.Ctx) error {
	graph, err := s.GraphService.FindGraphByUID(c.Context(), c.Params("uid"))
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

	return c.JSON(graph)
}

// CreateGraph creates a new graph.
// @Summary Create new graph.
// @Description Create new graph.
// @Tags graphs
// @Accept json
// @Produce json
// @Param graph body api.Graph true "Create a graph"
// @Success 200 {object} api.Graph
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs [post]
func (s *Server) CreateGraph(c *fiber.Ctx) error {
	graph := new(api.Graph)
	if err := c.BodyParser(graph); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	// TODO(milosgajdos): validate graph here
	if err := s.GraphService.CreateGraph(c.Context(), graph); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(graph)
}

// UpdateGraph updates existing graph.
// @Summary Update graph
// @Description Update existing graph.
// @Tags graphs
// @Accept json
// @Produce json
// @Param uid path string true "Graph UID"
// @Param graph body api.GraphUpdate true "Update a graph"
// @Success 200 {object} api.Graph
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{uid} [patch]
func (s *Server) UpdateGraph(c *fiber.Ctx) error {
	update := new(api.GraphUpdate)
	if err := c.BodyParser(update); err != nil {
		// TODO(milosgajdos): validate update data
		// Return status 400 and error message.
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	graph, err := s.GraphService.UpdateGraph(c.Context(), c.Params("uid"), *update)
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

	return c.JSON(graph)
}

// DeleteGraph deletes graph by uid.
// @Summary Delete graph by UID.
// @Description Delete graph with the given UID.
// @Tags graphs
// @Produce json
// @Param uid path string true "Graph UID"
// @Success 204 {string} status "Graph was deleted successfully"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{uid} [delete]
func (s *Server) DeleteGraph(c *fiber.Ctx) error {
	if err := s.GraphService.DeleteGraph(c.Context(), c.Params("uid")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
