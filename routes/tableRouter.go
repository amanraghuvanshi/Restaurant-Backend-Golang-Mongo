package routes

import (
	"restaurantms/controllers"

	"github.com/gin-gonic/gin"
)

func TableRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/table", controllers.GetTable())
	incomingRoutes.GET("/table/:table_id", controllers.GetTablebyID())
	incomingRoutes.POST("/table", controllers.CreateTable())
	incomingRoutes.PATCH("/table/:table_id", controllers.UpdateTable())
}
