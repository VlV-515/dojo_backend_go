# 10 · Seguridad, configuración y secretos

> Cinturón café · 1–2 días

## Objetivos

- Saber dónde viven los secretos de un microservicio y dónde **nunca** deben aparecer.
- Configurar un servicio para correr en local sin poner en riesgo los entornos compartidos (`dev`/`staging`).
- Reconocer los riesgos más comunes en una API Go (inyección, timeouts, tokens, exposición de rutas).

## 1. Dónde viven los secretos

| Secreto | Vive en | Nunca en |
|---|---|---|
| Contraseñas de Postgres/Mongo/RabbitMQ | `config.json` (raíz del servicio, ignorado por git) y los configs por entorno del repo de infraestructura (privado, con acceso restringido) | commits del servicio, PRs, chats públicos, documentación, `README` |
| Certificados TLS (`ca.crt`, `client.crt`, `client.key`) | una carpeta `certs/<entorno>/` ignorada, con `chmod 600` a la llave | repos, chats |
| Llaves de firma de tokens (`APP_TOKEN_SECRET_KEY`, `DOJO_TOKEN`…) | variables de entorno: `.env` del servidor o gestor de secretos | código, configs versionados |
| Llaves de terceros (proveedor de identidad, APIs externas) | config/variables | `example:` de Swagger, logs |

- Un secreto que entra a git **se queda en el historial** aunque lo borres en el siguiente commit. La única corrección real es **rotarlo** (cambiarlo en el sistema que lo emite) y, si hace falta, reescribir el historial. Por eso no basta con "ya lo borré".
- Antes de crear un archivo con secretos: `git check-ignore -v <archivo>`. En `dojo-api`, [`.gitignore`](../practica/dojo-api/.gitignore) ignora `config.json` y versiona solo [`config.example.json`](../practica/dojo-api/config.example.json) con un placeholder.
- Para referirte a un secreto en documentación: di **dónde vive**, no su valor.
- Herramientas como [gitleaks](https://github.com/gitleaks/gitleaks) o el secret scanning de GitHub detectan secretos antes de que lleguen al remoto.

## 2. Configurar un servicio para local

Partiendo del config del entorno (`dev` o `staging`) que está en el repo de infraestructura:

1. Alias docker → hosts públicos (`dbpostgres:5432` → el host y puerto publicados del entorno; módulo 07).
2. Las URLs de otros servicios (`http://<servicio>`) → `localhost:<puerto>` si los corres tú, o el dominio público del entorno.
3. TLS: si la base exige certificado de cliente fuera de la red docker, activa SSL y apunta a tus certificados (`sslmode=verify-full` cuando el certificado del servidor trae el host en su SAN).
4. **`auto_migrate_on_boot: false`**: nunca migres una base compartida al arrancar tu máquina.
5. Apaga integraciones que "hablan hacia afuera" (alertas a un chat del equipo, correos, webhooks): no quieres mandar alertas del equipo desde tu laptop.
6. Colas: la compartida solo si el servicio no consume (módulo 09).
7. Valida el JSON: `python3 -m json.tool config.json` (o `jq . config.json`).

Y en las bases compartidas: **cero `POST/PUT/DELETE` ni migraciones (`up`/`down`/`force`) sin permiso explícito**. Leer (`GET`, `migrate status`) sí. Es una buena práctica universal: otras personas están probando contra esos datos.

**Nunca** uses configs ni certificados de **producción** en tu máquina.

## 3. Riesgos comunes en una API Go

| Riesgo | Ejemplo | Defensa |
|---|---|---|
| Inyección SQL | `db.Where("clave = '" + clave + "'")` | Siempre `?` |
| Ruta expuesta | Registrada antes del middleware (reto A) | Orden + test de inventario de rutas |
| Pánico por input | `strings.Split(header, " ")[1]` con un `Authorization` sin espacio | Parsear con cuidado (`strings.Cut`, como en [`middleware/auth.go`](../practica/dojo-api/src/infrastructure/web/middleware/auth.go)), `gin.Recovery()` |
| Sin timeout | Un cliente HTTP compartido sin límite, `http.Get` | `context.WithTimeout`, `http.Client{Timeout: ...}` |
| Comparar tokens con `==` | Filtra información por tiempos | `subtle.ConstantTimeCompare` |
| Bypass de autorización | Una bandera de config que apaga la validación y abre un hueco (módulo 09) | No apagar validaciones; leer el código de la librería |
| Secretos en logs | Loguear el config completo o el header `Authorization` | Loguear solo lo necesario |
| Default inseguro | Llave de firma sin definir → se usa un default conocido | En servidores siempre definida; en local **nunca** cargues la real de un entorno compartido (tu máquina podría firmar tokens válidos o propagar cambios a ese entorno) |

Referencia pública: [OWASP API Security Top 10](https://owasp.org/).

## Ejercicios

1. En `practica/dojo-api`, arranca **sin** `DOJO_TOKEN` y luego con `DOJO_TOKEN=otro go run main.go`. ¿Qué cambia en el log y en las peticiones?
2. Busca concatenación en consultas: `grep -rn "Where(" src/infrastructure/persistence | grep "+"`. ¿Hay algo? Escribe a mano un ejemplo vulnerable y explica qué input lo explota.
3. Desde `practica/dojo-api`, corre `git check-ignore -v config.json config.example.json certs/ .env` e interpreta la salida (qué está ignorado y por qué regla).
4. Lee [`middleware/auth.go`](../practica/dojo-api/src/infrastructure/web/middleware/auth.go) y anota qué riesgos de la tabla de la sección 3 cubre y cuáles no.

## Reto

Escribe un test para `practica/dojo-api` que verifique que la respuesta 401 **no** incluye el token esperado en ningún campo (ni `mensaje` ni `detalle`), probando con un token incorrecto. ¿Pasa? ¿Por qué es importante?

## Autoevaluación

1. Subiste un `config.json` por error y lo borraste en el siguiente commit. ¿Ya está resuelto? ¿Qué hay que hacer?
2. ¿Por qué apagar alertas e integraciones salientes en local?
3. ¿Por qué no debes cargar la llave de firma real de un entorno compartido cuando trabajas 100 % local?
4. Nombra tres defensas contra rutas expuestas.

**Listo cuando:** puedes configurar un servicio para local y explicar cada cambio y su riesgo.

Siguiente: [11 · Flujo de trabajo, CI/CD y despliegue](11-flujo-y-despliegue.md)
