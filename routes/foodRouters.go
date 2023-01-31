package routes

import (
	"restaurantms/controllers"

	"github.com/gin-gonic/gin"
)

func FoodRoutes(imcomingRoutes *gin.Engine) {
	imcomingRoutes.GET("/food", controllers.GetFoods())
	imcomingRoutes.GET("/food", controllers.GetFoodbyID())
	imcomingRoutes.POST("/food", controllers.CreateFood())
	imcomingRoutes.PATCH("/food", controllers.UpdateFood())
}
