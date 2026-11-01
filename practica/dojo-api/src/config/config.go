// Package config carga la configuración del servicio, desde un JSON.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

// TokenInseguro es el token que se usa si no hay DOJO_TOKEN. Sirve en local,
// jamás en un servidor.
const TokenInseguro = "dojo-local-inseguro"

// Postgres son los datos de conexión. Dentro de docker compose el host es el alias
// de la red (dbpostgres:5432); desde tu Mac es lo publicado (localhost:5433).
type Postgres struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// DSN arma la cadena de conexión key=value que entiende el driver.
func (p Postgres) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode)
}

// Config es el config.json del servicio.
type Config struct {
	Nombre string `json:"nombre"`
	Puerto int    `json:"puerto"`
	// Persistencia es "memoria" (default, sin base) o "postgres".
	Persistencia      string   `json:"persistencia"`
	AutoMigrateOnBoot bool     `json:"auto_migrate_on_boot"`
	Postgres          Postgres `json:"postgres"`
	// Token no viene del archivo: sale de la variable DOJO_TOKEN (los secretos
	// van en variables de entorno, no en archivos versionados).
	Token string `json:"-"`
}

// Default es la configuración de prueba que se usa cuando no hay archivo.
func Default() Config {
	return Config{Nombre: "dojo-api", Puerto: 1099, Persistencia: "memoria"}
}

// Cargar lee el archivo de la variable CONFIG o, si no está, ./config.json.
// Si el archivo no existe NO falla: avisa y usa Default(), como hacen muchas librerías internas.
// Esa "comodidad" es la que hace que a veces creas que estás conectado a una base
// y en realidad no: lee siempre el log de arranque.
func Cargar() (Config, error) {
	ruta := os.Getenv("CONFIG")
	if ruta == "" {
		ruta = "config.json"
	}

	cfg := Default()
	contenido, err := os.ReadFile(ruta)
	switch {
	case errors.Is(err, os.ErrNotExist):
		log.Printf("[config] %s no existe: uso la configuración de prueba interna (persistencia en memoria)", ruta)
	case err != nil:
		return Config{}, fmt.Errorf("leyendo %s: %w", ruta, err)
	default:
		if err := json.Unmarshal(contenido, &cfg); err != nil {
			return Config{}, fmt.Errorf("%s no es JSON válido: %w", ruta, err)
		}
		log.Printf("[config] cargada desde %s (persistencia: %s)", ruta, cfg.Persistencia)
	}

	cfg.Token = os.Getenv("DOJO_TOKEN")
	if cfg.Token == "" {
		cfg.Token = TokenInseguro
		log.Printf("[config] DOJO_TOKEN vacío: uso el token inseguro de desarrollo %q", TokenInseguro)
	}
	return cfg, nil
}
