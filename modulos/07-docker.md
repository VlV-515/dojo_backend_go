# 07 · Docker y Docker Compose

> Cinturón verde · 2 días

## Objetivos

- Entender imagen, contenedor, volumen, red y puerto publicado.
- Leer y escribir un `Dockerfile` (multi-stage) y un `docker-compose.yml`.
- Entender **la diferencia entre el host dentro de docker y el host desde tu máquina**, que es el error número uno al correr un microservicio en local contra una base que vive en Docker o en un servidor.
- Leer el `docker-compose.yml` de `dojo-api` como leerías el del repo de infraestructura de un equipo.

## 1. Conceptos

| Concepto | Analogía | Comando |
|---|---|---|
| **Imagen** | Plantilla de solo lectura (un "instalador") | `docker images`, `docker build -t x .` |
| **Contenedor** | Una instancia corriendo de la imagen | `docker ps`, `docker run`, `docker logs -f` |
| **Volumen** | Disco que sobrevive al contenedor | `docker volume ls` |
| **Red** | LAN privada entre contenedores, con DNS por nombre/alias | `docker network ls` |
| **Puerto publicado** | Agujero de tu máquina hacia el contenedor | `ports: "127.0.0.1:5433:5432"` |

`ports: "5433:5432"` se lee **HOST:CONTENEDOR**: el 5432 del contenedor se ve como 5433 en la máquina que corre Docker (tu Mac en el dojo, el servidor en un entorno real).

