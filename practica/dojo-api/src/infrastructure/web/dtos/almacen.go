// Package dtos define el contrato JSON. Los `example:` alimentan Swagger: NUNCA
// pongas valores reales (llaves, tokens, correos de clientes) en ellos.
package dtos

// AlmacenDTO es un almacén tal como sale por la API.
type AlmacenDTO struct {
	ID     uint   `json:"id" example:"1"`
	Clave  string `json:"clave" example:"ALM-01"`
	Nombre string `json:"nombre" example:"Almacén Centro"`
	Activo bool   `json:"activo" example:"true"`
}

// AlmacenAltaDTO es el cuerpo de POST /almacenes.
type AlmacenAltaDTO struct {
	Clave  string `json:"clave" binding:"required" example:"ALM-01"`
	Nombre string `json:"nombre" binding:"required" example:"Almacén Centro"`
}

// AlmacenCambioDTO es el cuerpo de PUT /almacenes/{id}.
type AlmacenCambioDTO struct {
	Nombre string `json:"nombre" binding:"required" example:"Almacén Centro Histórico"`
}

// RespuestaAlmacen documenta en Swagger la respuesta con un almacén.
type RespuestaAlmacen struct {
	Codigo  int        `json:"codigo" example:"200"`
	Mensaje string     `json:"mensaje" example:"Almacén consultado"`
	Datos   AlmacenDTO `json:"datos"`
}

// RespuestaAlmacenes documenta en Swagger la respuesta con una lista.
type RespuestaAlmacenes struct {
	Codigo  int          `json:"codigo" example:"200"`
	Mensaje string       `json:"mensaje" example:"Almacenes consultados"`
	Datos   []AlmacenDTO `json:"datos"`
}

// RespuestaError documenta en Swagger el cuerpo de un error.
type RespuestaError struct {
	Codigo  int    `json:"codigo" example:"400"`
	Mensaje string `json:"mensaje" example:"Los campos clave y nombre son requeridos"`
	Detalle string `json:"detalle" example:"Los campos clave y nombre son requeridos"`
}
