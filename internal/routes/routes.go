package routes

import (
	"github.com/gin-gonic/gin"
	"app/internal/handlers"
)

func Register(router *gin.Engine, productHandler *handlers.ProductHandler) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	products := router.Group("/products")
	{
		products.POST("", productHandler.Create)
		products.GET("", productHandler.GetAll)
		products.GET("/:id", productHandler.GetByID)
		products.PUT("/:id", productHandler.Update)
		products.DELETE("/:id", productHandler.Delete)
	}
}
