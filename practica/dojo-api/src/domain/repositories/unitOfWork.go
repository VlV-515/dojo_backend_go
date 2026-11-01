package repositories

import "context"

// UnitOfWork encapsula una transacción sin filtrar *gorm.DB al dominio. Do corre fn
// en una transacción; el handle viaja en el ctx, así que los repositorios que fn
// use escriben dentro de ella sin enterarse.
//
// Dos trampas que debes conocer:
//
//   - Do COMMITEA cuando fn devuelve nil. Un `return nil` temprano no cancela
//     nada: persiste lo hecho hasta ahí.
//   - Un Do dentro de otro Do abre una transacción INDEPENDIENTE, no un savepoint.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
