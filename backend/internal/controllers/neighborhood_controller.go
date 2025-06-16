package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/services"
)

// NeighborhoodController handles HTTP requests related to neighborhoods
type NeighborhoodController struct {
	svc services.NeighborhoodService
}

// NewNeighborhoodController creates a new instance of NeighborhoodController
func NewNeighborhoodController(svc services.NeighborhoodService) *NeighborhoodController {
	return &NeighborhoodController{svc: svc}
}

// Register registers the routes for the neighborhood controller
func (ctrl *NeighborhoodController) Register(g *echo.Group) {
	g.POST("/neighborhoods", ctrl.CreateNeighborhood)
	g.GET("/neighborhoods/:id", ctrl.GetNeighborhoodByID)
	g.GET("/neighborhoods", ctrl.GetAllNeighborhoods)
}

// CreateNeighborhood handles the creation of a new neighborhood
func (ctrl *NeighborhoodController) CreateNeighborhood(c echo.Context) error {
	var neighborhood models.Neighborhood
	if err := c.Bind(&neighborhood); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := ctrl.svc.CreateNeighborhood(c.Request().Context(), &neighborhood); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create neighborhood",
		})
	}

	return c.JSON(http.StatusCreated, neighborhood)
}

// GetNeighborhoodByID handles retrieving a neighborhood by ID
func (ctrl *NeighborhoodController) GetNeighborhoodByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid neighborhood ID",
		})
	}

	neighborhood, err := ctrl.svc.GetNeighborhoodByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Neighborhood not found",
		})
	}

	return c.JSON(http.StatusOK, neighborhood)
}

// GetAllNeighborhoods handles retrieving all neighborhoods
func (ctrl *NeighborhoodController) GetAllNeighborhoods(c echo.Context) error {
	neighborhoods, err := ctrl.svc.GetAllNeighborhoods(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve neighborhoods",
		})
	}

	return c.JSON(http.StatusOK, neighborhoods)
}