package routes

import (
	"restaurantms/controllers"

	"github.com/gin-gonic/gin"
)

func OrderItemsRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/orderitems", controllers.GetOrderItems())
	incomingRoutes.GET("/orderitems/:orderitems_id", controllers.GetOrderItemsbyID())
	incomingRoutes.GET("orderitems-orders/:order_id", controllers.GetOrderItemsbyOrder())
	incomingRoutes.POST("/orderitems", controllers.CreateOrderItems())
	incomingRoutes.PATCH("/orderitems/:orderitems_id", controllers.UpdateOrderItems())
}
