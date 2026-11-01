# 02 · Go fundamentos

> Cinturón amarillo · 2–3 días

## Objetivos

- Instalar y entender la herramienta `go` (módulos, build, test, fmt, vet).
- Dominar tipos, structs, slices, maps, funciones, métodos y punteros.
- Manejar errores al estilo Go (`error`, `errors.Is`, `fmt.Errorf("%w")`).
- Leer y escribir tests *table-driven*.

## 1. La herramienta `go`

```bash
go version                 # el dojo pide >= 1.24
go mod init ejemplo.com/x  # crea go.mod (un "módulo" = un proyecto)
go mod download            # baja dependencias declaradas en go.mod
go build ./...             # compila todo (./... = este paquete y los de abajo)
go run main.go             # compila y ejecuta
go test ./... -v           # corre tests
go vet ./...               # detecta errores comunes
gofmt -l .                 # lista archivos mal formateados (esperado: nada)
```

- **Módulo:** un proyecto con `go.mod`. En una arquitectura de microservicios cada servicio suele ser un módulo; en el dojo, `dojo-api` es `github.com/VlV-515/dojo_backend_go/practica/dojo-api`.
- **Paquete:** una carpeta. Todos los `.go` de una carpeta comparten `package`.
- **Exportado:** lo que empieza con **Mayúscula** (`NewAlmacenService`) se ve desde otros paquetes; lo que empieza con minúscula (`txOrDB`) no. Los dos ejemplos son de `dojo-api`.
- `go mod tidy` reescribe `go.mod` según lo que importa el código. En un proyecto que importa un paquete generado (como `docs/` de swag), **no** lo corras mientras ese paquete falte (lo verás en el módulo 05): borraría dependencias que sí se usan.

## 2. Tipos y variables

```go
var nombre string = "Ana"
edad := 30                 // inferencia, solo dentro de funciones
const MaxIntentos = 3

// Valor cero: toda variable nace con un valor útil
var n int      // 0
var s string   // ""
var b bool     // false
var p *int     // nil
var m map[string]int // nil  ← leer sí, escribir PANIC
```

El **valor cero** es clave en cualquier backend Go: GORM ignora campos en valor cero al hacer `Updates(struct)` (módulo 06), y un `map` nil truena al escribir.

## 3. Slices y maps

```go
nums := []int{1, 2, 3}
nums = append(nums, 4)            // append puede crear un arreglo nuevo: SIEMPRE reasigna
vacio := make([]int, 0, 10)       // largo 0, capacidad 10

for i, n := range nums { ... }

m := map[string]int{"a": 1}
m["b"]++                          // en map inicializado: ok
v, ok := m["z"]                   // ok=false si no existe; v=0
delete(m, "a")
```

**Detalle que verás en proyectos reales:** al serializar a JSON, un slice `nil` sale como `null` y uno vacío como `[]`. El front espera `[]`. Por eso los mappers hacen `make([]T, 0, len(x))` (mira [`mappers/almacen.go`](../practica/dojo-api/src/infrastructure/web/mappers/almacen.go) y el test `TestListaVaciaEsArreglo` de `dojo-api`).

## 4. Funciones, structs y métodos

```go
func dividir(a, b int) (int, error) {     // varios valores de retorno
    if b == 0 {
        return 0, errors.New("división entre cero")
    }
    return a / b, nil
}

type Persona struct {
    Nombre string
    Edad   int
}

func (p Persona) Saludo() string { return "Hola " + p.Nombre }   // receptor por valor: copia
func (p *Persona) Cumplir()      { p.Edad++ }                     // receptor puntero: modifica

func NuevaPersona(nombre string) *Persona {                      // "constructor" por convención
    return &Persona{Nombre: nombre}
}
```

**¿Valor o puntero?** Usa puntero cuando el método modifica el struct, cuando el struct es grande o cuando tiene un `sync.Mutex`. En `dojo-api` (y en casi cualquier backend con capas) los services y repositorios siempre son punteros (`*AlmacenServiceImpl`, `*AlmacenRepositoryImpl`).

