# 06 · PostgreSQL, GORM y migraciones (+ un poco de Mongo)

> Cinturón verde · 3–4 días

## Objetivos

- Escribir SQL básico en PostgreSQL y usar `psql`.
- Usar GORM v2 conociendo sus trampas: `Find` vs `First`, `Updates` con struct, transacciones.
- Entender el patrón `UnitOfWork` + `txOrDB` (transacción dentro del context).
- Escribir y aplicar migraciones SQL versionadas.
- Saber por qué, fuera de tu máquina, a Postgres se entra con SSL y certificado de cliente.

## 1. SQL mínimo

```sql
CREATE TABLE almacen (
  id serial PRIMARY KEY,
  clave varchar(20) NOT NULL UNIQUE,
  nombre varchar(120) NOT NULL,
  activo boolean NOT NULL DEFAULT true
);
INSERT INTO almacen (clave, nombre) VALUES ('ALM-01', 'Centro') RETURNING id;
SELECT id, clave FROM almacen WHERE activo ORDER BY id;
UPDATE almacen SET nombre = 'Norte' WHERE id = 1;
SELECT a.nombre, count(c.id) FROM almacen a LEFT JOIN caja c ON c.almacen_id = a.id GROUP BY a.nombre;
```

`psql` útil:

```bash
psql "host=127.0.0.1 port=5433 user=dojo dbname=dojo"
\dt            -- tablas
\d almacen     -- estructura
\q
```

El dojo usa `postgres:16-alpine`. En el trabajo revisa **siempre** qué versión corre en los servidores: no es raro encontrar versiones viejas en producción, y la sintaxis cambia. Ejemplo: `CREATE OR REPLACE TRIGGER` existe desde PostgreSQL 14; en versiones anteriores usa `DROP TRIGGER IF EXISTS` + `CREATE TRIGGER`. Escribe tus migraciones para la versión **más vieja** que las vaya a ejecutar.

## 2. GORM v2: lo que tienes que saber

```go
db.Where("id = ?", id).First(&a)          // 1 registro; si no hay: gorm.ErrRecordNotFound
db.Where("clave = ?", c).Find(&lista)     // N registros; si no hay: SIN error, RowsAffected == 0
db.Create(&a)                             // INSERT, llena a.ID
db.Model(&Almacen{}).Where("id = ?", id).Updates(map[string]any{"activo": false})
```

| Trampa | Consecuencia | Qué hacer |
|---|---|---|
| `Find` no devuelve `ErrRecordNotFound` | Tu rama de 404 nunca se ejecuta | `First` si esperas uno, o revisa `RowsAffected` |
| `Updates(struct)` ignora valores cero | `Activo=false`, `0`, `""` no se guardan | `Updates(map[string]any{...})` o `Select(...)` |
| `?` siempre | Concatenar strings = **inyección SQL** | Nunca `"id = " + id` |
| `Rollback()` sobre el `*gorm.DB` raíz | Deja el error pegado: todas las consultas siguientes del servicio fallan con *invalid transaction* hasta reiniciar (un incidente clásico en producción) | Transacciones solo con `uow.Do` |

## 3. UnitOfWork y txOrDB

El dominio no puede ver `*gorm.DB`. Entonces la transacción viaja **dentro del context**:

```go
// Puerto (dominio)
type UnitOfWork interface {
    Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// Adaptador (infraestructura) — src/infrastructure/persistence/postgres/unitOfWork.go
func (u *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
    return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        return fn(context.WithValue(ctx, txKey{}, tx))
    })
}
func txOrDB(ctx context.Context, db *gorm.DB) *gorm.DB {   // cada repositorio lo usa
    if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok { return tx }
    return db.WithContext(ctx)
}
```

Dos trampas (documentadas en [`src/domain/repositories/unitOfWork.go`](../practica/dojo-api/src/domain/repositories/unitOfWork.go) de `dojo-api`):

1. `Do` **commitea** si `fn` devuelve `nil`. Un `return nil` temprano **no** cancela: persiste lo hecho.
2. Un `Do` dentro de otro `Do` abre una transacción **independiente**, no un savepoint.

## 4. Migraciones

En un proyecto serio **el esquema NO sale de los structs** (nada de `AutoMigrate` en producción). Lo gobiernan archivos SQL versionados, embebidos en el binario con `//go:embed`. En `dojo-api` viven en [`src/infrastructure/database/migrations/`](../practica/dojo-api/src/infrastructure/database/migrations/0001_almacen.up.sql) y los aplica [`migrator.go`](../practica/dojo-api/src/infrastructure/database/migrator.go). Herramientas públicas que hacen lo mismo: golang-migrate, goose, atlas.

