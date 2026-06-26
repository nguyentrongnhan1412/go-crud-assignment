package app

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"app/config"
	"app/internal/handlers"
	"app/internal/infrastructure"
	"app/internal/repositories"
	"app/internal/routes"
	"app/internal/services"
)

type App struct {
	router *gin.Engine
	cfg    *config.Config
}

func New(cfg *config.Config) (*App, error) {
	db, err := infrastructure.NewDatabase(cfg.DatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	router := gin.Default()
	routes.Register(router, productHandler)

	return &App{
		router: router,
		cfg:    cfg,
	}, nil
}

func (a *App) Run() error {
	addr := fmt.Sprintf(":%s", a.cfg.ServerPort)
	log.Printf("server starting on %s", addr)
	return a.router.Run(addr)
}
