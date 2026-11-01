# Soluciones (para mentores)

> Si eres alumno: llegar aquí antes de intentar te quita justo lo que viniste a aprender. Usa `/pista` primero.

## Katas

Las soluciones viven junto a los ejercicios, en `solucion.go`, con build tag `solucion`:

```bash
cd practica/katas
go test -race -tags solucion ./... -v
```

## Retos de dojo-api

`dojo-api-retos.patch` resuelve los retos A, B y C sobre una copia limpia:

```bash
# desde la raíz del repo dojo_backend_go/
mkdir -p /tmp/dojo-sol && cp -R practica/dojo-api /tmp/dojo-sol/dojo-api
patch -p1 -d /tmp/dojo-sol < soluciones/dojo-api-retos.patch
cd /tmp/dojo-sol/dojo-api
swag init -g main.go
go test -race -tags "reto_seguridad reto_desactivar reto_caja" ./...
```

Qué contiene:

- **Reto A:** `routes.RutasReportes` se mueve después de `middleware.Autenticar` en `app.go`.
- **Reto B:** `Desactivar` en la interfaz `AlmacenService` y su `Impl` (404 propagado, 409 si ya inactivo, `Actualizar` una vez), handler `Desactiva` con swagger, ruta `PUT /almacenes/:id/desactivar` en el grupo protegido.
- **Reto C:** `models.Caja`, `CajaReader`/`CajaWriter`, `CajaService` (valida nombre y que el almacén exista vía `AlmacenReader`), adaptadores memoria y postgres, migración `0002_caja.up.sql` con `REFERENCES almacen (id)`, DTOs, mappers, `CajaController`, `RutasCajas` y cableado en `app.go`.

No es la única solución válida. Al revisar, prioriza la rúbrica del módulo 12 sobre el parecido con el patch.

## Verificación completa del dojo (antes de compartir cambios)

```bash
# Katas: sin solución fallan, con solución pasan
cd practica/katas && go vet ./... && go vet -tags solucion ./... && go test -race -tags solucion ./...

# dojo-api: base en verde y retos en rojo
cd ../dojo-api && swag init -g main.go && gofmt -l . && go vet ./... && go test -race ./...
go test -tags reto_seguridad ./src/infrastructure/app/ -run Reto   # debe FALLAR
```

Después aplica el patch en una copia y confirma que los tres retos pasan con `-race`.
