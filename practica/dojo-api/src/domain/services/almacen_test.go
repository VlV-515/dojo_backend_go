package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/pkg/apperr"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/models"
)

// almacenFalso es un fake en memoria de los dos puertos. Registra qué escrituras
// se pidieron para que el test verifique el comportamiento, no la implementación.
type almacenFalso struct {
	guardados  map[uint]models.Almacen
	escrituras []string
	fallar     error
}

func nuevoAlmacenFalso(almacenes ...models.Almacen) *almacenFalso {
	f := &almacenFalso{guardados: map[uint]models.Almacen{}}
	for _, a := range almacenes {
		f.guardados[a.ID] = a
	}
	return f
}

func (f *almacenFalso) Listar(_ context.Context, soloActivos bool) ([]models.Almacen, error) {
	var lista []models.Almacen
	for _, a := range f.guardados {
		if !soloActivos || a.Activo {
			lista = append(lista, a)
		}
	}
	return lista, f.fallar
}

func (f *almacenFalso) PorID(_ context.Context, id uint) (*models.Almacen, error) {
	a, ok := f.guardados[id]
	if !ok {
		return nil, apperr.New(http.StatusNotFound, "El almacén no existe", errors.New("no encontrado"))
	}
	return &a, nil
}

func (f *almacenFalso) PorClave(_ context.Context, clave string) (*models.Almacen, error) {
	for _, a := range f.guardados {
		if a.Clave == clave {
			return &a, nil
		}
	}
	return nil, nil
}

func (f *almacenFalso) Insertar(_ context.Context, a *models.Almacen) error {
	f.escrituras = append(f.escrituras, "insertar")
	a.ID = uint(len(f.guardados) + 1)
	f.guardados[a.ID] = *a
	return nil
}

func (f *almacenFalso) Actualizar(_ context.Context, a *models.Almacen) error {
	f.escrituras = append(f.escrituras, "actualizar")
	f.guardados[a.ID] = *a
	return nil
}

// uowFalso corre fn sin transacción.
type uowFalso struct{ llamadas int }

func (u *uowFalso) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	u.llamadas++
	return fn(ctx)
}

func TestAlmacenRegistrar(t *testing.T) {
	casos := []struct {
		nombre      string
		entra       models.Almacen
		codigo      int // 0 = sin error
		escrituras  int
		claveFinal  string
		usaUnitWork bool
	}{
		{"válido normaliza clave", models.Almacen{Clave: " alm-02 ", Nombre: "Norte"}, 0, 1, "ALM-02", true},
		{"sin clave", models.Almacen{Nombre: "Norte"}, http.StatusBadRequest, 0, "", false},
		{"sin nombre", models.Almacen{Clave: "ALM-02", Nombre: "  "}, http.StatusBadRequest, 0, "", false},
		{"clave duplicada", models.Almacen{Clave: "alm-01", Nombre: "Otro"}, http.StatusConflict, 0, "", true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := nuevoAlmacenFalso(models.Almacen{ID: 1, Clave: "ALM-01", Nombre: "Centro", Activo: true})
			uow := &uowFalso{}
			svc := NewAlmacenService(repo, repo, uow)

			entra := c.entra
			err := svc.Registrar(context.Background(), &entra)

			if c.codigo == 0 && err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if c.codigo != 0 && !apperr.IsCode(err, c.codigo) {
				t.Fatalf("err = %v (código %d); se esperaba %d", err, apperr.Code(err), c.codigo)
			}
			if len(repo.escrituras) != c.escrituras {
				t.Errorf("escrituras = %v; se esperaban %d", repo.escrituras, c.escrituras)
			}
			if c.claveFinal != "" && entra.Clave != c.claveFinal {
				t.Errorf("clave = %q; se esperaba %q", entra.Clave, c.claveFinal)
			}
			if c.codigo == 0 && !entra.Activo {
				t.Error("un almacén nuevo debe quedar activo")
			}
			if c.usaUnitWork && uow.llamadas != 1 {
				t.Errorf("UnitOfWork.Do se llamó %d veces; se esperaba 1", uow.llamadas)
			}
		})
	}
}

func TestAlmacenActualizar(t *testing.T) {
	repo := nuevoAlmacenFalso(models.Almacen{ID: 1, Clave: "ALM-01", Nombre: "Centro", Activo: true})
	svc := NewAlmacenService(repo, repo, &uowFalso{})

	cambio := models.Almacen{ID: 1, Clave: "NO-SE-CAMBIA", Nombre: "Centro Histórico"}
	if err := svc.Actualizar(context.Background(), &cambio); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if got := repo.guardados[1]; got.Nombre != "Centro Histórico" || got.Clave != "ALM-01" {
		t.Errorf("guardado = %+v; solo debía cambiar el nombre", got)
	}

	inexistente := models.Almacen{ID: 99, Nombre: "X"}
	if err := svc.Actualizar(context.Background(), &inexistente); !apperr.IsCode(err, http.StatusNotFound) {
		t.Errorf("actualizar inexistente = %v; se esperaba 404", err)
	}

	sinNombre := models.Almacen{ID: 1}
	if err := svc.Actualizar(context.Background(), &sinNombre); !apperr.IsCode(err, http.StatusBadRequest) {
		t.Errorf("actualizar sin nombre = %v; se esperaba 400", err)
	}
}
