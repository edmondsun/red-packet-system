package routes

import (
	"net/http"
	"red-packet-system/api"

	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the Gin router
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Red Packet System is running!"})
	})

	// Register `/grab` endpoint
	router.GET("/grab", api.GrabRedPacketHandler)

	return router
}
