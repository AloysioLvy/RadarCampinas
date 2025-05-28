package controllers

import (
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
    "github.com/AloysioLvy/TccRadarCampinas/backend/internal/services"
)

// DenunciaController é responsável por agrupar as rotas e
// lógica de negócio relacionadas a denúncias de crimes.
type DenunciaController struct {
    // svc é a interface de serviço que fornece métodos
    // para criar e manipular denúncias no banco de dados.
    svc services.DenunciaService
}

// NewDenunciaController é uma função fábrica que recebe
// uma implementação de DenunciaService e retorna um
// ponteiro para um DenunciaController configurado.
func NewDenunciaController(svc services.DenunciaService) *DenunciaController {
    return &DenunciaController{svc: svc}
}


// Register associa as rotas HTTP de denúncias a este controller.
// Recebe um *echo.Group, que normalmente representa um prefixo de rota
// (por exemplo, "/api/v1").
func (ctr *DenunciaController) Register(g *echo.Group) {
    // Quando chegar um POST para "/denuncias", o Echo
    // chamará o método ctr.CreateDenuncia para tratar a requisição.
    g.POST("/denuncias", ctr.CreateDenuncia)

    // Processar denuncias em texto livre
    g.POST("/denuncias/texto", ctr.ReceberDenunciaTexto)
}

// CreateDenuncia trata requisições HTTP POST em "/denuncias".
// Espera um corpo JSON representando uma models.Denuncia.
// Se tudo ocorrer bem, persiste a denúncia no banco e
// retorna status 201 Created junto com o objeto criado.
func (ctr *DenunciaController) CreateDenuncia(c echo.Context) error {
    // 1. Cria uma instância vazia de models.Denuncia
    //    que receberá os dados do JSON.
    d := new(models.Denuncia)

    // 2. Popula 'd' a partir do corpo da requisição.
    //    c.Bind faz parsing do JSON e preenche os campos.
    //    Se falhar (JSON inválido, campos faltando), retorna erro 400.
    if err := c.Bind(d); err != nil {
        // Retorna Bad Request com a mensagem de erro em JSON.
        return c.JSON(
            http.StatusBadRequest,
            map[string]string{"error": err.Error()},
        )
    }

    // 3. Chama o serviço de denúncias para criar o registro
    //    no banco de dados, passando o contexto da requisição
    //    (útil para timeouts, tracing, etc.).
    if err := ctr.svc.CreateDenuncia(c.Request().Context(), d); err != nil {
        // Em caso de falha no banco ou na lógica, retorna
        // Internal Server Error com a mensagem.
        return c.JSON(
            http.StatusInternalServerError,
            map[string]string{"error": err.Error()},
        )
    }

    // 4. Se tudo ocorrer bem, retorna status 201 Created
    //    e o objeto 'd' preenchido (incluindo o ID gerado).
    return c.JSON(http.StatusCreated, d)
}

// Trata reqs HTTP POST para processar denuncias em texto livre
func (ctr *DenunciaController) ReceberDenunciaTexto(c echo.Context) error {
    // 1. Recebe o JSON da requisição
    req := new(models.DenunciaRequest)
    if err := c.Bind(req); err != nil {
        return c.JSON(
            http.StatusBadRequest,
            map[string]string{"error": "Formato da req invalido: " + err.Error()},
        )
    }

    // 2. Valida os campos obrigatorios
    if req.Latitude == "" || req.Longitude == "" || req.TipoDeCrime == "" || req.DataCrime == "" {
        return c.JSON(
            http.StatusBadRequest,
            map[string]string{"error": "Campos obrigatorios nao fornecidos"},
        )
    }

    // 3. Processa a denuncia usando o serviço
    denuncia, err := ctr.svc.ProcessarDenunciaTexto(c.Request().Context(), req)
    if err != nil {
        return c.JSON(
            http.StatusInternalServerError,
            map[string]string{"error": "Falha ao processar denuncia: " + err.Error()},
        )
    }

    // gerar id_denuncia pela model (orm - query)
    // alocar as linhas da table em seus respectivos lugares.

    // 4. Retorna a denuncia criada com stts 201
    return c.JSON(http.StatusCreated, denuncia)

    

}