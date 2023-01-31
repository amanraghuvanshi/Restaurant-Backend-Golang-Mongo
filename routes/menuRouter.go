package routes

import (
	"restaurantms/controllers"

	"github.com/gin-gonic/gin"
)

func MenuRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/menu", controllers.GetMenu())
	incomingRoutes.GET("/menu/:menu_id", controllers.GetMenubyID())
	incomingRoutes.POST("/menu", controllers.CreateMenu())
	incomingRoutes.PATCH("/menu/:menu_id", controllers.UpdateMenu())
}
