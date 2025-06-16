package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/services"
)

// ReportController handles HTTP requests related to reports
type ReportController struct {
	svc services.ReportService
}

// NewReportController creates a new instance of ReportController
func NewReportController(svc services.ReportService) *ReportController {
	return &ReportController{svc: svc}
}

// Register registers the routes for the report controller
func (ctrl *ReportController) Register(g *echo.Group) {
	g.POST("/reports", ctrl.CreateReport)
	g.POST("/reports/process-text", ctrl.ProcessReportText)
}

// CreateReport handles the creation of a new report
func (ctrl *ReportController) CreateReport(c echo.Context) error {
	var report models.Report
	if err := c.Bind(&report); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := ctrl.svc.CreateReport(c.Request().Context(), &report); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create report",
		})
	}

	return c.JSON(http.StatusCreated, report)
}

// ProcessReportText handles processing of text-based report requests
func (ctrl *ReportController) ProcessReportText(c echo.Context) error {
	var req models.ReportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" || req.Latitude == "" || req.Longitude == "" || req.CrimeName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required fields: name, latitude, longitude, crime_name",
		})
	}

	report, err := ctrl.svc.ProcessReportText(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process report",
		})
	}

	return c.JSON(http.StatusCreated, report)
}