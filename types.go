package goconverter

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type errorResponse struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status string `json:"status"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	logrus.SetOutput(os.Stderr)
	logrus.Error(message)
	c.AbortWithStatusJSON(statusCode, errorResponse{message})
}
