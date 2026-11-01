//go:build !solucion

package intermedio

import (
	"context"
	"time"
)

// ---------------------------------------------------------------------------
// Kata 1. Implementar una interfaz
// ---------------------------------------------------------------------------

// Notificar guarda el aviso en n.Avisos. Si esta es la llamada número
// n.FallarEn, devuelve un error y NO guarda nada.
// Pregunta: ¿qué pasa si el receptor no es puntero? Pruébalo.
func (n *NotificadorMemoria) Notificar(ctx context.Context, destino, mensaje string) error {
	// TODO: tu código aquí
	return nil
}

// Esta línea no compila si *NotificadorMemoria deja de cumplir Notificador.
// Es un truco común en servicios: `var _ AlmacenService = (*AlmacenServiceImpl)(nil)`.
var _ Notificador = (*NotificadorMemoria)(nil)

// ---------------------------------------------------------------------------
// Kata 2. Depender de la interfaz, no del adaptador
// ---------------------------------------------------------------------------

// AvisarAtrasos notifica al cliente de cada pedido con MÁS de 3 días de atraso.
// El mensaje es "Tu pedido <folio> lleva <n> días de atraso".
// Devuelve cuántos avisos se enviaron. Al primer error se detiene y devuelve
// los enviados hasta ahí más el error envuelto con el folio:
// "pedido <folio>: <error original>" (usa %w).
func AvisarAtrasos(ctx context.Context, n Notificador, pedidos []Pedido) (int, error) {
	// TODO: tu código aquí
	return 0, nil
}

// ---------------------------------------------------------------------------
// Kata 3. Errores propios, errors.Is y errors.As
// ---------------------------------------------------------------------------

// Error devuelve "<Mensaje>: <Causa>" o solo el Mensaje si no hay Causa.
func (e *ErrorHTTP) Error() string {
	// TODO: tu código aquí
	return ""
}

// Unwrap expone la causa para que errors.Is/As puedan ver a través de ErrorHTTP.
func (e *ErrorHTTP) Unwrap() error {
	// TODO: tu código aquí
	return nil
}

// CodigoDe saca el código HTTP de cualquier error:
//   - nil -> 200
//   - si en la cadena hay un *ErrorHTTP -> su Codigo (usa errors.As)
//   - si en la cadena está ErrNoEncontrado -> 404 (usa errors.Is)
//   - cualquier otro -> 500
func CodigoDe(err error) int {
	// TODO: tu código aquí
	return 0
}

// ---------------------------------------------------------------------------
// Kata 4. context con límite de tiempo
// ---------------------------------------------------------------------------

// ConsultarConLimite corre consulta con un límite de tiempo d.
//   - Si consulta termina a tiempo, devuelve su resultado tal cual.
//   - Si se acaba el tiempo, devuelve ("", error que envuelve ErrTiempoAgotado)
//     SIN esperar a que consulta termine.
//
// Pista: context.WithTimeout, una goroutine, un canal con buffer 1 y select.
// ¿Por qué buffer 1? Si no, la goroutine se queda bloqueada para siempre (fuga).
// Aplica a cualquier cliente HTTP o SDK que no traiga timeout propio.
func ConsultarConLimite(ctx context.Context, d time.Duration, consulta func(ctx context.Context) (string, error)) (string, error) {
	// TODO: tu código aquí
	return "", nil
}

// ---------------------------------------------------------------------------
// Kata 5. Concurrencia segura con Mutex
// ---------------------------------------------------------------------------

// NuevoContador crea un ContadorSeguro listo para usarse.
func NuevoContador() *ContadorSeguro {
	// TODO: tu código aquí
	return &ContadorSeguro{}
}

// Incrementar suma 1 a clave. Debe ser seguro llamarlo desde muchas goroutines.
func (c *ContadorSeguro) Incrementar(clave string) {
	// TODO: tu código aquí
}

// Valor devuelve el conteo de clave.
func (c *ContadorSeguro) Valor(clave string) int {
	// TODO: tu código aquí
	return 0
}

// ---------------------------------------------------------------------------
// Kata 6. Worker pool
// ---------------------------------------------------------------------------

// ProcesarEnParalelo aplica fn a cada id usando como máximo `workers`
// goroutines a la vez y devuelve un map id -> resultado.
// Si algún fn falla, devuelve (nil, ese error). Basta con devolver el primero
// que veas.
// Pista: sync.WaitGroup + un canal de trabajos + un Mutex para el map.
// Nota: en servicios reales conviene envolver `go func(){...}()` en un helper que
// recupere pánicos y propague el contexto. Aquí no, porque estás aprendiendo cómo funciona por dentro.
func ProcesarEnParalelo(ctx context.Context, ids []int, workers int, fn func(ctx context.Context, id int) (int, error)) (map[int]int, error) {
	// TODO: tu código aquí
	return nil, nil
}

// ---------------------------------------------------------------------------
// Kata 7. Genéricos
// ---------------------------------------------------------------------------

// Mapear aplica f a cada elemento y devuelve un slice nuevo del mismo largo.
func Mapear[T, U any](xs []T, f func(T) U) []U {
	// TODO: tu código aquí
	return nil
}

// Agrupar arma un map clave -> elementos, conservando el orden de llegada.
func Agrupar[T any, K comparable](xs []T, clave func(T) K) map[K][]T {
	// TODO: tu código aquí
	return nil
}
