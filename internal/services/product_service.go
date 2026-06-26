package services

import (
	"context"
	"errors"
	"strings"

	"app/internal/models"
)

var ErrValidation = errors.New("validation error")

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
	if err := validateProductInput(req.Name, req.Price, req.Quantity); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, req)
}

func (s *ProductService) GetAll(ctx context.Context, keyword string) ([]models.Product, error) {
	return s.repo.GetAll(ctx, keyword)
}

func (s *ProductService) GetByID(ctx context.Context, id int) (*models.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, id int, req models.UpdateProductRequest) (*models.Product, error) {
	if err := validateProductInput(req.Name, req.Price, req.Quantity); err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, id, req)
}

func (s *ProductService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func validateProductInput(name string, price float64, quantity int) error {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return &ValidationError{Message: "name is required"}
	}
	if len(trimmedName) < 3 {
		return &ValidationError{Message: "name must be at least 3 characters"}
	}
	if price <= 0 {
		return &ValidationError{Message: "price must be greater than 0"}
	}
	if quantity < 0 {
		return &ValidationError{Message: "quantity must be greater than or equal to 0"}
	}
	return nil
}
