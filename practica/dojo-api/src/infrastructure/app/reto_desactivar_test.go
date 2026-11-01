//go:build reto_desactivar

// RETO B (parte 2) — Exponer PUT /almacenes/:id/desactivar (módulos 04, 05 y 08).
//
//	go test -tags reto_desactivar ./... -run Reto -v
//
// Necesitas: el método en la interfaz AlmacenService, un handler en el
// controller (con anotaciones swagger), la ruta en routes y regenerar swagger.
package app

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRetoDesactivarEndpoint(t *testing.T) {
	a := nuevaAppPrueba(t)
	pedir(t, a, http.MethodPost, "/almacenes", `{"clave":"ALM-01","nombre":"Centro"}`, true)
	pedir(t, a, http.MethodPost, "/almacenes", `{"clave":"ALM-02","nombre":"Norte"}`, true)

	if rec := pedir(t, a, http.MethodPut, "/almacenes/1/desactivar", "", false); rec.Code != http.StatusUnauthorized {
		t.Errorf("sin token = %d; se esperaba 401 (la ruta debe ir después del middleware)", rec.Code)
	}

	casos := []struct {
		nombre string
		ruta   string
		codigo int
	}{
		{"desactiva", "/almacenes/1/desactivar", http.StatusOK},
		{"otra vez", "/almacenes/1/desactivar", http.StatusConflict},
		{"no existe", "/almacenes/99/desactivar", http.StatusNotFound},
		{"id inválido", "/almacenes/x/desactivar", http.StatusBadRequest},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if rec := pedir(t, a, http.MethodPut, c.ruta, "", true); rec.Code != c.codigo {
				t.Errorf("PUT %s = %d; se esperaba %d\n%s", c.ruta, rec.Code, c.codigo, rec.Body.String())
			}
		})
	}

	var activos []struct {
		ID uint `json:"id"`
	}
	_ = json.Unmarshal(leer(t, pedir(t, a, http.MethodGet, "/almacenes?activos=true", "", true)).Datos, &activos)
	if len(activos) != 1 || activos[0].ID != 2 {
		t.Errorf("activos = %+v; se esperaba solo el almacén 2", activos)
	}
}
