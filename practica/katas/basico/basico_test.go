package basico

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// Las pruebas de este archivo usan el patrón "table-driven": una tabla de casos
// y un ciclo con t.Run. Es el estilo estándar de tests en Go.

func TestSaludo(t *testing.T) {
	casos := []struct {
		nombre string
		entra  string
		espera string
	}{
		{"nombre normal", "Ana", "Hola, Ana"},
		{"con espacios", "  Luis  ", "Hola, Luis"},
		{"vacío", "", "Hola, desconocido"},
		{"solo espacios", "   ", "Hola, desconocido"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := Saludo(c.entra); got != c.espera {
				t.Errorf("Saludo(%q) = %q; se esperaba %q", c.entra, got, c.espera)
			}
		})
	}
}

func TestSumarMontos(t *testing.T) {
	casos := []struct {
		nombre string
		montos []float64
		espera float64
	}{
		{"nil", nil, 0},
		{"vacío", []float64{}, 0},
		{"varios", []float64{10.5, 20, 0.5}, 31},
		{"negativos", []float64{100, -40}, 60},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := SumarMontos(c.montos); got != c.espera {
				t.Errorf("SumarMontos(%v) = %v; se esperaba %v", c.montos, got, c.espera)
			}
		})
	}
}

func TestFiltrar(t *testing.T) {
	esPar := func(n int) bool { return n%2 == 0 }

	got := Filtrar([]int{1, 2, 3, 4, 5, 6}, esPar)
	if !reflect.DeepEqual(got, []int{2, 4, 6}) {
		t.Errorf("Filtrar pares = %v; se esperaba [2 4 6]", got)
	}

	vacio := Filtrar([]int{1, 3}, esPar)
	if vacio == nil {
		t.Error("Filtrar sin coincidencias devolvió nil; se esperaba []int{} (al serializar a JSON, nil sale como null y un slice vacío como [])")
	}
	if len(vacio) != 0 {
		t.Errorf("Filtrar sin coincidencias = %v; se esperaba vacío", vacio)
	}

	original := []int{2, 4}
	resultado := Filtrar(original, esPar)
	if len(resultado) > 0 {
		resultado[0] = 99
	}
	if original[0] != 2 {
		t.Error("Filtrar modificó el slice original: debes construir un slice nuevo")
	}
}

func TestContarPalabras(t *testing.T) {
	got := ContarPalabras("Hola hola  MUNDO\tmundo mundo")
	espera := map[string]int{"hola": 2, "mundo": 3}
	if !reflect.DeepEqual(got, espera) {
		t.Errorf("ContarPalabras = %v; se esperaba %v", got, espera)
	}
	if got := ContarPalabras(""); got == nil || len(got) != 0 {
		t.Errorf("ContarPalabras(\"\") = %v; se esperaba un map vacío, no nil", got)
	}
}

func TestPersona(t *testing.T) {
	casos := []struct {
		persona     Persona
		nombre      string
		mayorDeEdad bool
	}{
		{Persona{"Ana", "López", 30}, "Ana López", true},
		{Persona{"Luis", "", 18}, "Luis", true},
		{Persona{"Eva", "Ruiz", 17}, "Eva Ruiz", false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := c.persona.NombreCompleto(); got != c.nombre {
				t.Errorf("NombreCompleto() = %q; se esperaba %q", got, c.nombre)
			}
			if got := c.persona.EsMayorDeEdad(); got != c.mayorDeEdad {
				t.Errorf("EsMayorDeEdad() con edad %d = %v; se esperaba %v", c.persona.Edad, got, c.mayorDeEdad)
			}
		})
	}
}

func TestInventario(t *testing.T) {
	inv := NuevoInventario()
	if inv == nil || inv.existencias == nil {
		t.Fatal("NuevoInventario debe devolver un inventario con el map inicializado")
	}

	if err := inv.Agregar("TUB-1", 10); err != nil {
		t.Fatalf("Agregar válido devolvió error: %v", err)
	}
	if err := inv.Agregar("TUB-1", 5); err != nil {
		t.Fatalf("Agregar válido devolvió error: %v", err)
	}
	if got := inv.Existencia("TUB-1"); got != 15 {
		t.Errorf("Existencia = %d; se esperaba 15", got)
	}

	if err := inv.Agregar("TUB-1", 0); !errors.Is(err, ErrCantidadInvalida) {
		t.Errorf("Agregar(0) = %v; se esperaba ErrCantidadInvalida", err)
	}
	if err := inv.Retirar("TUB-1", -1); !errors.Is(err, ErrCantidadInvalida) {
		t.Errorf("Retirar(-1) = %v; se esperaba ErrCantidadInvalida", err)
	}

	if err := inv.Retirar("TUB-1", 4); err != nil {
		t.Fatalf("Retirar válido devolvió error: %v", err)
	}
	if got := inv.Existencia("TUB-1"); got != 11 {
		t.Errorf("Existencia tras retirar = %d; se esperaba 11", got)
	}

	err := inv.Retirar("TUB-1", 50)
	if !errors.Is(err, ErrSinExistencia) {
		t.Fatalf("Retirar de más = %v; se esperaba un error que envuelva ErrSinExistencia", err)
	}
	if err == ErrSinExistencia {
		t.Error("el error debe ENVOLVER ErrSinExistencia con detalle (fmt.Errorf con %w), no ser el centinela pelón")
	}
	if !strings.Contains(err.Error(), "TUB-1") {
		t.Errorf("el mensaje %q debería mencionar la clave", err.Error())
	}
	if got := inv.Existencia("TUB-1"); got != 11 {
		t.Errorf("un retiro fallido no debe cambiar la existencia: hay %d, se esperaba 11", got)
	}

	if got := inv.Existencia("NO-EXISTE"); got != 0 {
		t.Errorf("Existencia de clave inexistente = %d; se esperaba 0", got)
	}
}

func TestParsearClave(t *testing.T) {
	casos := []struct {
		entra   string
		prefijo string
		numero  int
		error   bool
	}{
		{"ALM-0012", "ALM", 12, false},
		{"alm-1", "ALM", 1, false},
		{"CAJ-0", "CAJ", 0, false},
		{"ALM", "", 0, true},
		{"-12", "", 0, true},
		{"ALM-", "", 0, true},
		{"ALM-12-3", "", 0, true},
		{"ALM-abc", "", 0, true},
		{"ALM--5", "", 0, true},
	}
	for _, c := range casos {
		t.Run(c.entra, func(t *testing.T) {
			prefijo, numero, err := ParsearClave(c.entra)
			if c.error {
				if !errors.Is(err, ErrClaveInvalida) {
					t.Errorf("ParsearClave(%q) err = %v; se esperaba ErrClaveInvalida", c.entra, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsearClave(%q) err inesperado: %v", c.entra, err)
			}
			if prefijo != c.prefijo || numero != c.numero {
				t.Errorf("ParsearClave(%q) = (%q, %d); se esperaba (%q, %d)", c.entra, prefijo, numero, c.prefijo, c.numero)
			}
		})
	}
}
