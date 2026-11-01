# Práctica

Dos módulos Go independientes (cada uno con su `go.mod`, igual que cada microservicio en un repo propio).

## `katas/` — módulos 02 y 03

```bash
cd katas
go test ./basico/ -v                 # módulo 02
go test ./intermedio/ -race -v       # módulo 03
```

- `tipos.go`: no se edita.
- `ejercicios.go`: aquí resuelves (funciones con `// TODO`).
- `*_test.go`: la especificación. Léela.
- `solucion.go`: referencia del sensei. Solo compila con `-tags solucion` (los build tags `//go:build` eligen qué archivos entran a la compilación). Intenta primero.

Es normal que `go test ./...` salga en rojo al principio: así empiezan las katas.

## `dojo-api/` — módulos 04 a 12

Mini microservicio con arquitectura hexagonal: dominio (models, repositories, services), infraestructura (persistence memoria/postgres, web, database, app), `apperr` que simula una librería de errores compartida, migraciones embebidas, swagger y Docker.

```bash
cd dojo-api
swag init -g main.go          # genera docs/ (no se versiona; sin esto no compila)
go test ./...                 # la base está en verde
go run main.go                # sin config.json: memoria, puerto 1099
```

Token por default en local: `dojo-local-inseguro` (cámbialo con `DOJO_TOKEN`).

### Retos (tests con build tags)

| Reto | Módulo | Comando |
|---|---|---|
| A · Cazador de rutas expuestas | 04 | `go test -tags reto_seguridad ./src/infrastructure/app/ -run Reto -v` |
| B · Desactivar almacenes | 08 | `go test -tags reto_desactivar ./... -run Reto -v` |
| C · Cajas de punta a punta | 12 | `go test -tags reto_caja ./... -run Reto -v` |

Sin el tag, esos tests no se compilan y el proyecto queda en verde. Con el tag, empiezan en rojo hasta que los resuelves.

### Con Postgres

```bash
docker compose up -d dbpostgres
cp config.example.json config.json   # cambia persistencia a "postgres" y password a "dojo"
go run main.go migrate status
go run main.go migrate up
go run main.go
```

O todo en contenedores: `docker compose up -d --build`.
