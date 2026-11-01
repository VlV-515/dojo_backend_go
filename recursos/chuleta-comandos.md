# Chuleta de comandos

## Go

```bash
go build ./...            # compila todo
go vet ./...              # análisis estático (en cero)
gofmt -l .                # archivos mal formateados (esperado: nada)
go test ./... -v          # tests con detalle
go test ./... -race       # con detector de carreras
go test ./paquete/ -run TestNombre -v
go test -tags reto_caja ./...      # incluir archivos con //go:build reto_caja
go mod download           # bajar dependencias (no hagas tidy si falta docs/ de swag)
go env GOPRIVATE GOMODCACHE
go env -w GOPRIVATE='github.com/tu-org/*'   # módulos privados sin proxy público
```

## Swagger

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.3   # misma versión que go.mod
swag init -g main.go                                  # genera docs/ (no se versiona)
```

## Git

```bash
git switch -c feature/1234-desactivar-almacen develop
git commit -m "feat: desactivar almacén (#1234)"
git check-ignore -v config.json      # confirma que el secreto no se versiona
git status --ignored
git log --oneline -10
git diff develop...HEAD --stat       # qué lleva tu rama contra la base
```

## Red y diagnóstico

```bash
nc -vz -w 5 127.0.0.1 5433           # ¿el puerto responde?
lsof -i :1099                        # ¿quién ocupa el puerto?
python3 -m json.tool config.json     # ¿el JSON es válido?
```

## Docker

```bash
docker compose up -d [servicio]
docker compose up -d --build
docker compose ps
docker compose logs -f api
docker compose exec dbpostgres psql -U dojo
docker compose down            # conserva volúmenes
docker compose down -v         # BORRA volúmenes (solo práctica)
```

## psql

```bash
psql "host=127.0.0.1 port=5433 user=dojo dbname=dojo sslmode=disable"
```

```sql
\dt                 -- tablas
\d almacen          -- columnas, índices y llaves
SELECT * FROM schema_migrations;   -- migraciones aplicadas por dojo-api
\q
```

## dojo-api

```bash
go run main.go                         # memoria, :1099
DOJO_TOKEN=otro go run main.go
CONFIG=./config.json go run main.go migrate status   # solo lee
go run main.go migrate up              # solo contra TU Postgres de práctica
curl -s localhost:1099/salud           # pública
curl -s -H "Authorization: Bearer dojo-local-inseguro" localhost:1099/almacenes
curl -s -H "Authorization: Bearer dojo-local-inseguro" localhost:1099/almacenes/1
curl -s -X POST -H "Authorization: Bearer dojo-local-inseguro" \
     -H "Content-Type: application/json" -d '{"clave":"ALM-02","nombre":"Almacén Norte"}' \
     localhost:1099/almacenes          # solo en local
open http://localhost:1099/swagger/index.html
```
