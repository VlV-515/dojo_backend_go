//go:build !solucion

package basico

// ---------------------------------------------------------------------------
// Kata 1. Strings y condicionales
// ---------------------------------------------------------------------------

// Saludo devuelve "Hola, <nombre>". Si el nombre viene vacío o solo con
// espacios devuelve "Hola, desconocido". Quita los espacios de los extremos.
// Pista: paquete strings.
func Saludo(nombre string) string {
	// TODO: tu código aquí
	return ""
}

// ---------------------------------------------------------------------------
// Kata 2. Slices y ciclos
// ---------------------------------------------------------------------------

// SumarMontos suma todos los montos. Un slice vacío o nil suma 0.
func SumarMontos(montos []float64) float64 {
	// TODO: tu código aquí
	return 0
}

// Filtrar devuelve un slice NUEVO solo con los números que cumplen pred.
// Nunca devuelve nil: si nada cumple, devuelve un slice vacío ([]int{}).
// Pista: las funciones son valores; pred se llama como cualquier función.
func Filtrar(nums []int, pred func(int) bool) []int {
	// TODO: tu código aquí
	return nil
}

// ---------------------------------------------------------------------------
// Kata 3. Maps
// ---------------------------------------------------------------------------

// ContarPalabras cuenta cuántas veces aparece cada palabra. Ignora mayúsculas
// ("Hola" y "hola" son la misma) y separa por cualquier espacio en blanco.
// Pista: strings.Fields y strings.ToLower.
func ContarPalabras(texto string) map[string]int {
	// TODO: tu código aquí
	return nil
}

// ---------------------------------------------------------------------------
// Kata 4. Structs y métodos
// ---------------------------------------------------------------------------

// NombreCompleto devuelve "Nombre Apellido". Si falta el apellido, solo el nombre.
func (p Persona) NombreCompleto() string {
	// TODO: tu código aquí
	return ""
}

// EsMayorDeEdad indica si la persona tiene 18 años o más.
func (p Persona) EsMayorDeEdad() bool {
	// TODO: tu código aquí
	return false
}

// ---------------------------------------------------------------------------
// Kata 5. Punteros, constructores y errores
// ---------------------------------------------------------------------------

// NuevoInventario crea un inventario listo para usarse (con el map inicializado).
func NuevoInventario() *Inventario {
	// TODO: tu código aquí
	return &Inventario{}
}

// Agregar suma cantidad a la existencia de clave.
// Si cantidad <= 0 devuelve ErrCantidadInvalida y no cambia nada.
// Pregunta para pensar: ¿por qué el receptor es *Inventario y no Inventario?
func (i *Inventario) Agregar(clave string, cantidad int) error {
	// TODO: tu código aquí
	return nil
}

// Retirar resta cantidad de la existencia de clave.
// Si cantidad <= 0 devuelve ErrCantidadInvalida.
// Si no alcanza la existencia devuelve un error que envuelve ErrSinExistencia
// con el detalle, por ejemplo: "clave TUB-1: pedidas 5, hay 2: no hay existencia suficiente".
// Pista: fmt.Errorf con %w.
func (i *Inventario) Retirar(clave string, cantidad int) error {
	// TODO: tu código aquí
	return nil
}

// Existencia devuelve la existencia de clave (0 si no existe).
func (i *Inventario) Existencia(clave string) int {
	// TODO: tu código aquí
	return 0
}

// ---------------------------------------------------------------------------
// Kata 6. Varios valores de retorno y parseo
// ---------------------------------------------------------------------------

// ParsearClave separa claves como "ALM-0012" en ("ALM", 12, nil).
// Reglas:
//   - Debe haber exactamente un guion.
//   - El prefijo no puede estar vacío y se devuelve en MAYÚSCULAS ("alm-1" -> "ALM").
//   - El número debe ser un entero >= 0.
//
// Cualquier otra cosa devuelve ("", 0, error que envuelve ErrClaveInvalida).
// Pista: strings.Split y strconv.Atoi.
func ParsearClave(clave string) (prefijo string, numero int, err error) {
	// TODO: tu código aquí
	return "", 0, nil
}
