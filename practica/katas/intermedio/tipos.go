// Package intermedio reúne las katas del módulo 03 (Go intermedio): interfaces,
// errores con contexto, context, concurrencia y genéricos.
//
// Igual que en basico: tipos.go no se edita, ejercicios.go es tuyo, el
// test manda y solucion.go solo compila con `-tags solucion`.
//
// Corre SIEMPRE con el detector de carreras:
//
//	go test ./intermedio/ -race -v
package intermedio

import (
	"context"
	"errors"
	"sync"
)

// ---------------------------------------------------------------------------
// Interfaces
// ---------------------------------------------------------------------------

// Notificador es un PUERTO: el código de negocio depende de esta interfaz, no
// de RabbitMQ, correo o websocket. En dojo-api los puertos viven en
// src/domain/repositories y los adaptadores en src/infrastructure.
type Notificador interface {
	Notificar(ctx context.Context, destino, mensaje string) error
}

// Pedido es lo mínimo para el ejercicio de avisos.
type Pedido struct {
	Folio      string
	Cliente    string
	DiasAtraso int
}

// Aviso es lo que guarda NotificadorMemoria por cada notificación.
type Aviso struct {
	Destino string
	Mensaje string
}

// NotificadorMemoria es un adaptador falso: guarda los avisos en memoria. Es lo
// que en los tests se llama "fake" (ver almacenFalso en
// dojo-api/src/domain/services/almacen_test.go).
type NotificadorMemoria struct {
	Avisos []Aviso
	// FallarEn hace que la notificación número N (empezando en 1) falle.
	// 0 = nunca falla.
	FallarEn int
}

// ---------------------------------------------------------------------------
// Errores
// ---------------------------------------------------------------------------

// ErrNoEncontrado es un error centinela.
var ErrNoEncontrado = errors.New("registro no encontrado")

// ErrTiempoAgotado lo devuelve ConsultarConLimite cuando se acaba el tiempo.
var ErrTiempoAgotado = errors.New("tiempo agotado")

// ErrorHTTP es un error con código HTTP, parecido a apperr.AppError de
// dojo-api. DIFERENCIA IMPORTANTE: este sí implementa Unwrap; AppError no.
type ErrorHTTP struct {
	Codigo  int
	Mensaje string
	Causa   error
}

// ---------------------------------------------------------------------------
// Concurrencia
// ---------------------------------------------------------------------------

// ContadorSeguro cuenta por clave y se puede usar desde varias goroutines.
type ContadorSeguro struct {
	mu     sync.Mutex
	conteo map[string]int
}
