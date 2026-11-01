//go:build reto_desactivar

// RETO B (parte 1) — Desactivar un almacén en el service (módulo 08).
//
//	go test -tags reto_desactivar ./src/domain/services/ -run Reto -v
//
// Implementa (s *AlmacenServiceImpl) Desactivar(ctx, id uint) error:
//   - 404 si no existe (propaga el error del reader).
//   - 409 "El almacén ya está inactivo" si ya estaba inactivo, sin escribir.
//   - si estaba activo: Activo=false y una sola llamada a writer.Actualizar.
package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/pkg/apperr"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/models"
)

func TestRetoDesactivar(t *testing.T) {
	nuevo := func() (*almacenFalso, *AlmacenServiceImpl) {
		repo := nuevoAlmacenFalso(
			models.Almacen{ID: 1, Clave: "ALM-01", Nombre: "Centro", Activo: true},
			models.Almacen{ID: 2, Clave: "ALM-02", Nombre: "Norte", Activo: false},
		)
		return repo, NewAlmacenService(repo, repo, &uowFalso{})
	}

	t.Run("activo pasa a inactivo", func(t *testing.T) {
		repo, svc := nuevo()
		if err := svc.Desactivar(context.Background(), 1); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if repo.guardados[1].Activo {
			t.Error("el almacén sigue activo")
		}
		if len(repo.escrituras) != 1 || repo.escrituras[0] != "actualizar" {
			t.Errorf("escrituras = %v; se esperaba una sola \"actualizar\"", repo.escrituras)
		}
		if repo.guardados[1].Nombre != "Centro" {
			t.Error("desactivar no debe cambiar el nombre")
		}
	})

	t.Run("ya inactivo es 409 y no escribe", func(t *testing.T) {
		repo, svc := nuevo()
		err := svc.Desactivar(context.Background(), 2)
		if !apperr.IsCode(err, http.StatusConflict) {
			t.Fatalf("err = %v (código %d); se esperaba 409", err, apperr.Code(err))
		}
		if len(repo.escrituras) != 0 {
			t.Errorf("escrituras = %v; no debía escribir", repo.escrituras)
		}
	})

	t.Run("no existe es 404", func(t *testing.T) {
		_, svc := nuevo()
		if err := svc.Desactivar(context.Background(), 99); !apperr.IsCode(err, http.StatusNotFound) {
			t.Fatalf("err = %v (código %d); se esperaba 404", err, apperr.Code(err))
		}
	})
}
