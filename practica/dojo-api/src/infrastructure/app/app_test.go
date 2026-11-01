package app

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/config"
)

const tokenPrueba = "token-de-prueba"

// nuevaAppPrueba arma la app completa en memoria, como en producción pero sin red.
func nuevaAppPrueba(t *testing.T) *App {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := config.Default()
	cfg.Token = tokenPrueba
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	a.ArmarRouter()
	return a
}

// pedir hace una petición HTTP en memoria con httptest. conToken agrega el header.
func pedir(t *testing.T, a *App, metodo, ruta, cuerpo string, conToken bool) *httptest.ResponseRecorder {
	t.Helper()
	var body io.Reader
	if cuerpo != "" {
		body = strings.NewReader(cuerpo)
	}
	req := httptest.NewRequest(metodo, ruta, body)
	req.Header.Set("Content-Type", "application/json")
	if conToken {
		req.Header.Set("Authorization", "Bearer "+tokenPrueba)
	}
	rec := httptest.NewRecorder()
	a.Router().ServeHTTP(rec, req)
	return rec
}

// respuesta es la envoltura JSON de éxito y error.
type respuesta struct {
	Codigo  int             `json:"codigo"`
	Mensaje string          `json:"mensaje"`
	Datos   json.RawMessage `json:"datos"`
}

func leer(t *testing.T, rec *httptest.ResponseRecorder) respuesta {
	t.Helper()
	var r respuesta
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("respuesta no es JSON: %v\n%s", err, rec.Body.String())
	}
	return r
}

func TestSaludEsPublica(t *testing.T) {
	a := nuevaAppPrueba(t)
	if rec := pedir(t, a, http.MethodGet, "/salud", "", false); rec.Code != http.StatusOK {
		t.Fatalf("/salud sin token = %d; se esperaba 200", rec.Code)
	}
}

func TestAlmacenesExigenToken(t *testing.T) {
	a := nuevaAppPrueba(t)
	if rec := pedir(t, a, http.MethodGet, "/almacenes", "", false); rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /almacenes sin token = %d; se esperaba 401", rec.Code)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/almacenes", nil)
	req.Header.Set("Authorization", "Bearer"+tokenPrueba) // sin espacio: no debe tronar
	a.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Authorization sin espacio = %d; se esperaba 401", rec.Code)
	}
}

func TestFlujoAlmacenes(t *testing.T) {
	a := nuevaAppPrueba(t)

	rec := pedir(t, a, http.MethodPost, "/almacenes", `{"clave":"alm-01","nombre":"Centro"}`, true)
	if rec.Code != http.StatusCreated {
		t.Fatalf("alta = %d; se esperaba 201\n%s", rec.Code, rec.Body.String())
	}
	var creado struct {
		ID    uint   `json:"id"`
		Clave string `json:"clave"`
	}
	_ = json.Unmarshal(leer(t, rec).Datos, &creado)
	if creado.ID != 1 || creado.Clave != "ALM-01" {
		t.Errorf("creado = %+v; se esperaba id 1 y clave ALM-01", creado)
	}

	casos := []struct {
		nombre, metodo, ruta, cuerpo string
		codigo                       int
	}{
		{"duplicado", http.MethodPost, "/almacenes", `{"clave":"ALM-01","nombre":"Otro"}`, http.StatusConflict},
		{"JSON incompleto", http.MethodPost, "/almacenes", `{"clave":"ALM-02"}`, http.StatusBadRequest},
		{"consulta", http.MethodGet, "/almacenes/1", "", http.StatusOK},
		{"no existe", http.MethodGet, "/almacenes/99", "", http.StatusNotFound},
		{"id inválido", http.MethodGet, "/almacenes/abc", "", http.StatusBadRequest},
		{"cambio de nombre", http.MethodPut, "/almacenes/1", `{"nombre":"Centro Histórico"}`, http.StatusOK},
		{"lista", http.MethodGet, "/almacenes", "", http.StatusOK},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if rec := pedir(t, a, c.metodo, c.ruta, c.cuerpo, true); rec.Code != c.codigo {
				t.Errorf("%s %s = %d; se esperaba %d\n%s", c.metodo, c.ruta, rec.Code, c.codigo, rec.Body.String())
			}
		})
	}
}

func TestListaVaciaEsArreglo(t *testing.T) {
	a := nuevaAppPrueba(t)
	r := leer(t, pedir(t, a, http.MethodGet, "/almacenes", "", true))
	if string(r.Datos) != "[]" {
		t.Errorf("datos = %s; una lista vacía debe salir como [] y no null", r.Datos)
	}
}
