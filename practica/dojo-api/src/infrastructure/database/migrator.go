// Package database aplica las migraciones SQL versionadas. Es una versión mínima
// de herramientas como golang-migrate: aquí solo hay status y up.
//
// El esquema lo gobiernan estos .sql, NO los tags gorm de los structs. Los .sql se
// embeben en el binario: un .sql nuevo no existe hasta que recompilas.
package database

import (
	"embed"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var archivos embed.FS

// Migracion es un archivo NNNN_nombre.up.sql.
type Migracion struct {
	Version int
	Nombre  string
	SQL     string
}

// Cargar lee las migraciones embebidas, ordenadas por versión.
func Cargar() ([]Migracion, error) {
	entradas, err := archivos.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	var migraciones []Migracion
	for _, e := range entradas {
		nombre := e.Name()
		if !strings.HasSuffix(nombre, ".up.sql") {
			continue
		}
		version, err := strconv.Atoi(strings.SplitN(nombre, "_", 2)[0])
		if err != nil {
			return nil, fmt.Errorf("migración %s: el nombre debe empezar con NNNN_: %w", nombre, err)
		}
		contenido, err := archivos.ReadFile("migrations/" + nombre)
		if err != nil {
			return nil, err
		}
		migraciones = append(migraciones, Migracion{Version: version, Nombre: nombre, SQL: string(contenido)})
	}
	sort.Slice(migraciones, func(i, j int) bool { return migraciones[i].Version < migraciones[j].Version })
	return migraciones, nil
}

func asegurarTabla(db *gorm.DB) error {
	return db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     integer     PRIMARY KEY,
		aplicada_en timestamptz NOT NULL DEFAULT now()
	)`).Error
}

func aplicadas(db *gorm.DB) (map[int]bool, error) {
	var versiones []int
	if err := db.Raw("SELECT version FROM schema_migrations").Scan(&versiones).Error; err != nil {
		return nil, err
	}
	resultado := make(map[int]bool, len(versiones))
	for _, v := range versiones {
		resultado[v] = true
	}
	return resultado, nil
}

// Status escribe qué migraciones están aplicadas y cuáles pendientes. Solo lee
// (salvo crear schema_migrations si no existe).
func Status(db *gorm.DB, w io.Writer) error {
	if err := asegurarTabla(db); err != nil {
		return err
	}
	migraciones, err := Cargar()
	if err != nil {
		return err
	}
	hechas, err := aplicadas(db)
	if err != nil {
		return err
	}
	pendientes := 0
	for _, m := range migraciones {
		estado := "aplicada"
		if !hechas[m.Version] {
			estado = "PENDIENTE"
			pendientes++
		}
		fmt.Fprintf(w, "%04d  %-10s %s\n", m.Version, estado, m.Nombre)
	}
	fmt.Fprintf(w, "pendientes: %d\n", pendientes)
	return nil
}

// Up aplica las pendientes, cada una en su propia transacción junto con su
// registro en schema_migrations. Devuelve cuántas aplicó.
func Up(db *gorm.DB) (int, error) {
	if err := asegurarTabla(db); err != nil {
		return 0, err
	}
	migraciones, err := Cargar()
	if err != nil {
		return 0, err
	}
	hechas, err := aplicadas(db)
	if err != nil {
		return 0, err
	}
	aplicadasAhora := 0
	for _, m := range migraciones {
		if hechas[m.Version] {
			continue
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(m.SQL).Error; err != nil {
				return err
			}
			return tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.Version).Error
		})
		if err != nil {
			return aplicadasAhora, fmt.Errorf("migración %s: %w", m.Nombre, err)
		}
		aplicadasAhora++
	}
	return aplicadasAhora, nil
}

// RunMigrateCLI atiende `dojo-api migrate <status|up>`.
func RunMigrateCLI(args []string, db *gorm.DB, w io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: dojo-api migrate <status|up>")
	}
	switch args[0] {
	case "status":
		return Status(db, w)
	case "up":
		n, err := Up(db)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "migraciones aplicadas: %d\n", n)
		return nil
	default:
		return fmt.Errorf("subcomando desconocido %q (usa status o up)", args[0])
	}
}
