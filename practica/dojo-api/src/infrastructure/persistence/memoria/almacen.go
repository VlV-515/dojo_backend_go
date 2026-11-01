// Package memoria es un adaptador que guarda todo en memoria. Sirve para arrancar
// sin base de datos y para pruebas. Se pierde todo al reiniciar.
package memoria

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/pkg/apperr"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/models"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/repositories"
)

var (
	_ repositories.AlmacenReader = (*AlmacenRepositoryImpl)(nil)
	_ repositories.AlmacenWriter = (*AlmacenRepositoryImpl)(nil)
)

// AlmacenRepositoryImpl guarda almacenes en un map protegido por Mutex: gin atiende
// cada petición en su propia goroutine.
type AlmacenRepositoryImpl struct {
	mu        sync.Mutex
	siguiente uint
	almacenes map[uint]models.Almacen
}

func NewAlmacenRepository() *AlmacenRepositoryImpl {
	return &AlmacenRepositoryImpl{siguiente: 1, almacenes: map[uint]models.Almacen{}}
}

func (r *AlmacenRepositoryImpl) Listar(_ context.Context, soloActivos bool) ([]models.Almacen, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	lista := make([]models.Almacen, 0, len(r.almacenes))
	for _, a := range r.almacenes {
		if !soloActivos || a.Activo {
			lista = append(lista, a)
		}
	}
	sort.Slice(lista, func(i, j int) bool { return lista[i].ID < lista[j].ID })
	return lista, nil
}

func (r *AlmacenRepositoryImpl) PorID(_ context.Context, id uint) (*models.Almacen, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.almacenes[id]
	if !ok {
		return nil, apperr.New(http.StatusNotFound, "El almacén no existe", errors.New("almacén no encontrado en memoria"))
	}
	return &a, nil
}

func (r *AlmacenRepositoryImpl) PorClave(_ context.Context, clave string) (*models.Almacen, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.almacenes {
		if a.Clave == clave {
			return &a, nil
		}
	}
	return nil, nil
}

func (r *AlmacenRepositoryImpl) Insertar(_ context.Context, almacen *models.Almacen) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ahora := time.Now()
	almacen.ID = r.siguiente
	almacen.CreatedAt, almacen.UpdatedAt = ahora, ahora
	r.almacenes[almacen.ID] = *almacen
	r.siguiente++
	return nil
}

func (r *AlmacenRepositoryImpl) Actualizar(_ context.Context, almacen *models.Almacen) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	guardado, ok := r.almacenes[almacen.ID]
	if !ok {
		return apperr.New(http.StatusNotFound, "El almacén no existe", errors.New("almacén no encontrado en memoria"))
	}
	guardado.Nombre = almacen.Nombre
	guardado.Activo = almacen.Activo
	guardado.UpdatedAt = time.Now()
	r.almacenes[almacen.ID] = guardado
	*almacen = guardado
	return nil
}
