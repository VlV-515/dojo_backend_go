//go:build reto_caja

// RETO C — Recurso nuevo de punta a punta: cajas (módulos 06, 08 y 12).
//
//	go test -tags reto_caja ./... -run Reto -v
//
// Contrato:
//
//	POST /cajas                 {"nombre":"Caja 1","almacen_id":1}  -> 201, datos {id, nombre, almacen_id, abierta:false}
//	                            nombre vacío -> 400; almacén inexistente -> 404
//	GET  /cajas?almacen_id=1    -> 200, solo las cajas de ese almacén ([] si no hay)
//	GET  /cajas                 sin almacen_id -> 400
//	Todas exigen token.
//
// Toca todas las capas: models, repositories (CajaReader/CajaWriter), services
// (CajaService valida que el almacén exista usando AlmacenReader), persistence
// (memoria Y postgres), migración 0002, dtos, mappers, controller, routes, app.go.
package app

import (
	"encoding/json"
	"net/http"
	"testing"
)

type cajaRespuesta struct {
	ID        uint   `json:"id"`
	Nombre    string `json:"nombre"`
	AlmacenID uint   `json:"almacen_id"`
	Abierta   bool   `json:"abierta"`
}

func TestRetoCajas(t *testing.T) {
	a := nuevaAppPrueba(t)
	pedir(t, a, http.MethodPost, "/almacenes", `{"clave":"ALM-01","nombre":"Centro"}`, true)
	pedir(t, a, http.MethodPost, "/almacenes", `{"clave":"ALM-02","nombre":"Norte"}`, true)

	if rec := pedir(t, a, http.MethodGet, "/cajas?almacen_id=1", "", false); rec.Code != http.StatusUnauthorized {
		t.Errorf("GET /cajas sin token = %d; se esperaba 401", rec.Code)
	}

	rec := pedir(t, a, http.MethodPost, "/cajas", `{"nombre":"Caja 1","almacen_id":1}`, true)
	if rec.Code != http.StatusCreated {
		t.Fatalf("alta de caja = %d; se esperaba 201\n%s", rec.Code, rec.Body.String())
	}
	var creada cajaRespuesta
	if err := json.Unmarshal(leer(t, rec).Datos, &creada); err != nil {
		t.Fatalf("datos de la caja: %v", err)
	}
	if creada.ID == 0 || creada.Nombre != "Caja 1" || creada.AlmacenID != 1 || creada.Abierta {
		t.Errorf("caja creada = %+v", creada)
	}

	pedir(t, a, http.MethodPost, "/cajas", `{"nombre":"Caja 2","almacen_id":1}`, true)
	pedir(t, a, http.MethodPost, "/cajas", `{"nombre":"Caja Norte","almacen_id":2}`, true)

	casos := []struct {
		nombre, metodo, ruta, cuerpo string
		codigo                       int
	}{
		{"nombre vacío", http.MethodPost, "/cajas", `{"nombre":"  ","almacen_id":1}`, http.StatusBadRequest},
		{"almacén inexistente", http.MethodPost, "/cajas", `{"nombre":"Caja X","almacen_id":99}`, http.StatusNotFound},
		{"lista sin filtro", http.MethodGet, "/cajas", "", http.StatusBadRequest},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if rec := pedir(t, a, c.metodo, c.ruta, c.cuerpo, true); rec.Code != c.codigo {
				t.Errorf("%s %s = %d; se esperaba %d\n%s", c.metodo, c.ruta, rec.Code, c.codigo, rec.Body.String())
			}
		})
	}

	var delUno []cajaRespuesta
	_ = json.Unmarshal(leer(t, pedir(t, a, http.MethodGet, "/cajas?almacen_id=1", "", true)).Datos, &delUno)
	if len(delUno) != 2 {
		t.Errorf("cajas del almacén 1 = %+v; se esperaban 2", delUno)
	}

	pedir(t, a, http.MethodPost, "/almacenes", `{"clave":"ALM-03","nombre":"Sur"}`, true)
	r := leer(t, pedir(t, a, http.MethodGet, "/cajas?almacen_id=3", "", true))
	if string(r.Datos) != "[]" {
		t.Errorf("almacén sin cajas: datos = %s; se esperaba []", r.Datos)
	}
}
