package routes

import (
	"restaurantms/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/user/", controllers.GetUser())
	incomingRoutes.GET("/user/:user_id", controllers.GetUserbyID())
	incomingRoutes.POST("/user/signup", controllers.Signup())
	incomingRoutes.POST("/user/login", controllers.Login())
}
