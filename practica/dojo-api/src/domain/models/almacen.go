// Package models tiene las entidades del dominio. No importa frameworks: los tags
// gorm/json son solo texto y no obligan a importar nada.
package models

import "time"

// Almacen es un almacén físico de la empresa.
type Almacen struct {
	ID        uint `gorm:"primaryKey"`
	Clave     string
	Nombre    string
	Activo    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName fija el nombre de la tabla (GORM pluralizaría a "almacens").
func (Almacen) TableName() string {
	return "almacen"
}
