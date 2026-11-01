// Package middleware tiene los middlewares de gin. Autenticar es una versión
// simplificada de un middleware de auth real (que además validaría permisos, p. ej.
// con casbin).
package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/pkg/apperr"
)

const msgNoAutorizado = "No autorizado"

// Autenticar exige `Authorization: Bearer <token>`.
//
// Recuerda: en gin un router.Use solo alcanza a las rutas que se registran
// DESPUÉS. Todo lo que se registre antes queda público.
func Autenticar(token string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// strings.Cut no truena si el header no trae espacio. Un
		// strings.Split(...)[1] ingenuo sí entraría en pánico en ese caso.
		esquema, valor, ok := strings.Cut(ctx.GetHeader("Authorization"), " ")
		if !ok || esquema != "Bearer" || valor == "" {
			apperr.CtrlResponse(http.StatusUnauthorized, msgNoAutorizado, errors.New("falta el header Authorization: Bearer <token>"), ctx)
			return
		}
		// Comparación en tiempo constante: no filtra cuántos caracteres coinciden.
		if subtle.ConstantTimeCompare([]byte(valor), []byte(token)) != 1 {
			apperr.CtrlResponse(http.StatusUnauthorized, msgNoAutorizado, errors.New("token inválido"), ctx)
			return
		}
		ctx.Next()
	}
}
