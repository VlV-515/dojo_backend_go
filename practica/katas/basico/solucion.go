//go:build solucion

package basico

import (
	"fmt"
	"strconv"
	"strings"
)

func Saludo(nombre string) string {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return "Hola, desconocido"
	}
	return "Hola, " + nombre
}

func SumarMontos(montos []float64) float64 {
	var total float64
	for _, m := range montos {
		total += m
	}
	return total
}

func Filtrar(nums []int, pred func(int) bool) []int {
	resultado := make([]int, 0, len(nums))
	for _, n := range nums {
		if pred(n) {
			resultado = append(resultado, n)
		}
	}
	return resultado
}

func ContarPalabras(texto string) map[string]int {
	conteo := make(map[string]int)
	for _, palabra := range strings.Fields(texto) {
		conteo[strings.ToLower(palabra)]++
	}
	return conteo
}

func (p Persona) NombreCompleto() string {
	if p.Apellido == "" {
		return p.Nombre
	}
	return p.Nombre + " " + p.Apellido
}

func (p Persona) EsMayorDeEdad() bool {
	return p.Edad >= 18
}

func NuevoInventario() *Inventario {
	return &Inventario{existencias: make(map[string]int)}
}

func (i *Inventario) Agregar(clave string, cantidad int) error {
	if cantidad <= 0 {
		return ErrCantidadInvalida
	}
	i.existencias[clave] += cantidad
	return nil
}

func (i *Inventario) Retirar(clave string, cantidad int) error {
	if cantidad <= 0 {
		return ErrCantidadInvalida
	}
	hay := i.existencias[clave]
	if hay < cantidad {
		return fmt.Errorf("clave %s: pedidas %d, hay %d: %w", clave, cantidad, hay, ErrSinExistencia)
	}
	i.existencias[clave] = hay - cantidad
	return nil
}

func (i *Inventario) Existencia(clave string) int {
	return i.existencias[clave]
}

func ParsearClave(clave string) (string, int, error) {
	partes := strings.Split(clave, "-")
	if len(partes) != 2 || partes[0] == "" {
		return "", 0, fmt.Errorf("%q: %w", clave, ErrClaveInvalida)
	}
	numero, err := strconv.Atoi(partes[1])
	if err != nil || numero < 0 {
		return "", 0, fmt.Errorf("%q: %w", clave, ErrClaveInvalida)
	}
	return strings.ToUpper(partes[0]), numero, nil
}
