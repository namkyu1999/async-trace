package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func GetLogicHandler(tracer trace.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := tracer.Start(c, "hello")
		defer span.End()
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		username := c.Param("username")

		c.JSON(http.StatusOK, "Hello, "+username+"! This is the logic handler.")
	}
}
