# 09 · Librerías compartidas y módulos privados de Go

> Cinturón azul · 2–3 días

## Objetivos

- Entender qué es la **librería compartida interna** que tienen casi todas las empresas con varios microservicios en Go, y por qué cambiarla es caro.
- Consumir módulos Go privados: `GOPRIVATE`, autenticación con git, proxy y checksum DB.
- Versionar una librería con **semver** y tags, y usar `replace` (o `go.work`) para desarrollarla en local.
- Usar bien una librería de errores compartida y **convivir con sus trampas**, practicando con `pkg/apperr`.
- Conocer las piezas típicas del ecosistema (config, auth/casbin, traza, colas, búsqueda, websockets) y cuándo **no** conectarte a las de un entorno compartido.

## 1. La librería compartida interna

Cuando un equipo tiene 5, 20 o 50 microservicios, el código que se repite (formato de errores, carga de config, middleware de auth, conexión a la base, logger, cliente HTTP entre servicios) termina en **un módulo Go privado** que todos importan, algo como `github.com/tu-org/go-commons`.

Ventajas: consistencia (todos responden errores igual), menos copia-pega, un solo lugar para arreglar bugs. Costo: **cada cambio afecta a todos los servicios** que la importan. Por eso:

- Las trampas de una librería compartida suelen **sobrevivir años**: arreglarlas puede romper servicios que (sin saberlo) dependen del comportamiento raro.
- Tu trabajo como desarrollador nuevo no es "arreglarla" a la primera, sino **conocer sus trampas y escribir código que conviva con ellas**.

En el dojo la simulamos con tres paquetes de `dojo-api`:

| Paquete | Simula |
|---|---|
| [`pkg/apperr`](../practica/dojo-api/pkg/apperr/apperr.go) | Librería de errores con código HTTP (con trampas a propósito) |
| [`pkg/general`](../practica/dojo-api/pkg/general/general.go) | Forma estándar de responder éxito (`{codigo, mensaje, datos}`) |
| [`src/config`](../practica/dojo-api/src/config/config.go) | Carga de config con default "cómodo" |

## 2. Módulos privados de Go

Por defecto, `go get` / `go mod download` bajan los módulos a través del proxy público (`proxy.golang.org`) y verifican su hash contra la checksum DB pública (`sum.golang.org`). Un repo privado **no existe** para ellos, así que hay que decirle a Go que esos módulos se bajan directo del repo y no se verifican en público:

```bash
go env -w GOPRIVATE=github.com/tu-org/*
go env GOPRIVATE GOPROXY GONOSUMDB      # GOPRIVATE implica GONOPROXY y GONOSUMDB
```

Luego Go usa **git** para bajarlo, así que git necesita credenciales para ese host:

- **HTTPS** con un token personal guardado en el gestor de credenciales del sistema (`osxkeychain` en macOS, `manager` en Windows) o en `~/.netrc`.
- **SSH** forzando que las URLs HTTPS se reescriban:

  ```bash
  git config --global url."git@github.com:tu-org/".insteadOf "https://github.com/tu-org/"
  ```

Prueba de acceso, antes de culpar a Go:

```bash
git ls-remote --tags https://github.com/tu-org/go-commons.git
```

Errores típicos y su causa:

| Mensaje | Causa probable |
|---|---|
| `410 Gone` / `not found` desde `proxy.golang.org` | Falta `GOPRIVATE`: Go fue a buscarlo al proxy público |
| `verifying module: ... checksum database` | Falta `GOPRIVATE` (o `GONOSUMDB`) |
| `terminal prompts disabled` / `could not read Username` | Git no tiene credenciales para ese host |
| Un askpass del editor que truena | El terminal integrado del IDE inyecta su propio `GIT_ASKPASS`; prueba en una terminal normal o con `GIT_ASKPASS=` |

En CI pasa lo mismo: el pipeline necesita `GOPRIVATE` y un token de solo lectura (como secreto del pipeline, nunca en el repo).

