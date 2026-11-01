package src

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReglaDeDependencia hace cumplir la arquitectura hexagonal: el dominio no
// conoce la infraestructura ni los frameworks. Si alguien rompe la
// regla, el test truena en CI antes de llegar a revisión.
func TestReglaDeDependencia(t *testing.T) {
	prohibidos := []string{
		"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure",
		"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/config",
		"github.com/gin-gonic/gin",
		"gorm.io/",
		"net/http/httptest",
	}

	err := filepath.WalkDir("domain", func(ruta string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(ruta, ".go") || strings.HasSuffix(ruta, "_test.go") {
			return err
		}
		archivo, err := parser.ParseFile(token.NewFileSet(), ruta, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range archivo.Imports {
			camino := strings.Trim(imp.Path.Value, `"`)
			for _, p := range prohibidos {
				if strings.HasPrefix(camino, p) {
					t.Errorf("%s importa %s: el dominio no puede depender de infraestructura ni frameworks", ruta, camino)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
