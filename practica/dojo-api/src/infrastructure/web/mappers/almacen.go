// Package mappers traduce entre DTOs (contrato JSON) y modelos de dominio.
package mappers

import (
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/models"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/web/dtos"
)

func ToAlmacenDTO(a models.Almacen) dtos.AlmacenDTO {
	return dtos.AlmacenDTO{ID: a.ID, Clave: a.Clave, Nombre: a.Nombre, Activo: a.Activo}
}

// ToAlmacenDTOs nunca devuelve nil: una lista vacía debe salir como [] y no null.
func ToAlmacenDTOs(almacenes []models.Almacen) []dtos.AlmacenDTO {
	resultado := make([]dtos.AlmacenDTO, 0, len(almacenes))
	for _, a := range almacenes {
		resultado = append(resultado, ToAlmacenDTO(a))
	}
	return resultado
}

func AltaToAlmacen(d dtos.AlmacenAltaDTO) models.Almacen {
	return models.Almacen{Clave: d.Clave, Nombre: d.Nombre}
}
