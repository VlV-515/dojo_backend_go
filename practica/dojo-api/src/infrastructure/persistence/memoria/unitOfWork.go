package memoria

import (
	"context"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/repositories"
)

var _ repositories.UnitOfWork = UnitOfWork{}

// UnitOfWork en memoria solo corre fn: NO hay rollback. Si fn escribe y luego
// falla, lo escrito se queda. Es suficiente para practicar, no para producción.
type UnitOfWork struct{}

func (UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