## 5. Errores

En Go los errores son **valores** que se devuelven, no excepciones.

```go
var ErrSinExistencia = errors.New("no hay existencia")   // error centinela

func retirar(...) error {
    if hay < pedidas {
        return fmt.Errorf("clave %s: %w", clave, ErrSinExistencia)  // %w ENVUELVE
    }
    return nil
}

err := retirar(...)
if errors.Is(err, ErrSinExistencia) { ... }   // ve a través de la envoltura
```

- Siempre revisa `err` justo después de la llamada: `if err != nil { return ... }`.
- Nunca compares errores por texto (`err.Error() == "..."`).
- En muchas empresas los errores que cruzan capas pasan por la librería compartida interna (en el dojo, `pkg/apperr`) en vez de `fmt.Errorf`; lo verás en el módulo 09. La idea es la misma: añadir contexto sin perder el error original.

## 6. Tests

```go
func TestSaludo(t *testing.T) {
    casos := []struct {
        nombre, entra, espera string
    }{
        {"normal", "Ana", "Hola, Ana"},
        {"vacío", "", "Hola, desconocido"},
    }
    for _, c := range casos {
        t.Run(c.nombre, func(t *testing.T) {
            if got := Saludo(c.entra); got != c.espera {
                t.Errorf("Saludo(%q) = %q; se esperaba %q", c.entra, got, c.espera)
            }
        })
    }
}
```

- Archivos `*_test.go`, funciones `TestXxx(t *testing.T)`.
- `t.Errorf` marca falla y sigue; `t.Fatalf` marca falla y corta.
- *Table-driven* es el estilo idiomático de Go (mira [`src/domain/services/almacen_test.go`](../practica/dojo-api/src/domain/services/almacen_test.go) en `dojo-api`).

## En un proyecto real

Abre [`src/domain/models/almacen.go`](../practica/dojo-api/src/domain/models/almacen.go) y [`src/domain/services/almacen.go`](../practica/dojo-api/src/domain/services/almacen.go) de `dojo-api`. Identifica: un struct, una constante, un método con receptor puntero, un constructor `New...` y una función que devuelve `(valor, error)`.

## Ejercicios — katas `practica/katas/basico`

```bash
cd practica/katas
go test ./basico/ -v          # todo empieza en FAIL
```

Abre `basico/ejercicios.go` y resuelve en orden. Cada kata tiene su test en `basico_test.go`: **léelo**, el test es la especificación.

1. `Saludo` — strings y condicionales.
2. `SumarMontos`, `Filtrar` — slices, ciclos, funciones como valor.
3. `ContarPalabras` — maps.
4. `Persona.NombreCompleto`, `EsMayorDeEdad` — métodos.
5. `Inventario` — punteros, constructor, errores centinela y `%w`.
6. `ParsearClave` — varios retornos, `strconv`.

## Reto

Cuando todo esté en verde:

1. Cambia el receptor de `Agregar` a valor (`func (i Inventario)`) y corre los tests. ¿Qué pasa y por qué? Regrésalo.
2. Agrega a la tabla de `TestParsearClave` tres casos nuevos que se te ocurran (por ejemplo espacios o minúsculas mezcladas) y decide qué debe pasar.
3. Corre `go vet ./basico/` y `gofmt -l .`: ambos sin salida.

Revisa con `/revisar-reto`.

## Autoevaluación

1. ¿Qué diferencia hay entre `var m map[string]int` y `m := map[string]int{}`?
2. ¿Por qué `append` se reasigna (`s = append(s, x)`)?
3. ¿Qué hace `%w` y en qué se diferencia de `%v` en `fmt.Errorf`?
4. ¿Cuándo usarías receptor puntero?
5. ¿Qué significa que un identificador empiece con mayúscula?

**Listo cuando:** `go test ./basico/ -v` pasa completo y puedes explicar cada solución sin verla.

Siguiente: [03 · Go intermedio](03-go-intermedio.md)