```bash
go run main.go migrate status    # solo lee: qué está aplicado y qué falta
go run main.go migrate up        # aplica pendientes
```

- `auto_migrate_on_boot: false` **siempre** que tu servicio local apunte a una base compartida (dev/staging): si no, tu rama le aplica migraciones a la base de todos. Lo mismo para `migrate up`: contra una base compartida, nunca sin permiso.
- Un `.sql` nuevo no existe para el binario hasta **recompilar** (está embebido).
- En proyectos con historia, la migración *baseline* suele salir de un dump de producción (ids `integer`, montos `numeric`, cosas que nadie recuerda): no deduzcas el tipo de una columna leyendo el struct; léelo en la base (`\d tabla`) o en la migración.

## 5. Conexión segura a Postgres

En tu máquina, `dojo-api` usa `sslmode: disable` porque la base vive en un contenedor local. En un servidor real es distinto: una configuración común es que `pg_hba.conf` solo acepte conexiones de fuera de la red interna con `hostssl ... cert` (TLS **y** certificado de cliente). Entonces, para conectarte desde tu máquina a una base de dev/staging necesitas `sslmode=verify-full` y tres archivos: `ca.crt` (para validar al servidor), `client.crt` y `client.key` (para que el servidor te valide a ti; la llave con `chmod 600`, o el driver la rechaza). Con autenticación `cert`, el CN de `client.crt` debe coincidir con el usuario de Postgres. Detalle en la documentación oficial: [postgresql.org/docs](https://www.postgresql.org/docs/) (busca "SSL Support" y "pg_hba.conf").

Esos archivos son secretos: van en una carpeta ignorada por git (`certs/`), nunca en el repo.

## 6. MongoDB (breve)

Muchos backends combinan Postgres con Mongo para casos puntuales: guardar políticas de permisos de casbin (tiene adaptadores para Mongo), logs o documentos sin esquema fijo. Si un sistema genera IDs `ObjectId` de Mongo y los guarda en Postgres, la columna suele ser `text` con el ObjectId en hexadecimal. Regla práctica: no mezcles los dos sin una razón clara; cada base extra es una dependencia más que operar.

## Ejercicios — `practica/dojo-api` con Postgres

Requiere Docker abierto (Docker Desktop en macOS/Windows; el módulo 07 explica todo lo de Docker; aquí basta con copiar).

```bash
cd practica/dojo-api
docker compose up -d dbpostgres
cp config.example.json config.json       # cambia persistencia a "postgres" y la contraseña a "dojo"
go run main.go migrate status            # 0001 PENDIENTE
go run main.go migrate up
go run main.go migrate status            # 0001 aplicada
go run main.go                           # /salud debe decir persistencia: postgres
```

1. Crea 3 almacenes con `curl`. Míralos en `psql`.
2. Lee `src/infrastructure/persistence/postgres/almacen.go`: ubica dónde se usa `First`, dónde `Find` + `RowsAffected` y por qué `Actualizar` usa un map.
3. **Experimento:** cambia temporalmente `Actualizar` a `Updates(almacen)` (struct). Desactiva un almacén (cuando tengas el reto B) y revisa en `psql` si se guardó. Regrésalo.
4. Escribe en `psql` una consulta que cuente almacenes activos e inactivos.

## Reto

Escribe `0002_almacen_direccion.up.sql` que agregue la columna `direccion varchar(200)` nullable. Aplícala, revisa `\d almacen` y `migrate status`. ¿Necesitaste cambiar el struct para que la migración funcione? ¿Qué pasaría con `SELECT *` del código actual? (Bórrala al terminar o renómbrala para no chocar con el reto C.)

> Para "deshacer" en tu base de práctica: `docker compose down -v` borra el volumen y empiezas de cero. **Jamás** hagas algo equivalente en una base compartida (dev/staging) ni, por supuesto, en producción.

## Autoevaluación

1. ¿Por qué `Find` + "si err == ErrRecordNotFound devuelve 404" es un bug?
2. ¿Qué pasa si haces `Updates(Almacen{Activo: false})`?
3. ¿Cómo llega la transacción del service al repositorio sin pasar `*gorm.DB`?
4. ¿Por qué `auto_migrate_on_boot` debe ser `false` cuando tu servicio local apunta a una base compartida?
5. ¿Qué tres archivos de certificado necesitas y por qué la llave lleva `chmod 600`?

**Listo cuando:** puedes levantar Postgres, migrar y explicar las 4 trampas de la tabla de GORM.

Siguiente: [07 · Docker y Docker Compose](07-docker.md)
