# 08 · Arquitectura hexagonal

> Cinturón azul · 3–4 días

## Objetivos

- Explicar capas, puertos, adaptadores y la **regla de dependencia**.
- Ubicar cualquier archivo de `dojo-api` en su capa y saber qué puede importar.
- Agregar una operación nueva respetando la arquitectura y probándola con fakes.
- Seguir convenciones de nombres consistentes, como en un equipo real.

## 1. La idea

```
        ┌────────────────── infrastructure ──────────────────┐
        │  web (gin)      persistence (gorm)   external, messaging, search
        │      │                  ▲                     ▲
        │      ▼                  │ implementa          │ implementa
        │  ┌────────────── domain ─┴─────────────────────┴──┐
        │  │ services  ──usa──►  repositories (PUERTOS)     │
        │  │     └── models                                  │
        │  └─────────────────────────────────────────────────┘
        │                app.go = composition root (cablea todo)
        └─────────────────────────────────────────────────────┘
```

- **Dominio** (`src/domain/`): modelos, puertos (interfaces) y services. **No** importa gin, gorm ni nada de `infrastructure`.
- **Infraestructura** (`src/infrastructure/`): adaptadores que implementan los puertos (GORM, HTTP a otros servicios, colas de mensajes) y el borde HTTP (controllers, dtos, mappers, routes).
- **Regla de dependencia:** las flechas apuntan **hacia adentro**. El dominio no sabe que existe Gin.
- **Composition root** ([`app.go`](../practica/dojo-api/src/infrastructure/app/app.go)): único lugar donde se crean adaptadores y se inyectan a services y controllers.

**¿Por qué?** Puedes probar la lógica sin base (fakes), cambiar un adaptador sin tocar negocio, y evitas el código legacy típico (controllers con SQL, transacciones a mano) que provoca incidentes como el `Rollback()` sobre el `*gorm.DB` raíz que tumba todas las consultas siguientes del servicio (lo viste en el módulo 06; está comentado en [`postgres/unitOfWork.go`](../practica/dojo-api/src/infrastructure/persistence/postgres/unitOfWork.go)).

## 2. Mapa de `dojo-api` (`practica/dojo-api/src/`)

| Carpeta | Qué va | Ejemplo |
|---|---|---|
| `domain/models` | Entidades | [`almacen.go`](../practica/dojo-api/src/domain/models/almacen.go) |
| `domain/repositories` | Puertos `...Reader` / `...Writer`, `UnitOfWork` | [`almacen.go`](../practica/dojo-api/src/domain/repositories/almacen.go), [`unitOfWork.go`](../practica/dojo-api/src/domain/repositories/unitOfWork.go) |
| `domain/services` | Interfaz + `Impl` con la lógica | [`almacen.go`](../practica/dojo-api/src/domain/services/almacen.go) |
| `infrastructure/persistence/memoria` | Adaptador en memoria (default, sin base) | [`almacen.go`](../practica/dojo-api/src/infrastructure/persistence/memoria/almacen.go) |
| `infrastructure/persistence/postgres` | Adaptador GORM, `txOrDB` | [`almacen.go`](../practica/dojo-api/src/infrastructure/persistence/postgres/almacen.go) |
| `infrastructure/database` | Migraciones embebidas y CLI `migrate` | [`migrator.go`](../practica/dojo-api/src/infrastructure/database/migrator.go) |
| `infrastructure/web/{controllers,dtos,mappers,routes}` | Borde HTTP | `almacen.go` en cada una |
| `infrastructure/web/middleware` | Autenticación | [`auth.go`](../practica/dojo-api/src/infrastructure/web/middleware/auth.go) |
| `infrastructure/app/app.go` | Composition root | `New`, `ArmarRouter` |
| `architecture_test.go` | Hace cumplir la regla de dependencia | [`architecture_test.go`](../practica/dojo-api/src/architecture_test.go) |

