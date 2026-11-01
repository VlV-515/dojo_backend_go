# 11 · Flujo de trabajo, CI/CD y despliegue

> Cinturón café · 1 día

## Objetivos

- Recorrer un ticket de principio a fin.
- Entender cómo se despliega típicamente a `staging` y a `production`.
- Preparar un PR que se revise rápido.
- Usar un flujo spec → plan → ejecución (con o sin asistente de IA) con criterio.

## 1. Vida de un ticket

```
Ticket PROJ-123 (Jira, Linear, GitHub Issues…)
   │
   ├─ rama feature/PROJ-123-desactivar-almacenes desde develop (o main)
   ├─ código + tests + swagger + migración (si aplica)
   ├─ verificación local: build, vet, fmt, test, curl
   ├─ deploy a STAGING (entorno compartido de pruebas)
   ├─ PR → CI en verde + revisión + QA → merge
   └─ el pipeline despliega (develop → dev/staging; main → production)
```

Variantes que vas a encontrar:

- **Git flow** (`develop` + `main`): lo que se integra a `develop` llega a `dev`/`staging`; un merge o release a `main` despliega producción.
- **Trunk-based** (solo `main` + ramas cortas): cada merge a `main` despliega a staging automáticamente y producción se promueve con un tag o una aprobación manual.

Lo importante es que **sepas cuál usa tu equipo antes de tu primer PR**: está en el `README`, en la carpeta del pipeline (`.github/workflows/`, `.gitlab-ci.yml`, etc.) o pregúntalo.

## 2. CI: lo que corre en cada PR

Un pipeline mínimo para un servicio Go (ejemplo con GitHub Actions, en `.github/workflows/ci.yml`):

```yaml
name: ci
on: [pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go install github.com/swaggo/swag/cmd/swag@v1.16.3 && swag init -g main.go
      - run: test -z "$(gofmt -l .)"
      - run: go vet ./...
      - run: go test -race ./...
```

Si el servicio usa una librería privada (módulo 09), el pipeline además necesita `GOPRIVATE` y un token de solo lectura guardado como **secreto del pipeline**. Documentación: [docs.github.com/actions](https://docs.github.com/actions).

## 3. Staging compartido

Staging suele ser **un solo entorno para todo el equipo**. Según cómo lo maneje tu equipo:

- Si se despliega **una rama a la vez**, desplegar la tuya puede **tirar** lo que tu compañero estaba probando: avisa en el canal del equipo antes.
- Si el equipo despliega un **conjunto de ramas** (todas las que están en prueba + la tuya), revisa primero qué hay desplegado y vuelve a incluirlas.
- Después de desplegar, confirma el estado (salud del servicio, versión desplegada, logs).

> En el dojo, Claude **no** despliega ni ejecuta comandos de despliegue: solo te guía.

## 4. Antes de desplegar: lo que se olvida

- **Ruta nueva** → si hay autorización por rol (casbin u otra), nace sin permisos (403). Asignarlos **en cada entorno**.
- **Nodo nuevo en config** → agregarlo a `config.example.json` **y** a los configs por entorno del repo de infraestructura (o a las variables del entorno), **antes** de desplegar el código. Si no, el servicio arranca sin ese nodo (o con su valor cero) y falla de forma silenciosa.
- **Migración nueva** → define quién y cuándo la aplica en cada entorno. En producción muchas veces se aplica a mano (`migrate up`) o con un job controlado; confírmalo.
- **Variables/secretos nuevos** → darlos de alta en el gestor de secretos de cada entorno antes del despliegue.

## 5. Un buen PR

- Título: `PROJ-123: [tipo] mensaje` (o el formato de tu equipo, p. ej. Conventional Commits `feat: ...`).
- Descripción: qué cambia, por qué, cómo se probó (comandos y `curl`), riesgos, pasos de despliegue (permisos, config, migraciones, secretos).
- Diff chico y enfocado. Nada de `docs/` generados, binarios ni `config.json`.
- Verificación: `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` + ejercicio manual del endpoint. **Si no pudiste ejercitarlo, dilo.**
- Un `ok` de `go test` con todo en `Skip` (p. ej. un test de integración que se salta si no hay `DOJO_TEST_DSN`) **no** verificó nada: corre con `-v` y lee qué se ejecutó.

## 6. Spec → plan → ejecución

Para cambios no triviales funciona separar tres pasos, a mano o con un asistente de IA como Claude Code:

1. **Spec**: qué y por qué (contrato, casos de error, fuera de alcance).
2. **Plan**: el cambio partido en rebanadas pequeñas, cada una verificable (por ejemplo, las capas de la receta del módulo 08).
3. **Ejecución**: rebanada por rebanada, con tests en verde entre cada una.

Regla: el asistente **no commitea por ti**; revisas el diff y commiteas tú. Úsalo cuando el cambio toca varias capas; para un fix de una línea es excesivo.

## Ejercicios

1. Escribe un `.github/workflows/ci.yml` para `practica/dojo-api` (no hace falta subirlo). ¿Qué paso falla si alguien olvida correr `gofmt`? ¿Qué pasaría con `go test` si no corres `swag init` antes?
2. Redacta (en un archivo local, no lo subas) la descripción del PR del reto B con todas las secciones de la sección 5.
3. Escribe los pasos para desplegar `feature/PROJ-999-cajas` a un staging compartido donde ya están `feature/PROJ-111-x` y `feature/PROJ-222-y`, suponiendo que el equipo despliega un conjunto de ramas. ¿Qué pasa si solo despliegas la tuya?
4. Escribe un spec de una página para el reto C (módulo 12) y su plan en rebanadas. ¿Qué te hubiera ahorrado en el reto B?

## Reto

Simulación: tu feature agregó la ruta `PUT /almacenes/:id/desactivar` y un nodo `limites` al config. Escribe el checklist de despliegue completo, en orden, para staging → production. Revísalo con `/revisar-reto`.

## Autoevaluación

1. En un staging compartido, ¿por qué te tiene que importar qué otras ramas están desplegadas?
2. Desplegaste y tu ruta responde 403. ¿Qué falta?
3. ¿Qué va antes: dar de alta el nodo nuevo en los configs del entorno o desplegar el código? ¿Por qué?
4. ¿Qué tiene que decir un PR sobre cómo se probó?
5. ¿Qué pasos mínimos debe tener el CI de un servicio Go?

**Listo cuando:** puedes llevar un ticket de la rama al merge sin preguntar el proceso.

Siguiente: [12 · Proyecto final](12-proyecto-final.md)
