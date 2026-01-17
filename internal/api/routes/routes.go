package routes

import (
	"os"
	controllers "vm-controller/internal/api/controllers"

	gin "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Health Check
	controllers.GetHealthController().RegisterRoutes(r.Group("/"))

	// API Group
	api := r.Group("/api")
	controllers.GetAuthController().RegisterRoutes(api)
	controllers.GetVirtualMachineController().RegisterRoutes(api)
	controllers.GetUserController().RegisterRoutes(api)

	if os.Getenv("GIN_MODE") == "debug" {
		controllers.GetTestController().RegisterRoutes(api)
	}

	return r
}
