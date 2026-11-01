// Package routes registra las rutas de cada recurso, un archivo por concepto.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/VlV-515/dojo_backend_go/practica/dojo-api/src/infrastructure/web/controllers"
)

func RutasAlmacenes(app *gin.Engine, ctrl *controllers.AlmacenController) {
	almacenes := app.Group("almacenes")
	{
		almacenes.GET("", ctrl.GetAll)
		almacenes.POST("", ctrl.Alta)
		almacenes.GET("/:id", ctrl.GetByID)
		almacenes.PUT("/:id", ctrl.Actualiza)
	}
}

func RutasReportes(app *gin.Engine, ctrl *controllers.AlmacenController) {
	reportes := app.Group("reportes")
	{
		reportes.GET("/almacenes.csv", ctrl.Exportar)
	}
}

func RutasSalud(app *gin.Engine, ctrl *controllers.SaludController) {
	app.GET("/salud", ctrl.Get)
}
