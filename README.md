# 🥋 Dojo Backend Go

Curso práctico, gratuito y en español para aprender **backend con Go a nivel empresarial**. No son ejemplos sueltos: construyes y extiendes un microservicio con la forma que tienen los de un equipo real (capas, pruebas, migraciones, Docker, Swagger, librerías compartidas, seguridad y CI/CD), guiado por retos que empiezan en rojo y terminan en verde.

**Para quién:** programadores que vienen de otro lenguaje (JS, Python, Java, PHP, C#…) o que ya saben algo de Go y quieren dar el salto a cómo se trabaja en una empresa.

**Qué aprendes:**

| Área | Temas |
|---|---|
| Go | Tipos, structs, interfaces, errores (`%w`, `errors.Is/As`), genéricos, goroutines, canales, `context`, `sync`, tests table-driven, fakes, `-race` |
| APIs | Gin: rutas, grupos, middlewares, binding y validación, DTOs y mappers, códigos HTTP, `httptest` |
| Documentación | Swagger/OpenAPI con `swaggo/swag` |
| Datos | PostgreSQL, GORM v2, transacciones (Unit of Work), migraciones SQL versionadas y embebidas |
| Infraestructura | Docker multi-stage, Docker Compose, redes y alias, healthchecks |
| Arquitectura | Hexagonal (puertos y adaptadores), composition root, tests de arquitectura |
| Equipo | Librerías compartidas y módulos privados, configuración y secretos, Git, PRs, CI/CD y despliegue por entornos |

## Requisitos

- Go ≥ 1.24, git y Docker (Docker Desktop, OrbStack, Colima o Docker Engine).
- macOS, Linux o Windows con WSL2.
- `swag` (se instala en el módulo 05).
- Opcional pero recomendado: [Claude Code](https://claude.com/claude-code), que funciona como sensei del dojo (ver abajo).

## Cómo empezar

```bash
git clone https://github.com/VlV-515/dojo_backend_go.git
cd dojo_backend_go
```

1. Lee [`modulos/00-bienvenida.md`](modulos/00-bienvenida.md).
2. Comprueba que todo compila:
   ```bash
   cd practica/katas && go test ./...          # en rojo: así empiezan las katas
   cd ../dojo-api && go install github.com/swaggo/swag/cmd/swag@v1.16.3 \
     && swag init -g main.go && go test ./...  # en verde
   ```
3. Ten a la mano el **[cheatsheet en PDF](recursos/cheatsheet-backend-go.pdf)** (4 páginas, carta horizontal). Resume todo el dojo y sirve también en el día a día.

### Con Claude Code como sensei (opcional)

```bash
cd dojo_backend_go
claude
```

El [`CLAUDE.md`](CLAUDE.md) de esta carpeta convierte a Claude en tu **sensei**: explica, da pistas graduadas y revisa tu código como en un PR, pero **no te resuelve los retos**. Trae estos comandos:

| Comando | Qué hace |
|---|---|
| `/progreso` | Crea tu bitácora (`progreso.local.md`, no se versiona) y te dice qué sigue |
| `/pista` | Pista en 3 niveles del ejercicio en el que estás, sin darte la respuesta |
| `/revisar-reto` | Corre los tests del reto y te hace code review |
| `/examen` | Autoevaluación del módulo, una pregunta a la vez |

Sin Claude el dojo funciona igual: cada módulo trae teoría, ejercicios, tests y autoevaluación.

## Ruta

| # | Módulo | Cinturón | Práctica | Tiempo |
|---|---|---|---|---|
| 00 | [Bienvenida](modulos/00-bienvenida.md) | Blanco | Lectura guiada | 1 h |
| 01 | [Terminal, Git y flujo de equipo](modulos/01-terminal-git-y-flujo.md) | Blanco | Git + SSH | 3–4 h |
| 02 | [Go fundamentos](modulos/02-go-fundamentos.md) | Amarillo | `katas/basico` | 2–3 d |
| 03 | [Go intermedio](modulos/03-go-intermedio.md) | Amarillo | `katas/intermedio` | 3–4 d |
| 04 | [APIs HTTP con Gin](modulos/04-http-con-gin.md) | Naranja | dojo-api · **Reto A** | 2–3 d |
| 05 | [Swagger con swaggo](modulos/05-swagger.md) | Naranja | dojo-api | 1 d |
| 06 | [PostgreSQL, GORM y migraciones](modulos/06-postgres-gorm-migraciones.md) | Verde | dojo-api + Postgres | 3–4 d |
| 07 | [Docker y Docker Compose](modulos/07-docker.md) | Verde | dojo-api en compose | 2 d |
| 08 | [Arquitectura hexagonal](modulos/08-arquitectura-hexagonal.md) | Azul | dojo-api · **Reto B** | 3–4 d |
| 09 | [Librerías compartidas y módulos privados](modulos/09-librerias-compartidas.md) | Azul | `pkg/apperr` | 2–3 d |
| 10 | [Seguridad, configuración y secretos](modulos/10-seguridad-y-config.md) | Café | dojo-api | 1–2 d |
| 11 | [Flujo de trabajo, CI/CD y despliegue](modulos/11-flujo-y-despliegue.md) | Café | PR simulado | 1 d |
| 12 | [Proyecto final](modulos/12-proyecto-final.md) | Negro | **Reto C** + stack completo | 3–5 d |

Total aproximado: 4–5 semanas a medio tiempo.

## Qué hay aquí

```
dojo_backend_go/
├── README.md                  ← estás aquí
├── CLAUDE.md                  ← instrucciones del sensei (Claude Code)
├── PROGRESO.plantilla.md      ← plantilla de tu bitácora
├── .claude/skills/            ← /progreso, /pista, /revisar-reto, /examen
├── modulos/                   ← 13 módulos: teoría, ejercicios, reto, autoevaluación
├── practica/
│   ├── katas/                 ← Go con tests (módulos 02–03)
│   └── dojo-api/              ← mini microservicio hexagonal (módulos 04–12)
├── recursos/                  ← cheatsheet PDF, glosario, chuleta de comandos, lecturas
└── soluciones/                ← para mentores (y para ti, DESPUÉS de intentar)
```

### dojo-api en 30 segundos

Un catálogo de almacenes con Gin + GORM, con esta arquitectura:

```
src/domain/          models · repositories (puertos) · services (casos de uso)
src/infrastructure/  persistence (memoria | postgres) · web (controllers, dtos, mappers, routes, middleware)
                     database (migraciones embebidas + CLI) · app (composition root)
pkg/                 apperr (errores estilo librería interna) · general (respuestas)
```

```bash
cd practica/dojo-api
go run main.go                          # en memoria, puerto 1099
curl localhost:1099/salud
open http://localhost:1099/swagger/index.html
docker compose up -d --build            # API + Postgres
```

## Cómo se trabaja cada módulo

1. Lee la teoría y abre los archivos de `practica/` que menciona.
2. Haz los ejercicios. Si te atoras más de 20 minutos, pide una pista (`/pista` o a tu mentor).
3. Resuelve el reto. Pide revisión (`/revisar-reto` o un PR a tu mentor).
4. Contesta la autoevaluación sin ver la teoría.
5. Marca el módulo en tu bitácora.

Las pruebas son la especificación: un ejercicio está hecho cuando su test pasa **y** puedes explicar por qué.

## Reglas del dojo (y de cualquier equipo serio)

1. **Bases compartidas:** en dev/staging nada de `POST/PUT/DELETE` ni migraciones sin permiso del equipo. Para experimentar usa el Postgres de `practica/dojo-api` (Docker), que es solo tuyo.
2. **Nada de producción** desde tu máquina.
3. **Secretos:** ni en commits, ni en chats, ni en la bitácora. Di dónde viven, no qué valen.
4. **Intenta antes de ver la solución.** `solucion.go` y `soluciones/` existen para comparar, no para copiar.

## Para mentores

¿Lo usas para hacer onboarding de tu equipo? Adelante.

- `soluciones/README.md` explica cómo verificar katas y retos.
- Katas: `go test -tags solucion ./...` en `practica/katas` corre las soluciones de referencia.
- Retos de dojo-api: `soluciones/dojo-api-retos.patch` (se aplica sobre una copia limpia).
- Para adaptarlo a tu empresa, agrega un módulo con tus repos, entornos y proceso de despliegue. Mantenlo fuera del fork público si trae datos internos.

## Contribuir

Issues y PRs son bienvenidos: erratas, ejercicios nuevos, mejores explicaciones o traducciones. Antes de abrir un PR corre la verificación completa de [`soluciones/README.md`](soluciones/README.md) (katas con y sin solución, dojo-api en verde, retos en rojo sin patch y en verde con él). Si cambias algo del cheatsheet, regenera el PDF con `recursos/cheatsheet/generar.py`.

## Licencia

[MIT](LICENSE).
