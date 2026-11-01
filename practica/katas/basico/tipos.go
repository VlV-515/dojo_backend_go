// Package basico reúne las katas del módulo 02 (Go fundamentos).
//
// Cómo se usa:
//
//   - Este archivo define los tipos y errores que comparten los ejercicios. NO lo edites.
//   - ejercicios.go trae las funciones vacías: ahí escribes tu solución.
//   - basico_test.go dice qué se espera. Corre `go test ./basico/ -v` y ve
//     cómo cada prueba pasa de FAIL a PASS.
//   - solucion.go tiene la solución del sensei; solo compila con `-tags solucion`.
//     Intenta primero. Si te atoras, pide una pista con /pista en Claude.
package basico

import "errors"

// Errores centinela del inventario. Se comparan con errors.Is, nunca por texto.
var (
	ErrCantidadInvalida = errors.New("la cantidad debe ser mayor a cero")
	ErrSinExistencia    = errors.New("no hay existencia suficiente")
	ErrClaveInvalida    = errors.New("la clave no tiene el formato PREFIJO-NUMERO")
)

// Persona es una persona del ERP reducida a lo mínimo.
type Persona struct {
	Nombre   string
	Apellido string
	Edad     int
}

// Inventario guarda existencias por clave de producto.
//
// El map NO se inicializa solo: un Inventario{} tiene existencias == nil y
// escribir en un map nil provoca panic. Por eso existe NuevoInventario.
type Inventario struct {
	existencias map[string]int
}
