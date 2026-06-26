package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"app/internal/handlers"
	"app/internal/models"
	"app/internal/repositories"
	"app/internal/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockProductService struct {
	createFn  func(ctx context.Context, req models.CreateProductRequest) (*models.Product, error)
	getAllFn  func(ctx context.Context, keyword string) ([]models.Product, error)
	getByIDFn func(ctx context.Context, id int) (*models.Product, error)
	updateFn  func(ctx context.Context, id int, req models.UpdateProductRequest) (*models.Product, error)
	deleteFn  func(ctx context.Context, id int) error
}

func (m *mockProductService) Create(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
	return m.createFn(ctx, req)
}

func (m *mockProductService) GetAll(ctx context.Context, keyword string) ([]models.Product, error) {
	return m.getAllFn(ctx, keyword)
}

func (m *mockProductService) GetByID(ctx context.Context, id int) (*models.Product, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockProductService) Update(ctx context.Context, id int, req models.UpdateProductRequest) (*models.Product, error) {
	return m.updateFn(ctx, id, req)
}

func (m *mockProductService) Delete(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}

func setupRouter(handler *handlers.ProductHandler) *gin.Engine {
	router := gin.New()
	router.POST("/products", handler.Create)
	router.GET("/products", handler.GetAll)
	router.GET("/products/:id", handler.GetByID)
	router.PUT("/products/:id", handler.Update)
	router.DELETE("/products/:id", handler.Delete)
	return router
}

func TestProductHandler_Create_Success(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	service := &mockProductService{
		createFn: func(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
			return &models.Product{
				ID:          1,
				Name:        req.Name,
				Description: req.Description,
				Price:       req.Price,
				Quantity:    req.Quantity,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}
	handler := handlers.NewProductHandler(service)
	router := setupRouter(handler)

	body := `{"name":"Mechanical Keyboard","description":"Wireless","price":120.5,"quantity":10}`
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var product models.Product
	if err := json.Unmarshal(rec.Body.Bytes(), &product); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if product.ID != 1 || product.Name != "Mechanical Keyboard" {
		t.Fatalf("unexpected product: %+v", product)
	}
}

func TestProductHandler_Create_InvalidJSON(t *testing.T) {
	t.Parallel()

	service := &mockProductService{
		createFn: func(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
			t.Fatal("service should not be called for invalid JSON")
			return nil, nil
		},
	}
	handler := handlers.NewProductHandler(service)
	router := setupRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertMessage(t, rec, http.StatusBadRequest, "invalid request body")
}

func TestProductHandler_Create_ValidationError(t *testing.T) {
	t.Parallel()

	service := &mockProductService{
		createFn: func(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
			return nil, &services.ValidationError{Message: "name is required"}
		},
	}
	handler := handlers.NewProductHandler(service)
	router := setupRouter(handler)

	body := `{"name":"","price":10,"quantity":1}`
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertMessage(t, rec, http.StatusBadRequest, "name is required")
}

func TestProductHandler_GetAll_ReturnsEmptyArray(t *testing.T) {
	t.Parallel()

	service := &mockProductService{
		getAllFn: func(ctx context.Context, keyword string) ([]models.Product, error) {
			return nil, nil
		},
	}
	handler := handlers.NewProductHandler(service)
	router := setupRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "[]" {
		t.Fatalf("expected empty array, got %s", rec.Body.String())
	}
}

func TestProductHandler_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	service := &mockProductService{
		getByIDFn: func(ctx context.Context, id int) (*models.Product, error) {
			return nil, repositories.ErrProductNotFound
		},
	}
	handler := handlers.NewProductHandler(service)
	router := setupRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/products/99", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertMessage(t, rec, http.StatusNotFound, "product not found")
}

func TestProductHandler_GetByID_InvalidID(t *testing.T) {
	t.Parallel()

	service := &mockProductService{
		getByIDFn: func(ctx context.Context, id int) (*models.Product, error) {
			t.Fatal("service should not be called for invalid id")
			return nil, nil
		},
	}
	handler := handlers.NewProductHandler(service)
	router := setupRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/products/abc", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertMessage(t, rec, http.StatusBadRequest, "invalid product id")
}

func TestProductHandler_Delete_Success(t *testing.T) {
	t.Parallel()

	service := &mockProductService{
		deleteFn: func(ctx context.Context, id int) error {
			if id != 1 {
				t.Fatalf("expected id 1, got %d", id)
			}
			return nil
		},
	}
	handler := handlers.NewProductHandler(service)
	router := setupRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertMessage(t, rec, http.StatusOK, "product deleted successfully")
}

func TestProductHandler_GetAll_RequestTimeout(t *testing.T) {
	t.Parallel()

	service := &mockProductService{
		getAllFn: func(ctx context.Context, keyword string) ([]models.Product, error) {
			return nil, context.DeadlineExceeded
		},
	}
	handler := handlers.NewProductHandler(service)
	router := setupRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertMessage(t, rec, http.StatusGatewayTimeout, "request timeout")
}

func TestProductHandler_Update_InternalServerError(t *testing.T) {
	t.Parallel()

	service := &mockProductService{
		updateFn: func(ctx context.Context, id int, req models.UpdateProductRequest) (*models.Product, error) {
			return nil, errors.New("database unavailable")
		},
	}
	handler := handlers.NewProductHandler(service)
	router := setupRouter(handler)

	body := `{"name":"Updated Keyboard","description":"Updated","price":135,"quantity":15}`
	req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assertMessage(t, rec, http.StatusInternalServerError, "internal server error")
}

func assertMessage(t *testing.T, rec *httptest.ResponseRecorder, status int, message string) {
	t.Helper()

	if rec.Code != status {
		t.Fatalf("expected status %d, got %d body=%s", status, rec.Code, rec.Body.String())
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response["message"] != message {
		t.Fatalf("expected message %q, got %q", message, response["message"])
	}
}
