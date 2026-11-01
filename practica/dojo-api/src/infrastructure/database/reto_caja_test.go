//go:build reto_caja

package database

import (
	"strings"
	"testing"
)

// RETO C — la tabla caja debe nacer de una migración nueva, no de AutoMigrate.
func TestRetoCajaTieneMigracion(t *testing.T) {
	migraciones, err := Cargar()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range migraciones {
		if m.Version >= 2 && strings.Contains(strings.ToLower(m.SQL), "create table") && strings.Contains(strings.ToLower(m.SQL), "caja") {
			if !strings.Contains(strings.ToLower(m.SQL), "references almacen") {
				t.Errorf("%s: caja.almacen_id debe tener llave foránea (REFERENCES almacen)", m.Nombre)
			}
			return
		}
	}
	t.Error("no hay una migración >= 0002 que cree la tabla caja")
}
