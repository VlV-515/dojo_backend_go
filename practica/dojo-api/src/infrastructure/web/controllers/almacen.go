// Package controllers es el borde HTTP: lee la petición, llama al service y
// responde. Aquí NO hay reglas de negocio.
package controllers

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/pkg/apperr"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/pkg/general"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/models"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/domain/services"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/web/dtos"
	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/web/mappers"
)

const (
	msgLeerDatos   = "No fue posible leer los datos enviados"
	msgIDInvalido  = "El id debe ser un número entero positivo"
	msgSinServicio = "Servicio no disponible"
)

// AlmacenController expone /almacenes.
type AlmacenController struct {
	almacenService services.AlmacenService
}

func NewAlmacenController(almacenService services.AlmacenService) *AlmacenController {
	return &AlmacenController{almacenService: almacenService}
}

// idDeRuta lee el parámetro :id. Si no sirve, ya respondió 400 y devuelve false.
func idDeRuta(ctx *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		apperr.CtrlResponse(http.StatusBadRequest, msgIDInvalido, errors.New("id inválido: "+ctx.Param("id")), ctx)
		return 0, false
	}
	return uint(id), true
}

// GetAll godoc
// @Summary     Lista almacenes
// @Tags        almacenes
// @Produce     json
// @Security    BearerToken
// @Param       activos query    bool false "Solo activos"
// @Success     200     {object} dtos.RespuestaAlmacenes
// @Failure     401     {object} dtos.RespuestaError
// @Router      /almacenes [get]
func (c *AlmacenController) GetAll(ctx *gin.Context) {
	soloActivos := ctx.Query("activos") == "true"
	almacenes, err := c.almacenService.Listar(ctx.Request.Context(), soloActivos)
	if err != nil {
		apperr.Response(err, ctx)
		return
	}
	general.Responder(http.StatusOK, "Almacenes consultados", mappers.ToAlmacenDTOs(almacenes), ctx)
}

// GetByID godoc
// @Summary     Consulta un almacén
// @Tags        almacenes
// @Produce     json
// @Security    BearerToken
// @Param       id  path     int true "ID del almacén"
// @Success     200 {object} dtos.RespuestaAlmacen
// @Failure     400 {object} dtos.RespuestaError
// @Failure     404 {object} dtos.RespuestaError
// @Router      /almacenes/{id} [get]
func (c *AlmacenController) GetByID(ctx *gin.Context) {
	id, ok := idDeRuta(ctx)
	if !ok {
		return
	}
	almacen, err := c.almacenService.Obtener(ctx.Request.Context(), id)
	if err != nil {
		apperr.Response(err, ctx)
		return
	}
	general.Responder(http.StatusOK, "Almacén consultado", mappers.ToAlmacenDTO(*almacen), ctx)
}

// Alta godoc
// @Summary     Registra un almacén
// @Tags        almacenes
// @Accept      json
// @Produce     json
// @Security    BearerToken
// @Param       almacen body     dtos.AlmacenAltaDTO true "Almacén"
// @Success     201     {object} dtos.RespuestaAlmacen
// @Failure     400     {object} dtos.RespuestaError
// @Failure     409     {object} dtos.RespuestaError
// @Router      /almacenes [post]
func (c *AlmacenController) Alta(ctx *gin.Context) {
	var req dtos.AlmacenAltaDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apperr.CtrlResponse(http.StatusBadRequest, msgLeerDatos, err, ctx)
		return
	}
	almacen := mappers.AltaToAlmacen(req)
	if err := c.almacenService.Registrar(ctx.Request.Context(), &almacen); err != nil {
		apperr.Response(err, ctx)
		return
	}
	general.Responder(http.StatusCreated, "Almacén registrado", mappers.ToAlmacenDTO(almacen), ctx)
}

// Actualiza godoc
// @Summary     Cambia el nombre de un almacén
// @Tags        almacenes
// @Accept      json
// @Produce     json
// @Security    BearerToken
// @Param       id      path     int                   true "ID del almacén"
// @Param       almacen body     dtos.AlmacenCambioDTO true "Cambio"
// @Success     200     {object} dtos.RespuestaAlmacen
// @Failure     400     {object} dtos.RespuestaError
// @Failure     404     {object} dtos.RespuestaError
// @Router      /almacenes/{id} [put]
func (c *AlmacenController) Actualiza(ctx *gin.Context) {
	id, ok := idDeRuta(ctx)
	if !ok {
		return
	}
	var req dtos.AlmacenCambioDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apperr.CtrlResponse(http.StatusBadRequest, msgLeerDatos, err, ctx)
		return
	}
	almacen := models.Almacen{ID: id, Nombre: req.Nombre}
	if err := c.almacenService.Actualizar(ctx.Request.Context(), &almacen); err != nil {
		apperr.Response(err, ctx)
		return
	}
	general.Responder(http.StatusOK, "Almacén actualizado", mappers.ToAlmacenDTO(almacen), ctx)
}

// Exportar godoc
// @Summary     Exporta los almacenes a CSV
// @Tags        reportes
// @Produce     text/csv
// @Success     200 {string} string "CSV"
// @Router      /reportes/almacenes.csv [get]
func (c *AlmacenController) Exportar(ctx *gin.Context) {
	almacenes, err := c.almacenService.Listar(ctx.Request.Context(), false)
	if err != nil {
		apperr.Response(err, ctx)
		return
	}
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", `attachment; filename="almacenes.csv"`)
	w := csv.NewWriter(ctx.Writer)
	_ = w.Write([]string{"id", "clave", "nombre", "activo"})
	for _, a := range almacenes {
		_ = w.Write([]string{strconv.FormatUint(uint64(a.ID), 10), a.Clave, a.Nombre, strconv.FormatBool(a.Activo)})
	}
	w.Flush()
}
