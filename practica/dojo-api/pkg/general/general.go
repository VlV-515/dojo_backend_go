// Package general define la forma estándar de responder éxito.
package general

import "github.com/gin-gonic/gin"

// Respuesta es la envoltura JSON de toda respuesta exitosa.
type Respuesta struct {
	Codigo  int    `json:"codigo" example:"200"`
	Mensaje string `json:"mensaje" example:"Consulta exitosa"`
	Datos   any    `json:"datos"`
}

// Responder escribe {codigo, mensaje, datos} con el código HTTP dado.
func Responder(code int, msg string, datos any, ctx *gin.Context) {
	ctx.JSON(code, Respuesta{Codigo: code, Mensaje: msg, Datos: datos})
}
