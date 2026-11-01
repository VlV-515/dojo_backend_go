package intermedio

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNotificadorMemoria(t *testing.T) {
	n := &NotificadorMemoria{FallarEn: 2}
	ctx := context.Background()

	if err := n.Notificar(ctx, "ana", "uno"); err != nil {
		t.Fatalf("primera notificación falló: %v", err)
	}
	if err := n.Notificar(ctx, "ana", "dos"); err == nil {
		t.Fatal("la segunda notificación debía fallar (FallarEn: 2)")
	}
	if len(n.Avisos) != 1 {
		t.Fatalf("se guardaron %d avisos; se esperaba 1 (la que falla no se guarda)", len(n.Avisos))
	}
	if n.Avisos[0] != (Aviso{"ana", "uno"}) {
		t.Errorf("aviso guardado = %+v", n.Avisos[0])
	}
}

func TestAvisarAtrasos(t *testing.T) {
	pedidos := []Pedido{
		{"P-1", "ana", 1},
		{"P-2", "luis", 5},
		{"P-3", "eva", 3},
		{"P-4", "ana", 10},
	}

	t.Run("todo bien", func(t *testing.T) {
		n := &NotificadorMemoria{}
		enviados, err := AvisarAtrasos(context.Background(), n, pedidos)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if enviados != 2 {
			t.Errorf("enviados = %d; se esperaba 2 (solo más de 3 días)", enviados)
		}
		espera := []Aviso{
			{"luis", "Tu pedido P-2 lleva 5 días de atraso"},
			{"ana", "Tu pedido P-4 lleva 10 días de atraso"},
		}
		if !reflect.DeepEqual(n.Avisos, espera) {
			t.Errorf("avisos = %+v; se esperaba %+v", n.Avisos, espera)
		}
	})

	t.Run("falla el segundo aviso", func(t *testing.T) {
		n := &NotificadorMemoria{FallarEn: 2}
		enviados, err := AvisarAtrasos(context.Background(), n, pedidos)
		if err == nil {
			t.Fatal("se esperaba error")
		}
		if enviados != 1 {
			t.Errorf("enviados = %d; se esperaba 1", enviados)
		}
		if got := err.Error(); len(got) < 10 || got[:10] != "pedido P-4" {
			t.Errorf("el error %q debe empezar con \"pedido P-4\"", got)
		}
		if errors.Unwrap(err) == nil {
			t.Error("el error debe envolver al original con %w")
		}
	})
}

func TestErrorHTTP(t *testing.T) {
	causa := errors.New("conexión rechazada")
	e := &ErrorHTTP{Codigo: 503, Mensaje: "Servicio no disponible", Causa: causa}
	if got := e.Error(); got != "Servicio no disponible: conexión rechazada" {
		t.Errorf("Error() = %q", got)
	}
	if got := (&ErrorHTTP{Codigo: 400, Mensaje: "Datos inválidos"}).Error(); got != "Datos inválidos" {
		t.Errorf("Error() sin causa = %q", got)
	}
	if !errors.Is(e, causa) {
		t.Error("errors.Is no ve la causa: falta Unwrap")
	}
}

func TestCodigoDe(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		espera int
	}{
		{"nil", nil, 200},
		{"ErrorHTTP directo", &ErrorHTTP{Codigo: 409, Mensaje: "duplicado"}, 409},
		{"ErrorHTTP envuelto", fmt.Errorf("guardando: %w", &ErrorHTTP{Codigo: 400, Mensaje: "x"}), 400},
		{"no encontrado", ErrNoEncontrado, 404},
		{"no encontrado envuelto", fmt.Errorf("almacén 7: %w", ErrNoEncontrado), 404},
		{"ErrorHTTP gana sobre su causa", &ErrorHTTP{Codigo: 422, Mensaje: "x", Causa: ErrNoEncontrado}, 422},
		{"cualquier otro", errors.New("boom"), 500},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := CodigoDe(c.err); got != c.espera {
				t.Errorf("CodigoDe = %d; se esperaba %d", got, c.espera)
			}
		})
	}
}

