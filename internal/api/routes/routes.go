package routes

import (
	"vm-controller/internal/api/controllers"

	gin "github.com/gin-gonic/gin"
)

func SetupRouter(healthController *controllers.HealthController) *gin.Engine {
	r := gin.Default()

	// Health Check
	r.GET("/health", healthController.Check)

	// API Group
	api := r.Group("/api")
	{
		// Add future routes here
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	return r
}
