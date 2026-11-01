package postgres

import (
	"context"
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/pkg/apperr"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/models"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/repositories"
)

var (
	_ repositories.AlmacenReader = (*AlmacenRepositoryImpl)(nil)
	_ repositories.AlmacenWriter = (*AlmacenRepositoryImpl)(nil)
)

// AlmacenRepositoryImpl es el adaptador GORM de los puertos de almacén.
type AlmacenRepositoryImpl struct {
	db *gorm.DB
}

func NewAlmacenRepository(db *gorm.DB) *AlmacenRepositoryImpl {
	return &AlmacenRepositoryImpl{db: db}
}

func (r *AlmacenRepositoryImpl) Listar(ctx context.Context, soloActivos bool) ([]models.Almacen, error) {
	almacenes := []models.Almacen{}
	consulta := txOrDB(ctx, r.db).Order("id")
	if soloActivos {
		consulta = consulta.Where("activo = ?", true)
	}
	if err := consulta.Find(&almacenes).Error; err != nil {
		return nil, apperr.New(http.StatusInternalServerError, "No fue posible consultar almacenes debido a un error interno", err)
	}
	return almacenes, nil
}

// PorID usa First: First/Take/Last SÍ devuelven gorm.ErrRecordNotFound.
func (r *AlmacenRepositoryImpl) PorID(ctx context.Context, id uint) (*models.Almacen, error) {
	var almacen models.Almacen
	if err := txOrDB(ctx, r.db).Where("id = ?", id).First(&almacen).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(http.StatusNotFound, "El almacén no existe", err)
		}
		return nil, apperr.New(http.StatusInternalServerError, "No fue posible consultar el almacén debido a un error interno", err)
	}
	return &almacen, nil
}

// PorClave usa Find: Find NUNCA devuelve ErrRecordNotFound; si no hay filas,
// RowsAffected es 0. Aquí "no existe" no es error, es (nil, nil).
func (r *AlmacenRepositoryImpl) PorClave(ctx context.Context, clave string) (*models.Almacen, error) {
	var almacen models.Almacen
	resultado := txOrDB(ctx, r.db).Where("clave = ?", clave).Limit(1).Find(&almacen)
	if resultado.Error != nil {
		return nil, apperr.New(http.StatusInternalServerError, "No fue posible consultar el almacén por clave debido a un error interno", resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return nil, nil
	}
	return &almacen, nil
}

func (r *AlmacenRepositoryImpl) Insertar(ctx context.Context, almacen *models.Almacen) error {
	if err := txOrDB(ctx, r.db).Create(almacen).Error; err != nil {
		return apperr.New(http.StatusInternalServerError, "No fue posible crear el almacén debido a un error interno", err)
	}
	return nil
}

// Actualizar usa un map y no el struct: Updates(struct) IGNORA los campos en su
// valor cero, así que Activo=false nunca se guardaría.
func (r *AlmacenRepositoryImpl) Actualizar(ctx context.Context, almacen *models.Almacen) error {
	almacen.UpdatedAt = time.Now()
	resultado := txOrDB(ctx, r.db).Model(&models.Almacen{}).Where("id = ?", almacen.ID).
		Updates(map[string]any{"nombre": almacen.Nombre, "activo": almacen.Activo, "updated_at": almacen.UpdatedAt})
	if resultado.Error != nil {
		return apperr.New(http.StatusInternalServerError, "No fue posible actualizar el almacén debido a un error interno", resultado.Error)
	}
	if resultado.RowsAffected == 0 {
		return apperr.New(http.StatusNotFound, "El almacén no existe", errors.New("update sin filas afectadas"))
	}
	return nil
}
