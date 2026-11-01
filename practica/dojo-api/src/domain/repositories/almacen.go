// Package repositories define los PUERTOS: interfaces que el dominio necesita y que
// la infraestructura implementa (memoria, postgres...). Se separan en
// Reader y Writer.
package repositories

import (
	"context"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/models"
)

// AlmacenReader lee almacenes.
type AlmacenReader interface {
	// Listar devuelve los almacenes ordenados por id; con soloActivos omite los inactivos.
	Listar(ctx context.Context, soloActivos bool) ([]models.Almacen, error)
	// PorID devuelve el almacén; si no existe responde 404.
	PorID(ctx context.Context, id uint) (*models.Almacen, error)
	// PorClave devuelve el almacén con esa clave, o (nil, nil) si no existe.
	PorClave(ctx context.Context, clave string) (*models.Almacen, error)
}

// AlmacenWriter escribe almacenes.
type AlmacenWriter interface {
	// Insertar crea el almacén y le asigna ID.
	Insertar(ctx context.Context, almacen *models.Almacen) error
	// Actualizar guarda nombre y activo; la clave no cambia.
	Actualizar(ctx context.Context, almacen *models.Almacen) error
}
