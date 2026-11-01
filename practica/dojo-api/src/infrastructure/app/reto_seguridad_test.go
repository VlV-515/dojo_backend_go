//go:build reto_seguridad

// RETO A — Cazador de rutas expuestas (módulo 04).
//
//	go test -tags reto_seguridad ./src/infrastructure/app/ -run Reto -v
//
// Toda ruta que no esté en la lista blanca debe responder 401 sin token.
package app

import (
	"net/http"
	"strings"
	"testing"
)

var rutasPublicas = map[string]bool{
	"GET /salud": true,
}

func TestRetoSeguridadTodasLasRutasPrivadasPidenToken(t *testing.T) {
	a := nuevaAppPrueba(t)
	for _, r := range a.Router().Routes() {
		clave := r.Method + " " + r.Path
		if rutasPublicas[clave] {
			continue
		}
		ruta := strings.ReplaceAll(r.Path, ":id", "1")
		rec := pedir(t, a, r.Method, ruta, "{}", false)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s respondió %d sin token; se esperaba 401. ¿Se registró antes del middleware?", clave, rec.Code)
		}
	}
}
