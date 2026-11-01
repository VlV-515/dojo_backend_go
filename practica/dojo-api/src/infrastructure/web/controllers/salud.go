package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SaludController responde si el servicio está vivo. Es pública a propósito:
// la usan balanceadores y healthchecks, que no traen token.
type SaludController struct {
	persistencia string
}

func NewSaludController(persistencia string) *SaludController {
	return &SaludController{persistencia: persistencia}
}

// Get godoc
// @Summary     Estado del servicio (pública)
// @Tags        salud
// @Produce     json
// @Success     200 {object} map[string]string
// @Router      /salud [get]
func (c *SaludController) Get(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"estado": "ok", "persistencia": c.persistencia})
}
