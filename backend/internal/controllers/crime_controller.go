package controllers

import (
	"net/http"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/services"
	"github.com/labstack/echo/v4"
)


type CrimeController struct {
    // svc is interface of service with methods to received data Crimes of database
    svc services.CrimeService
}

func NewCrimeController(svc services.CrimeService) *CrimeController {
	return &CrimeController{svc: svc}
}

func (ctr *CrimeController) Register(g *echo.Group) {
	// GET /crimes -> chama o método GetCrimes
	g.GET("/crimes", ctr.GetCrimes)
}


func (ctr *CrimeController) GetCrimes(c echo.Context) error {
    crimes, err := ctr.svc.ListCrimes(c.Request().Context())
    if err != nil {
        return c.JSON(
            http.StatusInternalServerError,
            map[string]string{"error": err.Error()},
        )
    }

    
    return c.JSON(http.StatusOK, crimes)
}
