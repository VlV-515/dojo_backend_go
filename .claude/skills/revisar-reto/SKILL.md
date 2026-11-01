---
name: revisar-reto
description: Revisa una kata o reto del Dojo Backend Go como lo haría un compañero de un equipo backend Go profesional en un PR - corre tests, vet, gofmt y da hallazgos con criterios de equipo sin reescribir el código. Úsalo con "/revisar-reto", "revisa mi código", "¿está bien mi solución?", "terminé el reto".
---

# Revisión de reto (estilo PR de equipo)

Argumento opcional: `basico`, `intermedio`, `A`, `B`, `C`, o un módulo/ejercicio. Si falta, deduce del avance y de los cambios recientes (`git status`/`git diff` si la carpeta tiene git; si no, por fechas de archivos).

## 1. Verificación automática

Corre y reporta resultado de cada uno (sin pegar salidas largas; cita el renglón decisivo):

| Qué | Dónde | Comando |
|---|---|---|
| Kata básico | `practica/katas` | `go test ./basico/ -v` |
| Kata intermedio | `practica/katas` | `go test ./intermedio/ -race -v` |
| Reto A | `practica/dojo-api` | `go test -tags reto_seguridad ./src/infrastructure/app/ -run Reto -v` |
| Reto B | `practica/dojo-api` | `go test -race -tags reto_desactivar ./...` |
| Reto C | `practica/dojo-api` | `go test -race -tags reto_caja ./...` |
| Siempre | el módulo Go tocado | `gofmt -l .` · `go vet ./...` (con el mismo `-tags`) · `go test ./...` (sin tags: la base debe seguir en verde) |

En dojo-api, si falta `docs/`, corre `swag init -g main.go` antes y menciónalo. Para B y C además confirma que `swag init` no da errores y que el endpoint aparece en `docs/swagger.json`.

## 2. Revisión de código

Lee los archivos que la persona cambió. Usa los criterios de un equipo backend Go profesional (los de "Revisar código" del `CLAUDE.md` del dojo, más: errores envueltos con contexto y códigos HTTP correctos, `ctx` propagado, dependencias hacia adentro en hexagonal, contrato y Swagger al día, tests table-driven con casos de error, sin secretos ni valores reales en código o `example:`) y la rúbrica de `modulos/12-proyecto-final.md`. Para retos que tocan Postgres, revisa también el adaptador `postgres` aunque los tests usen memoria.

## 3. Formato de la respuesta

1. **Veredicto:** ✅ aprobado / 🔁 casi / ❌ falta. Una línea.
2. **Bien hecho:** 2–3 puntos concretos.
3. **Hallazgos**, del más grave al menor:
   `archivo:línea — qué pasa — por qué importa en un equipo real (producción, PR, mantenimiento) — pregunta que lleve a corregirlo`.
4. **Para el PR:** una frase de cómo lo describiría en la sección "Cómo se probó".
5. Si está aprobado: ofrece marcarlo con `/progreso`.

## Reglas

- No reescribas su código ni pegues la versión corregida. Puedes mostrar a lo mucho una línea para ilustrar un hallazgo.
- No compares contra `soluciones/` en voz alta; úsalo solo si necesitas confirmar un comportamiento esperado.
- Si un test del dojo parece incorrecto, dilo como hallazgo del dojo, separado de la revisión.
