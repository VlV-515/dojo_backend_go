// Package services tiene la lógica de negocio. Depende solo de models y de los
// puertos de repositories; nunca de gin, gorm ni de infrastructure.
package services

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/pkg/apperr"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/models"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/repositories"
)

var _ AlmacenService = (*AlmacenServiceImpl)(nil)

const (
	msgAlmacenClaveYNombre = "Los campos clave y nombre son requeridos"
	msgAlmacenNombre       = "El campo nombre es requerido"
	msgAlmacenDuplicado    = "Ya existe un almacén con esa clave"
)

// AlmacenService orquesta el catálogo de almacenes.
type AlmacenService interface {
	Listar(ctx context.Context, soloActivos bool) ([]models.Almacen, error)
	Obtener(ctx context.Context, id uint) (*models.Almacen, error)
	// Registrar normaliza la clave a mayúsculas y responde 409 si ya existe.
	Registrar(ctx context.Context, almacen *models.Almacen) error
	// Actualizar solo cambia el nombre.
	Actualizar(ctx context.Context, almacen *models.Almacen) error
}

type AlmacenServiceImpl struct {
	reader repositories.AlmacenReader
	writer repositories.AlmacenWriter
	uow    repositories.UnitOfWork
}

func NewAlmacenService(reader repositories.AlmacenReader, writer repositories.AlmacenWriter, uow repositories.UnitOfWork) *AlmacenServiceImpl {
	return &AlmacenServiceImpl{reader: reader, writer: writer, uow: uow}
}

// badRequest arma un 400 cuya causa repite el mensaje (nunca New con causa nil).
func badRequest(msg string) error {
	return apperr.New(http.StatusBadRequest, msg, errors.New(msg))
}

func (s *AlmacenServiceImpl) Listar(ctx context.Context, soloActivos bool) ([]models.Almacen, error) {
	almacenes, err := s.reader.Listar(ctx, soloActivos)
	if err != nil {
		return nil, apperr.Wrap(err)
	}
	return almacenes, nil
}

func (s *AlmacenServiceImpl) Obtener(ctx context.Context, id uint) (*models.Almacen, error) {
	almacen, err := s.reader.PorID(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err)
	}
	return almacen, nil
}

// Registrar valida y revisa la clave dentro de la misma transacción del alta, para
// que la consulta y el insert vean el mismo estado.
func (s *AlmacenServiceImpl) Registrar(ctx context.Context, almacen *models.Almacen) error {
	almacen.Clave = strings.ToUpper(strings.TrimSpace(almacen.Clave))
	almacen.Nombre = strings.TrimSpace(almacen.Nombre)
	if almacen.Clave == "" || almacen.Nombre == "" {
		return badRequest(msgAlmacenClaveYNombre)
	}
	almacen.Activo = true

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		existente, err := s.reader.PorClave(ctx, almacen.Clave)
		if err != nil {
			return err
		}
		if existente != nil {
			return apperr.New(http.StatusConflict, msgAlmacenDuplicado, errors.New("clave repetida: "+almacen.Clave))
		}
		return s.writer.Insertar(ctx, almacen)
	})
	if err != nil {
		return apperr.Wrap(err)
	}
	return nil
}

func (s *AlmacenServiceImpl) Actualizar(ctx context.Context, almacen *models.Almacen) error {
	nombre := strings.TrimSpace(almacen.Nombre)
	if nombre == "" {
		return badRequest(msgAlmacenNombre)
	}
	guardado, err := s.reader.PorID(ctx, almacen.ID)
	if err != nil {
		return apperr.Wrap(err)
	}
	guardado.Nombre = nombre
	if err := s.writer.Actualizar(ctx, guardado); err != nil {
		return apperr.Wrap(err)
	}
	*almacen = *guardado
	return nil
}