En un microservicio de producción verás además carpetas como `infrastructure/external` (clientes HTTP a otros servicios, proveedores de identidad, autorización con casbin), `infrastructure/messaging` (RabbitMQ, `LISTEN` de Postgres, websockets) o `infrastructure/search` (Elasticsearch). Todas siguen la misma regla: implementan puertos del dominio y se cablean en `app.go`.

## 3. Convenciones

- **Idioma mixto por diseño:** inglés para el rol (`Reader`, `Service`, `Impl`, `UnitOfWork`), español para el negocio (`Almacen`, `PorID`, `Insertar`). Los nombres de negocio **no se renombran** a la ligera: son contrato JSON, rutas y columnas.
- **Un concepto = mismo nombre de archivo en todas las capas** (`almacen.go` en models, repositories, services, persistence, controllers, dtos, mappers, routes).
- Constructores `NewXxx` reciben dependencias por parámetro (DI por constructor).
- `var _ Puerto = (*Impl)(nil)` en cada implementación: si el adaptador deja de cumplir el puerto, no compila.
- Mensajes al usuario en constantes (`msgAlmacenDuplicado`).

## 4. Receta: agregar una operación

1. **Puerto** (si hace falta un método nuevo de lectura/escritura) en `repositories/`.
2. **Service**: método en la interfaz + implementación en `Impl`. Valida, decide, usa `uow.Do` si hay más de una escritura.
3. **Test del service** con fake (sin base).
4. **Adaptadores**: memoria y postgres.
5. **Borde**: DTO → mapper → handler (con swagger) → ruta (**después** del middleware).
6. **Cableado** en `app.go` si hay dependencias nuevas.
7. `swag init -g main.go`, `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...`, y prueba manual con `curl`.

## Ejercicios

1. Elige el concepto `almacen` y abre su archivo en cada capa (usa la tabla de la sección 2). Escribe en una frase qué hace cada una.
2. Lee `practica/dojo-api/src/architecture_test.go`. Agrega en `services/almacen.go` un import de `github.com/gin-gonic/gin` (usado en algo trivial) y corre `go test ./src/`. Lee el error y regrésalo.
3. En `services/almacen_test.go`, agrega un caso a `TestAlmacenRegistrar` donde el reader falle (usa el campo `fallar` del fake o agrégale uno a `PorClave`) y verifica que el error llega como 500.

## Reto B — Desactivar almacenes

Implementa `PUT /almacenes/:id/desactivar` siguiendo la receta completa. El contrato está en los tests [`services/reto_desactivar_test.go`](../practica/dojo-api/src/domain/services/reto_desactivar_test.go) y [`app/reto_desactivar_test.go`](../practica/dojo-api/src/infrastructure/app/reto_desactivar_test.go).

- 200 si estaba activo; 409 si ya estaba inactivo (sin escribir); 404 si no existe; 400 si el id no es válido; 401 sin token.
- Debe funcionar con persistencia en memoria **y** postgres.

```bash
go test -tags reto_desactivar ./src/domain/services/ -run Reto -v   # parte 1: service
go test -tags reto_desactivar ./... -run Reto -v                    # parte 2: endpoint
```

Luego pruébalo con `curl` contra Postgres (compose del módulo 07) y revisa en `psql` que `activo` quedó en `false`. Pide revisión con `/revisar-reto`.

## Autoevaluación

1. ¿Qué puede importar `domain/services` y qué no?
2. ¿Por qué los puertos se separan en `Reader` y `Writer`?
3. ¿Dónde se decide si se usa memoria o Postgres? ¿Por qué solo ahí?
4. ¿Qué ganas probando el service con un fake en vez de con Postgres? ¿Qué pierdes?
5. ¿Por qué un handler no debe hacer `db.Where(...)`?

**Listo cuando:** el reto B pasa, la revisión no encuentra violaciones de capas y puedes navegar `dojo-api` por capas sin perderte.

Siguiente: [09 · Librerías compartidas y módulos privados](09-librerias-compartidas.md)
