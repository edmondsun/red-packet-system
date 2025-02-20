package api

import (
	"net/http"
	"red-packet-system/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GrabRedPacketHandler - API handler for grabbing a red packet
func GrabRedPacketHandler(c *gin.Context) {
	// Parse `user_id` from query parameters
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	// Parse `red_packet_id` from query parameters
	redPacketID, err := strconv.Atoi(c.Query("red_packet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid red_packet_id"})
		return
	}

	// Call service layer to execute red packet grabbing logic
	amount, err := service.GrabRedPacket(uint(userID), uint(redPacketID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the grabbed amount
	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "Red packet grabbed successfully",
			"amount":  amount,
		},
	)
}