Documentación oficial: [go.dev/ref/mod](https://go.dev/ref/mod) (secciones de módulos privados y variables de entorno).

## 3. Versionado semver

Una librería compartida se publica con **tags de git** en formato semver `vMAYOR.MENOR.PARCHE`:

| Cambio | Sube | Ejemplo |
|---|---|---|
| Arreglo compatible | PARCHE | `v1.4.2` → `v1.4.3` |
| Funcionalidad nueva compatible | MENOR | `v1.4.3` → `v1.5.0` |
| Cambio que rompe a quien la usa | MAYOR | `v1.5.0` → `v2.0.0` |

Reglas de Go que sorprenden:

- `v0.x.y` no promete compatibilidad: cualquier versión puede romper.
- Desde `v2`, **la ruta del módulo cambia**: `module github.com/tu-org/go-commons/v2`. Para Go, `v1` y `v2` son módulos distintos y pueden convivir en el mismo binario.
- `go.mod` fija la versión exacta y `go.sum` sus hashes. Una actualización es un cambio de código que se revisa en PR como cualquier otro.

```bash
go list -m -versions github.com/gin-gonic/gin     # versiones publicadas de un módulo
go get github.com/tu-org/go-commons@v1.5.0        # subir a una versión concreta
go get github.com/tu-org/go-commons@latest
go mod why -m github.com/tu-org/go-commons        # quién la necesita
```

El código fuente de cada versión bajada vive, de solo lectura, en `$(go env GOMODCACHE)`. Las mayúsculas de la ruta se escriben con `!` (una organización `TuOrg` aparece como `github.com/!tu!org/...`). Leer ahí el código real de una librería es la forma más rápida de confirmar una trampa.

## 4. `replace` y `go.work` para desarrollo local

Cuando necesitas cambiar la librería **y** probarla en un servicio antes de publicar un tag:

```bash
# en el go.mod del servicio
go mod edit -replace github.com/tu-org/go-commons=../go-commons
go build ./...
# ...pruebas...
go mod edit -dropreplace github.com/tu-org/go-commons
```

- Un `replace` a una ruta local **nunca** se commitea: en CI y en el servidor esa ruta no existe y el build truena.
- Alternativa más limpia: un `go.work` (`go work init . ../go-commons`) **fuera** del commit (agrégalo a `.gitignore`). Go usa la copia local sin tocar `go.mod`.
- Flujo correcto: cambias la librería → PR en la librería → tag nuevo → `go get ...@vX.Y.Z` en cada servicio → PR en cada servicio.

## 5. Errores: `pkg/apperr` y sus trampas

| Función | Para qué |
|---|---|
| `apperr.New(code, msg, causa)` | Crear un error con código HTTP, mensaje para usuario y causa técnica |
| `apperr.Wrap(err)` | Propagar entre capas conservando el código |
| `apperr.Code(err)` | Leer el código HTTP (500 si no es `AppError`) |
| `apperr.IsCode(err, code)` | Decidir por código (no hay `Unwrap`) |
| `apperr.Response(err, ctx)` | Responder un error que viene del service |
| `apperr.CtrlResponse(code, msg, err, ctx)` | Responder un error propio del borde HTTP |

Las librerías reales suelen traer además algo como `GetFields(err)` para pasarle campos al logger.

**Trampas típicas de librerías internas, que el dojo trae a propósito** (están documentadas en el propio [`apperr.go`](../practica/dojo-api/pkg/apperr/apperr.go) y probadas en [`apperr_test.go`](../practica/dojo-api/pkg/apperr/apperr_test.go)):

- **Nunca** `return err` crudo entre paquetes: `return nil, apperr.Wrap(err)`.
- **`AppError` no implementa `Unwrap()`**: `errors.Is(err, gorm.ErrRecordNotFound)` sobre un `AppError` da `false`. Convierte a código en el adaptador (como hace `PorID` en [`postgres/almacen.go`](../practica/dojo-api/src/infrastructure/persistence/postgres/almacen.go)) y decide con `IsCode`.
- **`apperr.Wrap(nil)` NO devuelve nil.** `return x, apperr.Wrap(err)` con `err == nil` convierte un éxito en error. Si absorbes un not-found: `return nil, nil` explícito (mira `PorClave` en el mismo archivo).
- **`apperr.New(code, msg, nil)` revienta al responder** (`Response` llama `DevelopMessage.Error()` sobre nil). Siempre pasa causa: `apperr.New(400, msg, errors.New(msg))`, como hace `badRequest` en [`services/almacen.go`](../practica/dojo-api/src/domain/services/almacen.go).

¿Por qué no se arreglan? Agregar `Unwrap` cambia el resultado de cada `errors.Is` en todos los servicios; hacer que `Wrap(nil)` devuelva nil puede "arreglar" bugs de los que otro código depende sin saberlo. Es un cambio MAYOR disfrazado de parche. Se hace, pero con plan, versión nueva y revisión de cada consumidor.

## 6. Config

[`src/config`](../practica/dojo-api/src/config/config.go) lee `config.json` del directorio actual o la ruta de `CONFIG`. **Si el archivo no existe, no falla**: avisa en el log y arranca con una configuración de prueba interna (persistencia en memoria). Muchas librerías internas hacen lo mismo. Si "conecta" pero no ves tus datos, revisa esa línea del log.

Los secretos no vienen del archivo sino de variables de entorno (`DOJO_TOKEN` en el dojo; algo como `APP_TOKEN_SECRET_KEY` para firmar JWT en un sistema real). Si la variable falta, se usa un default inseguro **solo apto para local**. En servidores viven en un `.env` o en el gestor de secretos del entorno.

## 7. Autenticación, casbin y tokens entre servicios

Patrón común en una librería compartida de auth:

- Un middleware valida el JWT del usuario y pregunta a **[casbin](https://casbin.org/)** (políticas en una base) si su rol puede usar esa ruta. Sin permiso → **403**. En el dojo, [`middleware/auth.go`](../practica/dojo-api/src/infrastructure/web/middleware/auth.go) es la versión simplificada: solo compara un token (401 si falla).
- La librería registra las rutas del servicio para casbin. Una ruta nueva nace **sin rol**: da 403 hasta que se le asignan permisos **en cada entorno**.
- **Token de aplicación** (un header tipo `X-App-Token`): llamadas servicio-a-servicio. Trampa clásica: una bandera de config que "desactiva" la validación del token de aplicación y, por un bug, termina dejando pasar **cualquier** token no vacío sin revisar JWT ni casbin. **Nunca apagues una validación de seguridad para quitar un warning**; lee el código de la librería antes.
- Un cliente HTTP compartido que inyecta el token interno **nunca** se usa hacia terceros (un proveedor de identidad va con su propio `http.Client`). Y si ese cliente no tiene timeout, acótalo tú con `context.WithTimeout` (kata `ConsultarConLimite` del módulo 03).

## 8. Traza y logs

Cada petición lleva un **TraceID** en el `context`; los logs lo incluyen para seguir una operación entre servicios. Por eso se pasa siempre `ctx`, y por eso el trabajo en segundo plano se lanza con un helper de la librería (que propaga la traza y recupera pánicos) en vez de un `go func(){}` suelto.

## 9. Colas de mensajes (RabbitMQ)

Cola de mensajes entre servicios (notificaciones, eventos). Dos roles:

- **Publicar**: mandar un mensaje. Un servicio bien hecho, si la cola no responde, avisa en el log y sigue sin publisher.
- **Consumir**: leer de una cola. **Si tu servicio consume y lo conectas al RabbitMQ de un entorno compartido (`dev`/`staging`), le robas mensajes al entorno**: el mensaje lo procesa tu máquina y no el servidor. Regla: cola compartida solo si el servicio **no consume**; si consume, deja un host que no resuelva (el servicio sigue sin mensajería) o levanta tu propio RabbitMQ en Docker.

## 10. Búsqueda, `LISTEN` y websockets

- **Elasticsearch**: índices para búsquedas. En local apúntalo a un host inexistente (o a tu propio contenedor) para no reindexar el índice compartido.
- **`LISTEN`/`NOTIFY` de Postgres**: un servicio puede escuchar cambios de una tabla. Si la conexión de escucha es crítica y no abre, el arranque hace `Fatal`: un config o certificado mal puesto tumba el servicio completo.
- **Websockets**: avisos en tiempo real al front.

## 11. Shutdown

Si la librería coordina el apagado ordenado (`Shutdown`) pero no aplica su propio timeout, cada tarea registrada debe respetar `ctx.Done()` o el apagado se cuelga.

## Ejercicios

1. Abre el código real de una dependencia pública en tu caché: `ls $(go env GOMODCACHE)/github.com/gin-gonic/` y entra a la versión que usa `dojo-api` (míralo en `go.mod`). Localiza `func (c *Context) AbortWithStatusJSON`. Así se lee una librería compartida.
2. En `practica/dojo-api`, lee `pkg/apperr/apperr_test.go`. Escribe un test nuevo que demuestre que `apperr.New(400, "x", nil)` **revienta** al pasar por `Response` (pista: `httptest` + `gin.CreateTestContext` + `defer recover()`).
3. En `src/infrastructure/persistence/postgres/almacen.go`, `PorClave` absorbe un not-found con `return nil, nil`. Explica por qué no podía ser `return nil, apperr.Wrap(err)`.
4. Simula una librería privada con `replace`: copia `practica/dojo-api` a una carpeta temporal, crea al lado un módulo `example.com/mis-errores` con una función, agrégalo con `go mod edit -require=example.com/mis-errores@v0.0.0 -replace=example.com/mis-errores=../mis-errores` y úsalo. Luego quita el `replace` y lee el error de `go build`. ¿Qué tendría que existir para que funcione sin `replace`?
5. Corre `go env GOPRIVATE GOPROXY GONOSUMDB` y explica qué pasaría hoy si intentaras bajar un módulo privado de tu organización.

## Reto — Bug hunt de errores

En una copia del service de almacenes de `dojo-api`, escribe un método `BuscarPorClave(ctx, clave) (*models.Almacen, error)` que use `PorClave` y devuelva 404 si no existe. Escríbelo **primero mal** con `return almacen, apperr.Wrap(err)` en todos los caminos, escribe el test que lo atrapa, y luego corrígelo. Explica el bug en una frase.

## Autoevaluación

1. ¿Por qué con `apperr` se decide con `IsCode` y no con `errors.Is`?
2. ¿Qué pasa con `return x, apperr.Wrap(nil)`?
3. ¿Qué hace `GOPRIVATE` y qué dos variables implica?
4. ¿Por qué un `replace` a ruta local no se commitea? ¿Qué alternativa tienes?
5. Agregar `Unwrap` a `AppError`, ¿es parche, menor o mayor? ¿Por qué?
6. ¿Cuándo **no** debes apuntar tu local al RabbitMQ de un entorno compartido?
7. ¿Qué responde una ruta nueva recién desplegada en un sistema con casbin y cómo se arregla?

**Listo cuando:** puedes configurar Go para bajar un módulo privado, explicar cómo se publica una versión nueva de una librería, y revisar un PR detectando `return err` crudo, `Wrap(nil)` y `New(..., nil)` a simple vista.

Siguiente: [10 · Seguridad, configuración y secretos](10-seguridad-y-config.md)
