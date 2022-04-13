package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	authorizayionHeader = "crypt_key"
)

func (h *Handler) keyAuth(ctx *gin.Context) {
	authHeader := ctx.GetHeader(authorizayionHeader)

	if authHeader == "" {
		newErrorResponse(ctx, http.StatusUnauthorized, "auth header can`t be empty")
		return
	}

	if os.Getenv("CRYPT_KEY") != authHeader {
		newErrorResponse(ctx, http.StatusUnauthorized, "invalid key")
		return
	}

}
