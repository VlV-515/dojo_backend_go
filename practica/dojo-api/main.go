package main

import (
	"context"
	"log"
	"os"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// docs/ lo genera `swag init -g main.go` y NO se versiona: en un clon limpio
	// esta línea no compila hasta que lo generes.
	_ "github.com/VlV-515/dojo_backend_go/practica/dojo-api/docs"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/config"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/app"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/database"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/persistence/postgres"
)

// @title                      Dojo API
// @version                    1.0
// @description                Servicio de práctica del Dojo Backend Go: catálogo de almacenes.
// @BasePath                   /
// @securityDefinitions.apikey BearerToken
// @in                         header
// @name                       Authorization
// @description                Escribe: Bearer <token>
func main() {
	cfg, err := config.Cargar()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// CLI de migraciones: `dojo-api migrate <status|up>` corre y termina.
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if cfg.Persistencia != "postgres" {
			log.Fatal("migrate necesita persistencia postgres en el config")
		}
		db, err := postgres.Abrir(cfg.Postgres)
		if err != nil {
			log.Fatal(err)
		}
		if err := database.RunMigrateCLI(os.Args[2:], db, os.Stdout); err != nil {
			log.Fatal(err)
		}
		return
	}

	ctx := context.Background()
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("armando la app: %v", err)
	}
	if err := application.RunMigrations(ctx); err != nil {
		log.Fatalf("migraciones: %v", err)
	}

	// Swagger se registra ANTES de ArmarRouter para que quede público.
	application.Router().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	application.ArmarRouter()

	log.Printf("[arranque] %s escuchando en :%d (persistencia: %s)", cfg.Nombre, cfg.Puerto, cfg.Persistencia)
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
