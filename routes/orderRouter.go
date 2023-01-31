package routes

import (
	"restaurantms/controllers"

	"github.com/gin-gonic/gin"
)

func OrderRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/order", controllers.GetOrder())
	incomingRoutes.GET("/order/:order_id", controllers.GetOrderbyID())
	incomingRoutes.POST("/order", controllers.CreateOrder())
	incomingRoutes.PATCH("/order/:order_id", controllers.UpdateOrder())
}
