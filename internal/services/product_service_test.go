package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"app/internal/models"
	"app/internal/repositories"
	"app/internal/services"
)

type mockProductRepository struct {
	createFn  func(ctx context.Context, req models.CreateProductRequest) (*models.Product, error)
	getAllFn  func(ctx context.Context, keyword string) ([]models.Product, error)
	getByIDFn func(ctx context.Context, id int) (*models.Product, error)
	updateFn  func(ctx context.Context, id int, req models.UpdateProductRequest) (*models.Product, error)
	deleteFn  func(ctx context.Context, id int) error
}

func (m *mockProductRepository) Create(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
	return m.createFn(ctx, req)
}

func (m *mockProductRepository) GetAll(ctx context.Context, keyword string) ([]models.Product, error) {
	return m.getAllFn(ctx, keyword)
}

func (m *mockProductRepository) GetByID(ctx context.Context, id int) (*models.Product, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockProductRepository) Update(ctx context.Context, id int, req models.UpdateProductRequest) (*models.Product, error) {
	return m.updateFn(ctx, id, req)
}

func (m *mockProductRepository) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}

func TestProductService_Create_ValidationErrors(t *testing.T) {
	t.Parallel()

	repo := &mockProductRepository{
		createFn: func(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
			t.Fatal("repository should not be called when validation fails")
			return nil, nil
		},
	}
	service := services.NewProductService(repo)

	tests := []struct {
		name    string
		request models.CreateProductRequest
		message string
	}{
		{
			name:    "empty name",
			request: models.CreateProductRequest{Name: "", Price: 10, Quantity: 1},
			message: "name is required",
		},
		{
			name:    "short name",
			request: models.CreateProductRequest{Name: "ab", Price: 10, Quantity: 1},
			message: "name must be at least 3 characters",
		},
		{
			name:    "zero price",
			request: models.CreateProductRequest{Name: "Keyboard", Price: 0, Quantity: 1},
			message: "price must be greater than 0",
		},
		{
			name:    "negative price",
			request: models.CreateProductRequest{Name: "Keyboard", Price: -5, Quantity: 1},
			message: "price must be greater than 0",
		},
		{
			name:    "negative quantity",
			request: models.CreateProductRequest{Name: "Keyboard", Price: 10, Quantity: -1},
			message: "quantity must be greater than or equal to 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := service.Create(context.Background(), tt.request)
			if err == nil {
				t.Fatal("expected validation error")
			}

			var validationErr *services.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected ValidationError, got %T", err)
			}
			if validationErr.Message != tt.message {
				t.Fatalf("expected message %q, got %q", tt.message, validationErr.Message)
			}
		})
	}
}

func TestProductService_Create_Success(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	expected := &models.Product{
		ID:          1,
		Name:        "Mechanical Keyboard",
		Description: "Wireless",
		Price:       120.5,
		Quantity:    10,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := &mockProductRepository{
		createFn: func(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
			if req.Name != "Mechanical Keyboard" || req.Price != 120.5 || req.Quantity != 10 {
				t.Fatalf("unexpected request: %+v", req)
			}
			return expected, nil
		},
	}
	service := services.NewProductService(repo)

	product, err := service.Create(context.Background(), models.CreateProductRequest{
		Name:        "Mechanical Keyboard",
		Description: "Wireless",
		Price:       120.5,
		Quantity:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product != expected {
		t.Fatalf("expected %+v, got %+v", expected, product)
	}
}

func TestProductService_Update_ValidationErrors(t *testing.T) {
	t.Parallel()

	repo := &mockProductRepository{
		updateFn: func(ctx context.Context, id int, req models.UpdateProductRequest) (*models.Product, error) {
			t.Fatal("repository should not be called when validation fails")
			return nil, nil
		},
	}
	service := services.NewProductService(repo)

	_, err := service.Update(context.Background(), 1, models.UpdateProductRequest{
		Name:     "ab",
		Price:    10,
		Quantity: 1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr *services.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if validationErr.Message != "name must be at least 3 characters" {
		t.Fatalf("unexpected message: %q", validationErr.Message)
	}
}

func TestProductService_GetByID_DelegatesToRepository(t *testing.T) {
	t.Parallel()

	expected := &models.Product{ID: 2, Name: "Mouse", Price: 45.99, Quantity: 5}
	repo := &mockProductRepository{
		getByIDFn: func(ctx context.Context, id int) (*models.Product, error) {
			if id != 2 {
				t.Fatalf("expected id 2, got %d", id)
			}
			return expected, nil
		},
	}
	service := services.NewProductService(repo)

	product, err := service.GetByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product != expected {
		t.Fatalf("expected %+v, got %+v", expected, product)
	}
}

func TestProductService_Delete_ReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := &mockProductRepository{
		deleteFn: func(ctx context.Context, id int) error {
			return repositories.ErrProductNotFound
		},
	}
	service := services.NewProductService(repo)

	err := service.Delete(context.Background(), 99)
	if !errors.Is(err, repositories.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}
