package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLogicHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		username := c.Param("username")

		c.JSON(http.StatusOK, "Hello, "+username+"! This is the logic handler.")
	}
}
