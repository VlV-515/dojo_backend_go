# 03 · Go intermedio: interfaces, errores, context, concurrencia, genéricos

> Cinturón amarillo · 3–4 días

## Objetivos

- Diseñar con **interfaces** pequeñas (la base de la arquitectura hexagonal).
- Crear errores propios y navegarlos con `errors.Is` / `errors.As`.
- Usar `context.Context` para cancelar y poner límites de tiempo.
- Escribir código concurrente **seguro**: goroutines, canales, `sync.WaitGroup`, `sync.Mutex`, `-race`.
- Usar genéricos donde aportan.
- Probar con *fakes* (dobles de prueba) y `httptest`.

## 1. Interfaces

Una interfaz describe **qué** se puede hacer, no **cómo**. En Go se cumplen de forma implícita: si tu tipo tiene los métodos, cumple la interfaz.

```go
type Notificador interface {
    Notificar(ctx context.Context, destino, mensaje string) error
}

type NotificadorRabbit struct{ ... }    // adaptador real
type NotificadorMemoria struct{ ... }   // adaptador de prueba

func AvisarAtrasos(ctx context.Context, n Notificador, ...) // depende de la interfaz
```

**Por qué importa en un proyecto real:** en `dojo-api` todo el dominio depende de interfaces (`AlmacenReader`, `AlmacenWriter`, `UnitOfWork`, en [`src/domain/repositories/`](../practica/dojo-api/src/domain/repositories/almacen.go)). Así el service no sabe si detrás hay Postgres o un fake en memoria, y se puede probar sin base.

Reglas que conviene seguir (y que sigue `dojo-api`):

- Interfaces **chicas**: `Reader` y `Writer` separados.
- Verificación en compilación: `var _ AlmacenService = (*AlmacenServiceImpl)(nil)`.

## 2. Errores propios

```go
type ErrorHTTP struct {
    Codigo int
    Mensaje string
    Causa error
}
func (e *ErrorHTTP) Error() string { ... }
func (e *ErrorHTTP) Unwrap() error { return e.Causa }

var e *ErrorHTTP
if errors.As(err, &e) { e.Codigo ... }   // busca un TIPO en la cadena
if errors.Is(err, ErrNoEncontrado) { ... }  // busca un VALOR en la cadena
```

> **Trampa típica de librerías internas:** muchas librerías de errores compartidas no implementan `Unwrap()`. El dojo trae una a propósito: `apperr.AppError` ([`pkg/apperr`](../practica/dojo-api/pkg/apperr/apperr.go)). Con ella `errors.Is` no ve la causa, así que para decidir se usa `apperr.IsCode(err, http.StatusNotFound)`. Lo practicas en el módulo 09.

## 3. context.Context

`context` viaja como **primer parámetro** de toda función que hace I/O. Sirve para:

- **Cancelar**: si el cliente HTTP se va, se cancela la consulta.
- **Límites de tiempo**: `context.WithTimeout(ctx, 3*time.Second)`.
- **Valores de petición**: un `TraceID` para correlacionar logs o la transacción del `UnitOfWork` (`txOrDB` la saca del ctx; lo verás en el módulo 06).

```go
ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()                 // SIEMPRE, o fugas recursos
select {
case r := <-resultado:
case <-ctx.Done():             // se venció o cancelaron
    return ctx.Err()
}
```

En un controller Gin el contexto de la petición es `ctx.Request.Context()`.

## 4. Concurrencia

```go
go hacerAlgo()                       // goroutine: función en paralelo

ch := make(chan int, 1)              // canal con buffer 1
ch <- 5; v := <-ch

var wg sync.WaitGroup
wg.Add(1); go func(){ defer wg.Done(); ... }(); wg.Wait()

var mu sync.Mutex
mu.Lock(); defer mu.Unlock()         // protege datos compartidos
```

- Un `map` escrito desde varias goroutines sin `Mutex` es una **carrera de datos**: Go puede tronar con `fatal error: concurrent map writes`.
- Corre los tests con `-race` para detectarlas.
- **Gin atiende cada petición en su propia goroutine.** Todo lo que compartan (un repositorio en memoria, un caché) debe ser seguro.

> **En un proyecto real:** muchos equipos prohíben el `go func(){...}()` suelto en los servicios. En su lugar usan un *helper* propio (algo como `EnSegundoPlano(ctx, fn)` o `EnSegundoPlanoConTimeout`) que propaga el TraceID, pone un límite de tiempo y **recupera pánicos**: un `panic` en una goroutine sin `recover` tumba el proceso completo, aunque Gin tenga `Recovery()`. Aquí lo haces a mano para entender qué hay debajo.

## 5. Genéricos

```go
func Mapear[T, U any](xs []T, f func(T) U) []U { ... }
func Agrupar[T any, K comparable](xs []T, clave func(T) K) map[K][]T { ... }
```

Úsalos para utilerías de colecciones. No los fuerces en el dominio.

## 6. Probar con fakes y httptest

- **Fake:** implementación en memoria de una interfaz que además **registra** lo que pasó (`escrituras []string`). Ejemplo: `almacenFalso` y `uowFalso` en [`src/domain/services/almacen_test.go`](../practica/dojo-api/src/domain/services/almacen_test.go) de `dojo-api`.
- **httptest:** prueba controllers y rutas sin abrir puertos:

```go
req := httptest.NewRequest(http.MethodGet, "/almacenes", nil)
rec := httptest.NewRecorder()
router.ServeHTTP(rec, req)
// rec.Code, rec.Body
```

Lo usarás mucho en `dojo-api` ([`src/infrastructure/app/app_test.go`](../practica/dojo-api/src/infrastructure/app/app_test.go)).

## Ejercicios — katas `practica/katas/intermedio`

```bash
cd practica/katas
go test ./intermedio/ -race -v
```

1. `NotificadorMemoria.Notificar` — implementar una interfaz.
2. `AvisarAtrasos` — depender de la interfaz; envolver errores con contexto.
3. `ErrorHTTP` + `CodigoDe` — `Error()`, `Unwrap()`, `errors.As`, `errors.Is`.
4. `ConsultarConLimite` — `context.WithTimeout`, goroutine, canal con buffer, `select`.
5. `ContadorSeguro` — `sync.Mutex`.
6. `ProcesarEnParalelo` — *worker pool* con límite de goroutines.
7. `Mapear` y `Agrupar` — genéricos.

## Reto

1. En `ConsultarConLimite`, cambia el canal a **sin buffer** (`make(chan resultado)`). Los tests siguen pasando… ¿qué problema introdujiste? (pista: ¿qué le pasa a la goroutine cuando se vence el tiempo?).
2. Quita el `Mutex` de `ContadorSeguro` y corre con `-race`. Copia el primer renglón del reporte y explica qué dice.
3. Escribe un test nuevo para `ProcesarEnParalelo` con `workers = 0` y decide qué debe pasar.

## Autoevaluación

1. ¿Por qué el service de `dojo-api` recibe `AlmacenReader` (interfaz) y no `*postgres.AlmacenRepositoryImpl`?
2. ¿Qué diferencia hay entre `errors.Is` y `errors.As`?
3. ¿Por qué siempre `defer cancel()`?
4. ¿Qué detecta `-race` y por qué no lo detecta siempre?
5. ¿Para qué sirve `var _ Interfaz = (*Tipo)(nil)`?

**Listo cuando:** `go test ./intermedio/ -race -v` pasa y puedes explicar por qué el canal lleva buffer 1.

Siguiente: [04 · APIs HTTP con Gin](04-http-con-gin.md)