func TestConsultarConLimite(t *testing.T) {
	t.Run("a tiempo", func(t *testing.T) {
		got, err := ConsultarConLimite(context.Background(), time.Second, func(ctx context.Context) (string, error) {
			return "ok", nil
		})
		if err != nil || got != "ok" {
			t.Errorf("= (%q, %v); se esperaba (\"ok\", nil)", got, err)
		}
	})

	t.Run("error de la consulta pasa tal cual", func(t *testing.T) {
		boom := errors.New("boom")
		_, err := ConsultarConLimite(context.Background(), time.Second, func(ctx context.Context) (string, error) {
			return "", boom
		})
		if !errors.Is(err, boom) {
			t.Errorf("err = %v; se esperaba boom", err)
		}
	})

	t.Run("se agota", func(t *testing.T) {
		inicio := time.Now()
		_, err := ConsultarConLimite(context.Background(), 50*time.Millisecond, func(ctx context.Context) (string, error) {
			time.Sleep(2 * time.Second)
			return "tarde", nil
		})
		if !errors.Is(err, ErrTiempoAgotado) {
			t.Errorf("err = %v; se esperaba ErrTiempoAgotado", err)
		}
		if time.Since(inicio) > time.Second {
			t.Error("ConsultarConLimite esperó a la consulta lenta; debe regresar al vencer el límite")
		}
	})
}

func TestContadorSeguro(t *testing.T) {
	c := NuevoContador()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.Incrementar("pedidos")
			}
		}()
	}
	wg.Wait()
	if got := c.Valor("pedidos"); got != 10000 {
		t.Errorf("Valor = %d; se esperaba 10000 (¿te falta el Mutex? corre con -race)", got)
	}
	if got := c.Valor("nada"); got != 0 {
		t.Errorf("Valor de clave nueva = %d; se esperaba 0", got)
	}
}

func TestProcesarEnParalelo(t *testing.T) {
	t.Run("todos bien y respeta el límite de workers", func(t *testing.T) {
		var activos, maxActivos int32
		fn := func(ctx context.Context, id int) (int, error) {
			n := atomic.AddInt32(&activos, 1)
			for {
				m := atomic.LoadInt32(&maxActivos)
				if n <= m || atomic.CompareAndSwapInt32(&maxActivos, m, n) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&activos, -1)
			return id * id, nil
		}
		ids := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		got, err := ProcesarEnParalelo(context.Background(), ids, 3, fn)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if len(got) != len(ids) {
			t.Fatalf("resultados = %d; se esperaban %d", len(got), len(ids))
		}
		for _, id := range ids {
			if got[id] != id*id {
				t.Errorf("resultado[%d] = %d; se esperaba %d", id, got[id], id*id)
			}
		}
		if maxActivos > 3 {
			t.Errorf("hubo %d goroutines a la vez; el máximo era 3", maxActivos)
		}
		if maxActivos < 2 {
			t.Errorf("solo hubo %d goroutine a la vez: no estás procesando en paralelo", maxActivos)
		}
	})

	t.Run("un error", func(t *testing.T) {
		boom := errors.New("boom")
		got, err := ProcesarEnParalelo(context.Background(), []int{1, 2, 3}, 2, func(ctx context.Context, id int) (int, error) {
			if id == 2 {
				return 0, boom
			}
			return id, nil
		})
		if !errors.Is(err, boom) {
			t.Errorf("err = %v; se esperaba boom", err)
		}
		if got != nil {
			t.Errorf("con error el resultado debe ser nil; fue %v", got)
		}
	})
}

func TestMapear(t *testing.T) {
	got := Mapear([]int{1, 2, 3}, func(n int) string { return fmt.Sprintf("#%d", n) })
	if !reflect.DeepEqual(got, []string{"#1", "#2", "#3"}) {
		t.Errorf("Mapear = %v", got)
	}
	if got := Mapear([]int{}, func(n int) int { return n }); got == nil || len(got) != 0 {
		t.Errorf("Mapear de vacío = %v; se esperaba slice vacío, no nil", got)
	}
}

func TestAgrupar(t *testing.T) {
	pedidos := []Pedido{{"P-1", "ana", 1}, {"P-2", "luis", 5}, {"P-3", "ana", 3}}
	got := Agrupar(pedidos, func(p Pedido) string { return p.Cliente })
	espera := map[string][]Pedido{
		"ana":  {{"P-1", "ana", 1}, {"P-3", "ana", 3}},
		"luis": {{"P-2", "luis", 5}},
	}
	if !reflect.DeepEqual(got, espera) {
		t.Errorf("Agrupar = %+v; se esperaba %+v", got, espera)
	}
}
