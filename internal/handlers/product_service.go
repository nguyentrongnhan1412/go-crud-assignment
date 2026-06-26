package handlers

import (
	"context"

	"app/internal/models"
)

type ProductService interface {
	Create(ctx context.Context, req models.CreateProductRequest) (*models.Product, error)
	GetAll(ctx context.Context, keyword string) ([]models.Product, error)
	GetByID(ctx context.Context, id int) (*models.Product, error)
	Update(ctx context.Context, id int, req models.UpdateProductRequest) (*models.Product, error)
	Delete(ctx context.Context, id int) error
}
