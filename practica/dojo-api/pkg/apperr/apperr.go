// Package apperr imita, en pequeño, la librería de errores compartida que suelen
// tener las empresas con varios microservicios (un módulo Go privado que todos
// importan).
//
// Trae A PROPÓSITO tres trampas reales de ese tipo de librerías, para que aprendas
// a detectarlas antes de que te muerdan en un proyecto de verdad:
//
//   - AppError NO implementa Unwrap: errors.Is/As no ven la causa. Para decidir
//     por tipo de error se usa IsCode (por código HTTP).
//   - Wrap(nil) NO devuelve nil. `return x, apperr.Wrap(err)` con err == nil
//     convierte un éxito en error.
//   - New(code, msg, nil) revienta al responder: Response llama
//     DevelopMessage.Error() sobre un nil. Siempre pasa una causa.
//
// No las "arregles": el módulo 09 explica cómo convivir con ellas y por qué
// cambiar una librería compartida es más caro de lo que parece.
package apperr

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError es un error con código HTTP, mensaje para el usuario y causa técnica.
type AppError struct {
	Code           int
	Message        string
	DevelopMessage error
}

func (e *AppError) Error() string {
	return e.Message
}

// New crea un error con código HTTP. causa debe ser distinta de nil.
func New(code int, msg string, causa error) error {
	return &AppError{Code: code, Message: msg, DevelopMessage: causa}
}

// Wrap conserva un *AppError tal cual y convierte cualquier otro error en un 500.
// Ojo: Wrap(nil) devuelve un error, no nil.
func Wrap(err error) error {
	var ke *AppError
	if errors.As(err, &ke) {
		return ke
	}
	if err == nil {
		return &AppError{Code: http.StatusInternalServerError, Message: "Error desconocido"}
	}
	return &AppError{Code: http.StatusInternalServerError, Message: "Error interno", DevelopMessage: err}
}

// Code devuelve el código HTTP del error (500 si no es un AppError).
func Code(err error) int {
	var ke *AppError
	if errors.As(err, &ke) {
		return ke.Code
	}
	return http.StatusInternalServerError
}

// IsCode indica si err es un AppError con ese código.
func IsCode(err error, code int) bool {
	var ke *AppError
	return errors.As(err, &ke) && ke.Code == code
}

// Response escribe el error como JSON. Úsalo con errores que vienen del service.
func Response(err error, ctx *gin.Context) {
	ke := Wrap(err).(*AppError)
	ctx.AbortWithStatusJSON(ke.Code, gin.H{
		"codigo":  ke.Code,
		"mensaje": ke.Message,
		"detalle": ke.DevelopMessage.Error(), // panic si DevelopMessage es nil (trampa a propósito)
	})
}

// CtrlResponse arma y escribe el error en un solo paso. Úsalo en el controller
// para errores propios del borde HTTP (JSON mal formado, id inválido...).
func CtrlResponse(code int, msg string, causa error, ctx *gin.Context) {
	Response(New(code, msg, causa), ctx)
}
