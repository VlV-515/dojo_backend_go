# Glosario

| Término | Qué es | Dónde lo ves |
|---|---|---|
| **Adaptador** | Implementación concreta de un puerto (GORM, HTTP, colas, memoria) | `practica/dojo-api/src/infrastructure/persistence/` |
| **apperr** | Paquete del dojo que simula el de errores de una librería interna: errores con código HTTP | `practica/dojo-api/pkg/apperr/`, módulo 09 |
| **App token** | Token de servicio (por ejemplo, en un header `X-App-Token`) para llamadas entre microservicios | Módulo 09 |
| **auto_migrate_on_boot** | Si el servicio aplica migraciones al arrancar. `false` en local y contra bases compartidas | `config.json` de dojo-api |
| **Build tag** | `//go:build x`: el archivo solo compila con `-tags x` | Katas y retos |
| **Casbin** | Librería pública de autorización: permisos por rol y ruta. Sin permiso = 403 | Módulos 10 y 11 |
| **CI/CD** | Integración y despliegue continuos: el pipeline corre tests y despliega al hacer merge | Módulo 11 |
| **Composition root** | Único lugar donde se crean y conectan las dependencias | `src/infrastructure/app/app.go` |
| **context.Context** | Lleva cancelación, límite de tiempo y valores de petición (TraceID, transacción) | Primer parámetro de todo |
| **Contrato** | Lo que la API promete: rutas, JSON, códigos. No se rompe sin avisar | DTOs, Swagger |
| **dojonet** | Red de Docker Compose de dojo-api; sus alias (`dbpostgres`) solo resuelven dentro de la red | `practica/dojo-api/docker-compose.yml` |
| **DTO** | Struct del contrato JSON (entrada/salida HTTP) | `web/dtos/` |
| **Entornos** | `local` (tu máquina), `dev`/`staging` (bases compartidas del equipo) y `production` | Módulos 10 y 11 |
| **Fake** | Implementación en memoria de un puerto para tests, que registra lo que pasó | `*_test.go` de services |
| **GOPRIVATE** | Le dice a Go que baje ciertos módulos directo con git, sin proxy público (`GOPRIVATE=github.com/tu-org/*`) | Módulos 01 y 09 |
| **GORM** | ORM de Go (v2) | `persistence/postgres/` |
| **Hexagonal** | Arquitectura de puertos y adaptadores; dependencias hacia adentro | Módulo 08 |
| **Impl** | Sufijo de la implementación de una interfaz (`AlmacenServiceImpl`) | services, persistence |
| **Librería compartida interna** | Módulo Go privado de una empresa con errores, config, auth y DB comunes a todos sus servicios | Módulo 09; en el dojo, `pkg/` y `src/config` |
| **LISTEN** | Mecanismo de Postgres para recibir notificaciones de cambios | Módulo 09 |
| **Load-bearing** | "Que sostiene peso": si lo cambias, algo se rompe aunque no lo parezca | Orden de rutas |
| **Mapper** | Función que convierte DTO ↔ modelo | `web/mappers/` |
| **Migración** | Archivo SQL versionado que cambia el esquema | `database/migrations/` |
| **Pipeline** | Secuencia automática de CI/CD (por ejemplo, un workflow de GitHub Actions) | Módulo 11 |
| **Puerto** | Interfaz que el dominio necesita (`...Reader`, `...Writer`, `UnitOfWork`) | `domain/repositories/` |
| **Repo de infraestructura** | Repo del equipo con el compose y las configs por entorno; separado del código de cada servicio | Módulos 07 y 11 |
| **Service** | Lógica de negocio; depende solo de puertos y modelos | `domain/services/` |
| **Staging** | Entorno compartido parecido a producción donde se prueba antes del merge | Módulo 11 |
| **swag / Swagger** | Generador de documentación OpenAPI a partir de comentarios | Módulo 05 |
| **Table-driven test** | Test con tabla de casos y `t.Run` por caso | Todos los tests |
| **TraceID** | Identificador que sigue una operación por logs y servicios | Módulos 03 y 09 |
| **txOrDB** | Devuelve la transacción del ctx o la base raíz | `persistence/postgres/almacen.go` |
| **UnitOfWork** | Puerto para transacciones sin exponer `*gorm.DB` | Módulo 06 |
| **Valor cero** | Valor inicial de un tipo (`0`, `""`, `false`, `nil`) | Módulo 02 |
