# 04 · APIs HTTP con Gin

> Cinturón naranja · 2–3 días

## Objetivos

- Entender HTTP lo suficiente para diseñar y depurar una API: métodos, códigos, headers, JSON.
- Construir rutas, grupos, handlers y middlewares con **Gin**.
- Validar entrada con `ShouldBindJSON` y tags `binding`.
- Entender por qué **el orden de registro de rutas es load-bearing** (sostiene la seguridad) en Gin.
- Probar endpoints con `curl` y con `httptest`.

## 1. HTTP en 5 minutos

| Método | Uso típico |
|---|---|
| `GET` | Consultar (no cambia datos) |
| `POST` | Crear |
| `PUT` | Modificar (muchas APIs también lo usan para "eliminar lógico", p. ej. `PUT /almacenes/1/desactivar`, en vez de `DELETE`) |
| `DELETE` | Borrar |

| Código | Significado |
|---|---|
| 200 / 201 | OK / creado |
| 400 | Datos inválidos (culpa del cliente) |
| 401 | No autenticado (sin token o token inválido) |
| 403 | Autenticado pero **sin permiso** (p. ej. una política de roles con [casbin](https://casbin.org)) |
| 404 | No existe |
| 409 | Conflicto (duplicado, estado no válido) |
| 500 | Error nuestro |
| 503 | Dependencia caída (Rabbit, otro servicio) |

```bash
curl -s localhost:1099/salud
curl -s -H "Authorization: Bearer dojo-local-inseguro" localhost:1099/almacenes
curl -s -X POST -H "Content-Type: application/json" \
     -H "Authorization: Bearer dojo-local-inseguro" \
     -d '{"clave":"alm-01","nombre":"Centro"}' localhost:1099/almacenes
```

## 2. Gin: lo esencial

```go
r := gin.New()
r.Use(gin.Logger(), gin.Recovery())       // middlewares globales

r.GET("/salud", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"estado": "ok"})
})

g := r.Group("almacenes")                 // prefijo común
g.GET("", ctrl.GetAll)                    // GET /almacenes
g.GET("/:id", ctrl.GetByID)               // c.Param("id")
g.POST("", ctrl.Alta)                     // c.ShouldBindJSON(&dto)
// c.Query("activos") lee ?activos=true

r.Run(":1099")
```

### Validación de entrada

```go
type AlmacenAltaDTO struct {
    Clave  string `json:"clave" binding:"required"`
    Nombre string `json:"nombre" binding:"required"`
}
var req AlmacenAltaDTO
if err := c.ShouldBindJSON(&req); err != nil { /* 400 */ }
```

`binding:"required"` rechaza el campo ausente o vacío, pero **no** `"   "`. Las reglas de negocio (recortar, mayúsculas, duplicados) van en el **service**, no en el controller.

## 3. Middlewares y el orden que importa

Un middleware es un handler que corre antes (y/o después) del handler final. Llama `c.Next()` para seguir o `c.Abort()` para cortar.

> **En Gin, `r.Use(mw)` solo afecta a las rutas registradas DESPUÉS.**

Por eso [`main.go`](../practica/dojo-api/main.go) y `ArmarRouter` en [`src/infrastructure/app/app.go`](../practica/dojo-api/src/infrastructure/app/app.go) siguen este orden (simplificado):

```go
router.GET("/swagger/*any", ...)              // público (main.go, ANTES de ArmarRouter)
routes.RutasSalud(router, saludCtrl)          // /salud: pública a propósito
router.Use(middleware.Autenticar(cfg.Token))  // ← middleware de autenticación
routes.RutasAlmacenes(router, almacenCtrl)    // todo lo demás: protegido
```

Una ruta registrada antes de `Autenticar` queda **sin control de acceso**. En un servicio real, este middleware además consulta permisos por rol (casbin u otra herramienta) y las rutas se registran en el sistema de permisos al arrancar. La defensa más barata es un test que congele el inventario de rutas y verifique que toda ruta privada responde 401 sin token (eso es justo lo que hace el reto A).

### Pánicos

`gin.Recovery()` (o el handler de pánicos propio que traiga la librería del equipo) convierte un `panic` en 500 en vez de tumbar el servicio. Ejemplo clásico: un middleware de auth que lee el token con `strings.Split(header, " ")[1]` entra en pánico si el header `Authorization` no trae espacio, y responde 500 en vez de 401. Por eso [`middleware/auth.go`](../practica/dojo-api/src/infrastructure/web/middleware/auth.go) usa `strings.Cut`.

## 4. Anatomía de un handler

Así se ve `Alta` en [`controllers/almacen.go`](../practica/dojo-api/src/infrastructure/web/controllers/almacen.go):

```go
func (c *AlmacenController) Alta(ctx *gin.Context) {
    var req dtos.AlmacenAltaDTO
    if err := ctx.ShouldBindJSON(&req); err != nil {
        apperr.CtrlResponse(http.StatusBadRequest, msgLeerDatos, err, ctx)   // error del borde
        return
    }
    almacen := mappers.AltaToAlmacen(req)                                    // DTO -> modelo
    if err := c.almacenService.Registrar(ctx.Request.Context(), &almacen); err != nil {
        apperr.Response(err, ctx)                                            // error del service
        return
    }
    general.Responder(http.StatusCreated, "Almacén registrado", mappers.ToAlmacenDTO(almacen), ctx)
}
```

Leer → mapear → llamar al service → responder. **Nada de reglas de negocio ni SQL aquí.**

## Ejercicios — `practica/dojo-api`

```bash
cd practica/dojo-api
swag init -g main.go     # (el módulo 05 explica por qué hace falta)
go run main.go           # sin config.json: arranca en memoria
```

1. Prueba con `curl` los 5 endpoints de almacenes (lista, consulta, alta, cambio, salud). Provoca a propósito un 400, un 401, un 404 y un 409.
2. Lee `src/infrastructure/web/middleware/auth.go`. ¿Por qué usa `strings.Cut` y `subtle.ConstantTimeCompare`?
3. Lee `src/infrastructure/app/app_test.go` y agrega un caso a la tabla de `TestFlujoAlmacenes`: `PUT /almacenes/1` con nombre vacío debe dar 400.
4. Corre `go test ./...`.

## Reto A — Cazador de rutas expuestas

Alguien dejó una ruta sensible **pública por accidente**. Encuéntrala y corrígela.

```bash
go test -tags reto_seguridad ./src/infrastructure/app/ -run Reto -v
```

Antes de leer el test: intenta encontrarla tú con `curl` y leyendo `app.go`. Después del arreglo, explica en el PR (imaginario) por qué el bug pasó la revisión.

## Autoevaluación

1. ¿Diferencia entre 401 y 403? ¿Qué capa produce cada uno en un servicio con autenticación + permisos por rol?
2. ¿Por qué una ruta registrada antes de `router.Use(middleware.Autenticar(...))` no pasa por la autenticación?
3. ¿Qué validación va en el DTO y cuál en el service?
4. ¿Por qué `apperr.CtrlResponse` en un caso y `apperr.Response` en otro?
5. ¿Qué hace `gin.Recovery()` y por qué un 500 por pánico es un bug aunque "no se cayó"?

**Listo cuando:** el reto A pasa y puedes explicar el orden de `main.go` + `ArmarRouter` línea por línea.

Siguiente: [05 · Swagger con swaggo](05-swagger.md)
