package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIStatus struct {
	Status string `json:"status"`
}

// StatusHandler returns current status of this application
func StatusHandler(c *gin.Context) {
	var status = APIStatus{Status: "up"}
	statusByte, err := json.Marshal(status)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)

	c.Writer.Write(statusByte)
}
