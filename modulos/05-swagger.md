# 05 · Swagger con swaggo

> Cinturón naranja · 1 día

## Objetivos

- Entender qué es OpenAPI/Swagger y para qué lo usa el front.
- Escribir anotaciones `swag` en handlers y DTOs.
- Entender por qué un clon limpio de un servicio con swag **no compila** hasta correr `swag init`.

## 1. Qué es

Swagger (OpenAPI) es un documento JSON/YAML que describe la API: rutas, parámetros, cuerpos y respuestas. Con él:

- El equipo de front sabe qué mandar y qué recibir sin leer Go.
- Hay una UI (`/swagger/index.html`) para probar endpoints desde el navegador.

**swaggo** (`swag`) genera ese documento leyendo **comentarios** en tu código.

## 2. Anotaciones

En `main.go` (información general):

```go
// @title                      Dojo API
// @version                    1.0
// @BasePath                   /
// @securityDefinitions.apikey BearerToken
// @in                         header
// @name                       Authorization
```

En cada handler:

```go
// Alta godoc
// @Summary     Registra un almacén
// @Tags        almacenes
// @Accept      json
// @Produce     json
// @Security    BearerToken
// @Param       almacen body     dtos.AlmacenAltaDTO true "Almacén"
// @Success     201     {object} dtos.RespuestaAlmacen
// @Failure     400     {object} dtos.RespuestaError
// @Router      /almacenes [post]
```

En los DTOs, `example:"..."` llena los ejemplos.

> **Seguridad:** los `example:` **nunca** llevan valores reales. Es un descuido más común de lo que parece: alguien pega una llave real de un proveedor (Auth0, Stripe, AWS…) como ejemplo en un DTO, y queda publicada en el Swagger **y** en el historial de git. Borrarla en un commit nuevo no basta: hay que **rotarla**.

## 3. El paquete `docs/` y por qué el clon limpio no compila

```bash
swag init -g main.go      # lee anotaciones y genera docs/docs.go, swagger.json, swagger.yaml
```

`main.go` importa ese paquete:

```go
_ "github.com/VlV-515/dojo_backend_go/practica/dojo-api/docs"     // en tu servicio: _ "<ruta-del-módulo>/docs"
```

Pero `docs/` está en `.gitignore` (es generado; mira [`.gitignore`](../practica/dojo-api/.gitignore) de `dojo-api`). Resultado: en un clon recién hecho, `go build` falla con algo como `package <módulo>/docs is not in std` o `no required module provides package`. **No es que falte una dependencia**: falta generar.

Orden correcto en un clon nuevo:

1. `go mod download` (no `go mod tidy`: al no existir `docs/`, tidy cree que sobran dependencias y rompe `go.mod`).
2. `go install github.com/swaggo/swag/cmd/swag@<versión de go.mod>` — misma versión que el repo (en `dojo-api`: `@v1.16.3`).
3. `swag init -g <main>.go`.
4. `go build ./...`.

Tras modificar handlers o DTOs, **regenera** para no desalinear el contrato.

## Ejercicios

1. Borra `practica/dojo-api/docs/` y corre `go build ./...`. Copia el error exacto. Luego `swag init -g main.go` y vuelve a compilar.
2. Arranca el servicio y abre `http://localhost:1099/swagger/index.html`. Autentícate con el botón **Authorize** (`Bearer dojo-local-inseguro`) y crea un almacén desde la UI.
3. Agrega a `GetAll` la anotación `@Failure 500 {object} dtos.RespuestaError`, regenera y verifica que aparece.
4. ¿Qué versión de `swag` pide [`go.mod`](../practica/dojo-api/go.mod) de `dojo-api`? Compárala con `swag --version`. ¿Qué podría cambiar en `docs/` si generas con otra versión?

## Reto

`/swagger/*any` se registra en [`main.go`](../practica/dojo-api/main.go) **antes** de `ArmarRouter`. ¿Qué pasaría si lo movieras después? Predícelo, pruébalo con `curl -i` y regrésalo.

## Autoevaluación

1. ¿Por qué `docs/` no se versiona?
2. ¿Por qué no se corre `go mod tidy` antes de `swag init`?
3. ¿Dónde se declaran los esquemas de seguridad y dónde se aplican?
4. Cambiaste un DTO y no regeneraste. ¿Qué se rompe y para quién?

**Listo cuando:** puedes dejar compilando un clon limpio de `dojo-api` (sin `docs/`) sin mirar la guía.

Siguiente: [06 · PostgreSQL, GORM y migraciones](06-postgres-gorm-migraciones.md)
