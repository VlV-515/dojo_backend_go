# 00 · Bienvenida: cómo se ve un backend Go empresarial y cómo funciona este dojo

> Cinturón blanco · 1 hora

## Objetivos

Al terminar sabrás:

- Qué forma tiene un backend de microservicios en Go dentro de una empresa.
- Qué piezas lo componen (servicios, librería compartida, repo de infraestructura, bases, colas).
- Qué entornos existen y por qué **nunca** se toca producción desde tu máquina.
- Cómo está organizado el dojo y cómo usar a Claude como sensei (opcional).

## 1. El sistema en una página

Imagina un ERP (sistema de gestión de empresa: clientes, pedidos, inventarios, cajas, facturación…) partido en **microservicios**. Es una arquitectura muy común en equipos que usan Go. Cada microservicio es un programa en **Go** que:

- Expone una **API HTTP** (aquí con el framework **Gin**) en su propio puerto.
- Guarda sus datos en **PostgreSQL** (con el ORM **GORM**). Algunos además usan **MongoDB**, **RabbitMQ** o Kafka (mensajes entre servicios) y **Elasticsearch** (búsquedas).
- Documenta su API con **Swagger/OpenAPI** (generado con `swag`).
- Corre en un servidor dentro de un contenedor **Docker**, junto con los demás, levantados por un `docker-compose.yml` o un orquestador como Kubernetes.

```
            ┌──────────────── Servidor (dev / staging / production) ────────────────┐
 Front ───► │ proxy (nginx) ─► auth:8001   pedidos:8002   inventarios:8003 ...      │
 web/móvil  │                     │              │               │                 │
            │            red docker interna (alias: dbpostgres, rabbit)            │
            │         Postgres        Mongo        RabbitMQ                         │
            └───────────────────────────────────────────────────────────────────────┘
```

Casi siempre hay una **librería compartida interna**: un módulo Go privado que todos los servicios importan, con lo común. Por ejemplo: errores, configuración, conexión a la base, autenticación, mensajería y trazas. Tiene ventajas (consistencia) y costos (sus errores de diseño los heredan todos). El módulo 09 se dedica a eso.

## 2. Los repos típicos de un equipo

| Repo | Qué es |
|---|---|
| `servicio-<dominio>` | Un microservicio por dominio de negocio: auth, pedidos, inventarios… |
| `libs` / `platform` | La librería compartida (privada). |
| `infra` / `deploy` | Infraestructura: `docker-compose.yml`, configs por entorno y pipelines. Suele tener una carpeta o rama por entorno. |
| `frontend` | El cliente web (no es tema de este dojo). |

En el dojo, **`practica/dojo-api`** hace el papel de un microservicio y **`pkg/apperr`** el de la librería compartida.

## 3. Entornos

| Entorno | Para qué |
|---|---|
| **local** | Tu máquina. Bases propias en Docker: puedes romper todo. |
| **dev** / **staging** | Probar ramas y lo que ya pasó revisión. Bases **compartidas** con el resto del equipo y con QA. |
| **production** | Clientes reales. **Nunca** desde tu máquina; solo a través del pipeline. |

> **Regla de oro:** las bases de dev/staging son **compartidas**. Un `POST`, `PUT` o `DELETE` desde tu máquina cambia datos que usa el resto del equipo. Una migración cambia el esquema de todos. Pregunta antes.

## 4. Cómo está organizado el dojo

```
dojo_backend_go/
├── README.md          ← ruta y reglas
├── CLAUDE.md          ← le dice a Claude que es tu sensei
├── modulos/           ← teoría + ejercicios (este archivo es el 00)
├── practica/
│   ├── katas/         ← ejercicios cortos de Go con tests (módulos 02 y 03)
│   └── dojo-api/      ← mini microservicio hexagonal (módulos 04 a 12)
├── recursos/          ← cheatsheet PDF, glosario, chuleta de comandos, lecturas
└── soluciones/        ← para mentores (y para ti DESPUÉS de intentar)
```

Cada módulo tiene: **objetivos**, **teoría** (siempre con el *por qué*), **en un proyecto real** (archivos de `dojo-api` que puedes abrir), **ejercicios**, un **reto** y una **autoevaluación**.

### Cinturones

| Cinturón | Módulos | Al terminar puedes… |
|---|---|---|
| Blanco | 00–01 | Moverte en la terminal, git y el flujo de un equipo |
| Amarillo | 02–03 | Escribir y probar Go idiomático |
| Naranja | 04–05 | Construir y documentar una API HTTP |
| Verde | 06–07 | Trabajar con Postgres, GORM, migraciones y Docker |
| Azul | 08–09 | Entender y extender un microservicio hexagonal con librerías compartidas |
| Café | 10–11 | Trabajar seguro y seguir un proceso de entrega con CI/CD |
| Negro | 12 | Proyecto final: entregar una feature completa de punta a punta |

## 5. Claude como sensei (opcional)

Abre una sesión de [Claude Code](https://claude.com/claude-code) **dentro de la carpeta del repo**. El `CLAUDE.md` le indica que es tu sensei: te explica, te da pistas y revisa tu código, pero **no resuelve los retos por ti** a menos que se lo pidas explícitamente después de intentarlo.

| Comando | Qué hace |
|---|---|
| `/progreso` | Muestra en qué módulo vas y qué te falta |
| `/pista` | Pista graduada del ejercicio en el que estás (sin darte la respuesta) |
| `/revisar-reto` | Corre las pruebas del reto y revisa tu código como lo haría un compañero de equipo |
| `/examen` | Preguntas de autoevaluación del módulo |

Si no usas Claude, copia `PROGRESO.plantilla.md` a `progreso.local.md` y llévala a mano.

## Ejercicio

1. Abre [`practica/dojo-api/src/infrastructure/app/app.go`](../practica/dojo-api/src/infrastructure/app/app.go) y lee solo los comentarios. No tienes que entender el código todavía; anota 3 palabras que no conozcas. Al final del dojo vuelve a leerlo.
2. Abre [`practica/dojo-api/docker-compose.yml`](../practica/dojo-api/docker-compose.yml) y encuentra: el puerto de la API, el puerto publicado de Postgres y el alias que usa la API para llegar a la base.
3. Dibuja el diagrama de la sección 1 con los servicios de alguna app que conozcas (una tienda en línea, un banco, una app de delivery).

## Autoevaluación

1. ¿Por qué `dbpostgres:5432` funciona dentro de la red de Docker pero no desde tu máquina?
2. ¿Qué diferencia hay entre local, staging y production? ¿En cuáles las bases son compartidas?
3. ¿Qué es una librería compartida interna y por qué necesitas configuración especial (`GOPRIVATE`) para bajarla?
4. Nombra dos acciones que **no** debes hacer contra una base compartida sin que te lo pidan.

**Listo cuando:** puedes dibujar de memoria el diagrama de la sección 1 y explicarle a alguien la regla de oro.

Siguiente: [01 · Terminal, Git y flujo de equipo](01-terminal-git-y-flujo.md)
