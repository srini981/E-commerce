package routes

import (
	"monk/controllers"

	"github.com/gin-gonic/gin"
)

func Userroutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/user/login", controllers.Login)
	incomingRoutes.POST("/user/signup", controllers.SignUp)
}
