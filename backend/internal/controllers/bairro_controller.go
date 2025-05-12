package controllers

import (
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/AloysioLvy/TccRadarCampinas/backend/internal/services"
)

// BairroController agrupa as rotas e a lógica de negócio
// relacionadas à entidade Bairro (áreas geográficas).
type BairroController struct {
    // svc é a interface de serviço que expõe métodos
    // para recuperar dados de Bairros do banco.
    svc services.BairroService
}

// NewBairroController é a função fábrica que recebe uma
// implementação de BairroService e retorna um ponteiro
// para um BairroController configurado.
func NewBairroController(svc services.BairroService) *BairroController {
    return &BairroController{svc: svc}
}

// Register registra as rotas HTTP associadas a Bairros
// em um echo.Group, que já carrega o prefixo de rota
// (por exemplo "/api/v1").
func (ctr *BairroController) Register(g *echo.Group) {
    // GET /bairros -> chama o método GetBairros
    g.GET("/bairros", ctr.GetBairros)
}

// GetBairros é o handler para GET /bairros.
// - Extrai o contexto HTTP para passar ao serviço.
// - Chama o serviço para listar todos os bairros.
// - Em caso de erro, retorna 500 Internal Server Error.
// - Em sucesso, retorna 200 OK com o slice de bairros em JSON.
func (ctr *BairroController) GetBairros(c echo.Context) error {
    // 1. Chama o serviço para listar todos os bairros.
    bairros, err := ctr.svc.ListBairros(c.Request().Context())
    if err != nil {
        // 2. Se houver falha, retorna JSON com status 500.
        return c.JSON(
            http.StatusInternalServerError,
            map[string]string{"error": err.Error()},
        )
    }

    // 3. Se tudo correr bem, retorna JSON com status 200.
    return c.JSON(http.StatusOK, bairros)
}