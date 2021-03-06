package handlers

import "github.com/gin-gonic/gin"

type Handler struct {
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	api := router.Group("/api", h.keyAuth)
	{
		api.POST("/convert", h.convertVideo)
	}

	return router
}
