// Package postgres es el adaptador GORM v2 contra PostgreSQL, con el mismo patrón
// UnitOfWork + txOrDB (la transacción viaja en el context).
package postgres

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/config"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/repositories"
)

// Abrir conecta a Postgres. No toca el esquema: eso es trabajo de las migraciones.
func Abrir(cfg config.Postgres) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("conectando a postgres %s:%d/%s: %w", cfg.Host, cfg.Port, cfg.DBName, err)
	}
	return db, nil
}

type txKey struct{}

var _ repositories.UnitOfWork = (*UnitOfWork)(nil)

// UnitOfWork abre la transacción y la mete al ctx.
type UnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// Do usa db.Transaction, que hace commit/rollback solo. Nunca llames Rollback()
// sobre el *gorm.DB raíz: deja el error pegado y tumba todas las consultas
// siguientes del servicio (un error clásico en producción).
func (u *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// txOrDB devuelve la transacción del ctx si hay una; si no, la base raíz.
func txOrDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db.WithContext(ctx)
}
