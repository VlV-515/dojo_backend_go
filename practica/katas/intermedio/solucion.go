//go:build solucion

package intermedio

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func (n *NotificadorMemoria) Notificar(ctx context.Context, destino, mensaje string) error {
	llamada := len(n.Avisos) + 1
	if n.FallarEn != 0 && llamada == n.FallarEn {
		return errors.New("el canal de notificaciones no responde")
	}
	n.Avisos = append(n.Avisos, Aviso{Destino: destino, Mensaje: mensaje})
	return nil
}

var _ Notificador = (*NotificadorMemoria)(nil)

func AvisarAtrasos(ctx context.Context, n Notificador, pedidos []Pedido) (int, error) {
	enviados := 0
	for _, p := range pedidos {
		if p.DiasAtraso <= 3 {
			continue
		}
		msg := fmt.Sprintf("Tu pedido %s lleva %d días de atraso", p.Folio, p.DiasAtraso)
		if err := n.Notificar(ctx, p.Cliente, msg); err != nil {
			return enviados, fmt.Errorf("pedido %s: %w", p.Folio, err)
		}
		enviados++
	}
	return enviados, nil
}

func (e *ErrorHTTP) Error() string {
	if e.Causa == nil {
		return e.Mensaje
	}
	return e.Mensaje + ": " + e.Causa.Error()
}

func (e *ErrorHTTP) Unwrap() error {
	return e.Causa
}

func CodigoDe(err error) int {
	if err == nil {
		return 200
	}
	var errHTTP *ErrorHTTP
	if errors.As(err, &errHTTP) {
		return errHTTP.Codigo
	}
	if errors.Is(err, ErrNoEncontrado) {
		return 404
	}
	return 500
}

func ConsultarConLimite(ctx context.Context, d time.Duration, consulta func(ctx context.Context) (string, error)) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	type resultado struct {
		valor string
		err   error
	}
	canal := make(chan resultado, 1)
	go func() {
		valor, err := consulta(ctx)
		canal <- resultado{valor, err}
	}()

	select {
	case r := <-canal:
		return r.valor, r.err
	case <-ctx.Done():
		return "", fmt.Errorf("consulta tras %s: %w", d, ErrTiempoAgotado)
	}
}

func NuevoContador() *ContadorSeguro {
	return &ContadorSeguro{conteo: make(map[string]int)}
}

func (c *ContadorSeguro) Incrementar(clave string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conteo[clave]++
}

func (c *ContadorSeguro) Valor(clave string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conteo[clave]
}

func ProcesarEnParalelo(ctx context.Context, ids []int, workers int, fn func(ctx context.Context, id int) (int, error)) (map[int]int, error) {
	if workers < 1 {
		workers = 1
	}
	trabajos := make(chan int)
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		resultado = make(map[int]int, len(ids))
		primerErr error
	)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range trabajos {
				valor, err := fn(ctx, id)
				mu.Lock()
				if err != nil {
					if primerErr == nil {
						primerErr = err
					}
				} else {
					resultado[id] = valor
				}
				mu.Unlock()
			}
		}()
	}

	for _, id := range ids {
		trabajos <- id
	}
	close(trabajos)
	wg.Wait()

	if primerErr != nil {
		return nil, primerErr
	}
	return resultado, nil
}

func Mapear[T, U any](xs []T, f func(T) U) []U {
	resultado := make([]U, len(xs))
	for i, x := range xs {
		resultado[i] = f(x)
	}
	return resultado
}

func Agrupar[T any, K comparable](xs []T, clave func(T) K) map[K][]T {
	grupos := make(map[K][]T)
	for _, x := range xs {
		k := clave(x)
		grupos[k] = append(grupos[k], x)
	}
	return grupos
}
