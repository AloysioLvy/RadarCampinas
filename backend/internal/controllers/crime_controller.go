package controllers

import (
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/AloysioLvy/TccRadarCampinas/backend/internal/services"
)

// CrimeController agrupa as rotas e a lógica de negócio
// relacionadas à entidade Crime (tipo de crime).
type CrimeController struct {
    // svc é a interface de serviço que expõe métodos
    // para recuperar dados de Crimes do banco.
    svc services.CrimeService
}

// NewCrimeController é a função fábrica que recebe uma
// implementação de CrimeService e retorna um ponteiro
// para um CrimeController configurado.
func NewCrimeController(svc services.CrimeService) *CrimeController {
    return &CrimeController{svc: svc}
}

// Register registra as rotas HTTP associadas a Crimes
// em um echo.Group, que já carrega o prefixo de rota
// (por exemplo "/api/v1").
func (ctr *CrimeController) Register(g *echo.Group) {
    // GET /crimes -> chama o método GetCrimes
    g.GET("/crimes", ctr.GetCrimes)
}

// GetCrimes é o handler para GET /crimes.
// - Obtém o contexto HTTP via c.Request().Context(),
//   útil para controle de timeout e tracing.
// - Chama o serviço para buscar todos os crimes.
// - Em caso de erro, retorna 500 Internal Server Error.
// - Em sucesso, retorna 200 OK com o slice de crimes em JSON.
func (ctr *CrimeController) GetCrimes(c echo.Context) error {
    // 1. Chama o serviço para listar todos os crimes.
    crimes, err := ctr.svc.ListCrimes(c.Request().Context())
    if err != nil {
        // 2. Se houver falha, retorna JSON com status 500.
        return c.JSON(
            http.StatusInternalServerError,
            map[string]string{"error": err.Error()},
        )
    }

    // 3. Se tudo correr bem, retorna JSON com status 200.
    return c.JSON(http.StatusOK, crimes)
}