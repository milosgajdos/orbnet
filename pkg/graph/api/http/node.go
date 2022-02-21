package http

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/milosgajdos/orbnet/pkg/graph/api"
)

func (s *Server) registerNodeRoutes(r fiber.Router) {
	routes := fiber.New()
	// get nodes in the given graph
	routes.Get("/:guid/nodes", s.GetNodes)
	// get a node in the graph by uid
	routes.Get("/:guid/nodes/uid/:uid", s.GetNodeByUID)
	// get a node in the graph by id
	routes.Get("/:guid/nodes/:id", s.GetNodeByID)
	// create a new node in the graph
	routes.Post("/:guid/nodes", s.CreateNode)
	// update existing node in the graph
	routes.Patch("/:guid/nodes/:id", s.UpdateNode)
	// delete node in the graph by UID
	routes.Delete("/:guid/nodes/:id", s.DeleteNodeByID)
	// delete a node in the graph by uid
	routes.Delete("/:guid/nodes/uid/:uid", s.DeleteNodeByUID)
	// mount graph routes to /graphs
	r.Mount("/graphs", routes)
}

// GetNodes returns all nodes in the graph matching a query.
// @Summary Graph nodes matching a filter.
// @Description Get all nodes matching a query.
// @Tags nodes
// @Produce json
// @Param offset query int false "Result offset"
// @Param limit query int false "Result limit"
// @Param label query string false "Node label"
// @Param guid path string true "Graph UID"
// @Success 200 {object} NodesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/nodes [get]
func (s *Server) GetNodes(c *fiber.Ctx) error {
	// Grab Graph UID from request.
	uid, err := uuid.Parse(c.Params("guid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	var filter api.NodeFilter
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

	nodes, n, err := s.NodeService.FindNodes(context.TODO(), uid.String(), filter)
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

	return c.JSON(NodesResponse{
		Nodes: nodes,
		N:     n,
	})
}

// GetNodeByUID returns a single graph node with the given UID.
// @Summary Get graph node by UID.
// @Description Get graph returns a single graph node with the given UID.
// @Tags nodes
// @Produce json
// @Param guid path string true "Graph UID"
// @Param uid path string true "Node UID"
// @Success 200 {object} api.Node
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/nodes/uid/{uid} [get]
func (s *Server) GetNodeByUID(c *fiber.Ctx) error {
	// Grab Graph UID from request.
	guid, err := uuid.Parse(c.Params("guid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	uid, err := uuid.Parse(c.Params("uid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	node, err := s.NodeService.FindNodeByUID(context.TODO(), guid.String(), uid.String())
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

	return c.JSON(node)
}

// GetNodeByID returns a single graph node with the given ID.
// @Summary Get graph node by ID.
// @Description Get a single graph node with the given ID.
// @Tags nodes
// @Produce json
// @Param guid path string true "Graph UID"
// @Param id path string true "Node ID"
// @Success 200 {object} api.Node
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/nodes/{id} [get]
func (s *Server) GetNodeByID(c *fiber.Ctx) error {
	guid, err := uuid.Parse(c.Params("guid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	if id < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid node ID",
		})
	}

	node, err := s.NodeService.FindNodeByID(context.TODO(), guid.String(), id)
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

	return c.JSON(node)
}

// CreateNode creates a new graph node.
// @Summary Create new graph node.
// @Description Create a new graph node and returns it.
// @Tags nodes
// @Accept json
// @Produce json
// @Param graph body api.Node true "Create a node"
// @Success 200 {object} api.Node
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/nodes [post]
func (s *Server) CreateNode(c *fiber.Ctx) error {
	guid, err := uuid.Parse(c.Params("guid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	node := new(api.Node)
	if err := c.BodyParser(node); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	// TODO(milosgajdos): validate node here
	if err := s.NodeService.CreateNode(context.TODO(), guid.String(), node); err != nil {
		if code := api.ErrorCode(err); code == api.ENOTFOUND {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(node)
}

// UpdateNode updates existing graph node.
// @Summary Update graph node.
// @Description Update existing graph node.
// @Tags nodes
// @Accept json
// @Produce json
// @Param uid path string true "Graph UID"
// @Param id path string true "Node ID"
// @Param graph body api.NodeUpdate true "Update a node"
// @Success 200 {object} api.Node
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/nodes/{id} [patch]
func (s *Server) UpdateNode(c *fiber.Ctx) error {
	// Grab Graph UID from request.
	uid, err := uuid.Parse(c.Params("guid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	if id < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid node ID",
		})
	}

	update := new(api.NodeUpdate)
	if err := c.BodyParser(update); err != nil {
		// TODO(milosgajdos): validate update data
		// Return status 400 and error message.
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	node, err := s.NodeService.UpdateNode(context.TODO(), uid.String(), id, *update)
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

	return c.JSON(node)
}

// DeleteNodeByID deletes a graph node by id.
// @Summary Delete graph node by ID.
// @Description Delete graph node with the given ID.
// @Tags nodes
// @Produce json
// @Param uid path string true "Graph UID"
// @Param id path string true "Node ID"
// @Success 204 {string} status "Node was deleted successfully"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/nodes/{id} [delete]
func (s *Server) DeleteNodeByID(c *fiber.Ctx) error {
	guid, err := uuid.Parse(c.Params("guid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	if id < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid node ID",
		})
	}

	if err := s.NodeService.DeleteNodeByID(context.TODO(), guid.String(), id); err != nil {
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

// DeleteNodeByUID deletes a graph node by UID.
// @Summary Delete graph node by UID.
// @Description Delete graph node with the given UID.
// @Tags nodes
// @Produce json
// @Param uid path string true "Graph UID"
// @Param id path string true "Node ID"
// @Success 204 {string} status "Node was deleted successfully"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/graphs/{guid}/nodes/uid/{uid} [delete]
func (s *Server) DeleteNodeByUID(c *fiber.Ctx) error {
	guid, err := uuid.Parse(c.Params("guid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	uid, err := uuid.Parse(c.Params("uid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	if err := s.NodeService.DeleteNodeByUID(context.TODO(), guid.String(), uid.String()); err != nil {
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
