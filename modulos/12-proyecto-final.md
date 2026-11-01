# 12 · Proyecto final (cinturón negro)

> 3–5 días · Se presenta a un mentor (o a ti mismo, con `/examen` y `/revisar-reto`)

Dos partes. Las dos son obligatorias.

## Parte 1 — Levantar `dojo-api` completo como un servicio real

Levanta `practica/dojo-api` como levantarías un microservicio el primer día en un equipo: Postgres en Docker, config local, migraciones, Swagger y arranque verificado. **Tú** ejecutas los comandos; si usas a Claude, pídele que te guíe, no que lo haga.

| Paso | Evidencia que entregas |
|---|---|
| 1. Dependencias | `go mod download`, `swag init -g main.go`, `go build ./...` sin errores y `git status` limpio (`docs/` no aparece porque está ignorado) |
| 2. Base en Docker | `docker compose up -d dbpostgres` y `docker compose ps` con el contenedor `healthy` |
| 3. config.json | `cp config.example.json config.json`, ajustes (`persistencia: postgres`, password del compose) y `git check-ignore -v config.json`. Explica cada campo **sin pegar el password** |
| 4. Validación | `python3 -m json.tool config.json > /dev/null` sin error |
| 5. Migraciones | `go run main.go migrate status` (antes), `go run main.go migrate up`, `migrate status` (después) |
| 6. Red | `nc -vz 127.0.0.1 5433` y explicación de por qué desde tu máquina no es `dbpostgres:5432` |
| 7. Arranque | Línea `[arranque] dojo-api escuchando en :1099 (persistencia: postgres)`, `curl` a `/salud` y Swagger abierto en `http://localhost:1099/swagger/index.html` |
| 8. Todo en contenedores | `docker compose up -d --build`, `curl` a `/salud` y una ruta protegida con `Authorization: Bearer <token>` |

Reglas (las mismas que aplicarías contra una base compartida de `dev`/`staging`): `auto_migrate_on_boot: false` en tu `config.json` local, migraciones solo con el CLI y a conciencia, y ningún secreto en la evidencia.

**Bonus:** escribe un `.github/workflows/ci.yml` (módulo 11) que corra `swag init`, `gofmt`, `go vet` y `go test -race` sobre `dojo-api`.

## Parte 2 — Feature completa en `practica/dojo-api`: cajas

### Reto C (obligatorio)

Recurso `cajas` de punta a punta. Contrato en [`src/infrastructure/app/reto_caja_test.go`](../practica/dojo-api/src/infrastructure/app/reto_caja_test.go) y [`src/infrastructure/database/reto_caja_test.go`](../practica/dojo-api/src/infrastructure/database/reto_caja_test.go):

- `POST /cajas` `{"nombre","almacen_id"}` → 201 con `{id, nombre, almacen_id, abierta:false}`; nombre vacío → 400; almacén inexistente → 404.
- `GET /cajas?almacen_id=N` → 200 con lista (`[]` si no hay); sin `almacen_id` → 400.
- Todo con token.
- La tabla `caja` nace de una migración `0002` (no de `AutoMigrate`) con llave foránea `REFERENCES almacen`.

Debe incluir **todas** las capas (models, puertos `CajaReader`/`CajaWriter`, service con test de fake que valide el almacén usando `AlmacenReader`, adaptador memoria **y** postgres, migración `0002`, DTOs, mappers, controller con swagger, rutas después del middleware, cableado en `app.go`).

```bash
go test -tags reto_caja ./... -run Reto -v
go test -race -tags "reto_seguridad reto_desactivar reto_caja" ./...   # todo junto
```

### Extensiones (elige al menos dos)

1. **Abrir/cerrar caja**: `PUT /cajas/:id/abrir` y `/cerrar` con reglas (409 si ya está en ese estado; no se puede abrir una caja de un almacén inactivo). Usa `uow.Do`.
2. **Paginación** en `GET /almacenes?pagina=1&por_pagina=20` con total en la respuesta.
3. **Búsqueda** `GET /almacenes?q=cen` (insensible a mayúsculas) en ambos adaptadores, con `?` en SQL.
4. **Timeout** en todas las peticiones (middleware con `context.WithTimeout`) y test que lo demuestre.
5. **Test de integración** de `persistence/postgres` que haga `Skip` si no hay `DOJO_TEST_DSN`, y que corra contra el Postgres del compose cuando sí la hay.
6. **Inventario congelado de rutas**: test que falle si alguien agrega o quita una ruta sin actualizar la lista (generalización del reto A).

### Entregables

- Rama `feature/dojo-cajas-<tu-nombre>` en tu copia local (o en tu fork, si quieres mostrarlo).
- **PR simulado**: descripción con todas las secciones del módulo 11 (qué, por qué, cómo se probó, riesgos, pasos de despliegue: migración `0002`, permisos de rutas nuevas si hubiera autorización por rol). Si trabajas en un fork, abre el PR contra tu propio `main` y deja que corra tu CI.
- Corrida en Docker Compose (base + API) con `curl` de cada endpoint.
- `go vet ./...` y `gofmt -l .` sin salida.

## Rúbrica (la usa el mentor y `/revisar-reto`)

| Criterio | Peso |
|---|---|
| Tests del reto en verde, con `-race` | 20 % |
| Regla de dependencia y convenciones (nombres, capas, `var _`, constructores) | 20 % |
| Errores bien manejados (`Wrap`, `IsCode`, sin `New(...nil)`, sin `Wrap(nil)`) | 15 % |
| Persistencia correcta (migración, `Find`/`First`, `Updates` con map, `txOrDB`) | 15 % |
| Seguridad (rutas protegidas, sin secretos, entrada validada) | 10 % |
| Swagger y contrato JSON (`[]` y no `null`) | 10 % |
| PR y explicación oral | 10 % |

## Presentación (30 min)

1. Demo en vivo: `dojo-api` con cajas corriendo en compose contra Postgres, Swagger abierto y `curl` de cada endpoint.
2. Recorrido del código de cajas capa por capa.
3. Lectura del PR simulado: cómo lo probaste y qué harías para desplegarlo.
4. Preguntas sorpresa de cualquier módulo (`/examen` te ayuda a prepararte).

**Al aprobar:** tienes las bases para tomar tu primer ticket real en un equipo backend Go. Busca uno pequeño (un endpoint, un fix con test) y vuelve a este dojo cuando lo necesites.
