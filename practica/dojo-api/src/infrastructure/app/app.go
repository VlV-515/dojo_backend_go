// Package app es el COMPOSITION ROOT: el único lugar donde se eligen adaptadores y
// se cablean repositorios -> services -> controllers.
package app

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/config"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/repositories"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/services"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/database"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/persistence/memoria"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/persistence/postgres"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/web/controllers"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/web/middleware"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/web/routes"
)

// App guarda el grafo de dependencias ya armado.
type App struct {
	cfg    config.Config
	router *gin.Engine
	db     *gorm.DB

	almacenCtrl *controllers.AlmacenController
	saludCtrl   *controllers.SaludController
}

// New elige la persistencia según el config y arma el grafo. No registra rutas.
func New(cfg config.Config) (*App, error) {
	a := &App{cfg: cfg}

	var (
		almacenReader repositories.AlmacenReader
		almacenWriter repositories.AlmacenWriter
		uow           repositories.UnitOfWork
	)

	switch cfg.Persistencia {
	case "postgres":
		db, err := postgres.Abrir(cfg.Postgres)
		if err != nil {
			return nil, err
		}
		a.db = db
		repo := postgres.NewAlmacenRepository(db)
		almacenReader, almacenWriter = repo, repo
		uow = postgres.NewUnitOfWork(db)
	case "memoria", "":
		repo := memoria.NewAlmacenRepository()
		almacenReader, almacenWriter = repo, repo
		uow = memoria.UnitOfWork{}
	default:
		return nil, fmt.Errorf("persistencia desconocida %q (usa memoria o postgres)", cfg.Persistencia)
	}

	almacenService := services.NewAlmacenService(almacenReader, almacenWriter, uow)

	a.almacenCtrl = controllers.NewAlmacenController(almacenService)
	a.saludCtrl = controllers.NewSaludController(cfg.Persistencia)

	a.router = gin.New()
	a.router.Use(gin.Logger(), gin.Recovery())
	return a, nil
}

// Router expone el *gin.Engine para que main registre lo suyo (swagger).
func (a *App) Router() *gin.Engine {
	return a.router
}

// DB es nil con persistencia en memoria.
func (a *App) DB() *gorm.DB {
	return a.db
}

// RunMigrations aplica las migraciones al arrancar solo si auto_migrate_on_boot
// es true. Contra una base compartida (pruebas/desarrollo) SIEMPRE false.
func (a *App) RunMigrations(_ context.Context) error {
	if a.db == nil || !a.cfg.AutoMigrateOnBoot {
		return nil
	}
	n, err := database.Up(a.db)
	if err != nil {
		return err
	}
	log.Printf("[migraciones] aplicadas al arrancar: %d", n)
	return nil
}

// ArmarRouter registra las rutas. EL ORDEN ES LOAD-BEARING: lo que se registra
// antes de Autenticar queda público; lo de después exige token.
func (a *App) ArmarRouter() {
	// Públicas
	routes.RutasSalud(a.router, a.saludCtrl)
	routes.RutasReportes(a.router, a.almacenCtrl)

	// ======================== AUTENTICACIÓN ========================
	a.router.Use(middleware.Autenticar(a.cfg.Token))
	// ===============================================================

	// Protegidas
	routes.RutasAlmacenes(a.router, a.almacenCtrl)
}

// Run escucha en el puerto del config.
func (a *App) Run() error {
	return a.router.Run(fmt.Sprintf(":%d", a.cfg.Puerto))
}