Documentación oficial: [docs.docker.com](https://docs.docker.com/).

## 2. Dos caminos a la misma base

```
            Tu máquina                       Docker (local o servidor)
  go run ──► 127.0.0.1:5433 ──puerto──►  ┌── contenedor dbpostgres :5432 ──┐
                                         │       red "dojonet"             │
                         contenedor api ─┴──► dbpostgres:5432 (alias DNS)  │
                                         └──────────────────────────────────┘
```

- **Dentro** de la red docker: alias y puerto interno (`dbpostgres:5432`; en un sistema con más servicios verías cosas como `rabbit:5672` o `http://cajas`).
- **Fuera** (tu máquina): lo publicado (`127.0.0.1:5433`; contra un servidor real sería algo como `db.staging.tu-empresa.com:5432`).

Por eso un config pensado para correr **dentro** de docker no sirve tal cual cuando corres el servicio con `go run`: trae alias que tu máquina no resuelve. Compara los dos configs del dojo:

| Archivo | Para qué | Host de Postgres |
|---|---|---|
| [`docker/configs/config_dojo_api.json`](../practica/dojo-api/docker/configs/config_dojo_api.json) | La API corriendo en el compose | `dbpostgres:5432` |
| [`config.example.json`](../practica/dojo-api/config.example.json) | La API corriendo con `go run` en tu máquina | `127.0.0.1:5433` |

En un equipo real pasa lo mismo: el **repo de infraestructura** (compose + configs por entorno) trae los alias de la red docker del servidor, y para correr en local tienes que traducirlos a hosts y puertos que tu máquina sí alcanza.

## 3. Dockerfile

[`practica/dojo-api/Dockerfile`](../practica/dojo-api/Dockerfile) es **multi-stage**:

1. Etapa `builder` (`golang:1.24-bookworm`): baja dependencias (capa cacheada), instala `swag`, genera `docs/`, compila con `CGO_ENABLED=0`.
2. Etapa final (`debian:12-slim`): solo `ca-certificates` + el binario. Sin Go, sin código.

Muchos equipos hacen otra variante: el binario se compila **fuera** (en el pipeline de CI) y la imagen solo lo copia con `COPY`/`ADD`. Las dos son válidas; lo importante es que la imagen final sea chica y no traiga el código ni el toolchain. Y en ambos casos **la imagen no trae `config.json`**: lo provee quien la corre (montado como volumen, indicado con la variable `CONFIG`).

## 4. Docker Compose

```bash
docker compose up -d dbpostgres      # levantar un servicio
docker compose up -d --build         # todos, reconstruyendo imágenes
docker compose ps
docker compose logs -f api
docker compose exec dbpostgres psql -U dojo
docker compose down                  # apagar (los datos del volumen quedan)
docker compose down -v               # apagar y BORRAR volúmenes (solo en tu práctica)
```

Puntos de [`practica/dojo-api/docker-compose.yml`](../practica/dojo-api/docker-compose.yml) que verás en cualquier compose de un equipo serio:

- `image` con versión fija (`postgres:16-alpine`), nunca `latest`.
- `environment` para lo que cambia por entorno (`CONFIG`, `DOJO_TOKEN`).
- `volumes`: datos en un volumen con nombre y el config montado `:ro` (solo lectura).
- `networks` con `aliases`: así los demás servicios encuentran la base por nombre.
- `depends_on` + `healthcheck`: la API no arranca hasta que Postgres responde.
- Puertos atados a `127.0.0.1` para no exponerlos a la red de tu oficina o cafetería.

## 5. En un proyecto real

- **Base local vs base compartida.** Un patrón muy común es levantar en tu máquina un Postgres (y lo que haga falta: Mongo, Redis…) con la **misma versión** que el servidor y cargarle una copia de datos, para trabajar 100 % local sin tocar las bases compartidas de `dev`/`staging`. El compose del dojo es exactamente eso, en chico.
- **Los dumps con datos reales nunca se comparten ni se suben** a un repo: traen información de personas. Si necesitas datos de prueba versionados, usa seeds sintéticos.
- **El repo de infraestructura** del equipo suele tener un `docker-compose.yml` con todos los servicios y una carpeta de configs por entorno. Saber leerlo (qué puerto publica cada servicio, en qué red está, qué config monta) te ahorra horas de "¿por qué no conecta?".

## Ejercicios

Abre Docker Desktop (o el daemon de Docker que uses) primero.

1. `docker compose up -d dbpostgres` en `practica/dojo-api`. Con `docker ps` identifica nombre, imagen y puertos.
2. Conéctate a la base de **dos** formas: `psql` desde tu máquina por `127.0.0.1:5433` y `docker compose exec dbpostgres psql -U dojo`. ¿Qué host y puerto usó cada una?
3. `docker compose up -d --build`. Abre `http://localhost:1099/salud`: debe decir `postgres`. Lee `docker/configs/config_dojo_api.json`: ¿por qué ahí el host es `dbpostgres` y el puerto 5432?
4. `docker compose logs -f api`: encuentra la línea de migraciones aplicadas al arrancar. ¿Por qué ahí sí está `auto_migrate_on_boot: true` y contra una base compartida de `dev`/`staging` debe ser `false`?
5. Lee el `docker-compose.yml` completo y dibuja en papel: servicios, red, alias, volúmenes y puertos publicados. ¿Qué pasaría si quitas el `aliases: [dbpostgres]`?

```bash
docker network ls
docker network inspect dojo-api_dojonet    # el nombre real puede variar: búscalo en la lista
```

## Reto

Haz que la API del compose escuche en tu máquina en el puerto **2099** sin cambiar ni el config ni el código. Luego explica qué cambiaría si en vez de eso quisieras que **dentro** del contenedor escuche en 2099.

## Autoevaluación

1. Imagen vs contenedor.
2. En `"127.0.0.1:5433:5432"`, ¿qué es cada número y qué hace `127.0.0.1`?
3. ¿Por qué `dbpostgres` no resuelve desde tu máquina?
4. ¿Qué gana un Dockerfile multi-stage?
5. ¿Qué diferencia hay entre `down` y `down -v`, y por qué el segundo es peligroso?

**Listo cuando:** levantas base + API en compose y explicas el diagrama de la sección 2 con los hosts y puertos reales del dojo.

Siguiente: [08 · Arquitectura hexagonal](08-arquitectura-hexagonal.md)
