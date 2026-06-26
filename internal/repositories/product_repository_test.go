package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"app/internal/models"
	"app/internal/repositories"
)

func TestProductRepository_Create(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "quantity", "created_at", "updated_at"}).
		AddRow(1, "Keyboard", "Wireless", 120.5, 10, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO products (name, description, price, quantity)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, price, quantity, created_at, updated_at`)).
		WithArgs("Keyboard", "Wireless", 120.5, 10).
		WillReturnRows(rows)

	repo := repositories.NewProductRepository(db)
	product, err := repo.Create(context.Background(), models.CreateProductRequest{
		Name:        "Keyboard",
		Description: "Wireless",
		Price:       120.5,
		Quantity:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product.ID != 1 || product.Name != "Keyboard" {
		t.Fatalf("unexpected product: %+v", product)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProductRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, price, quantity, created_at, updated_at
		FROM products
		WHERE id = $1`)).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	repo := repositories.NewProductRepository(db)
	_, err = repo.GetByID(context.Background(), 99)
	if !errors.Is(err, repositories.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProductRepository_GetAll_WithKeyword(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "quantity", "created_at", "updated_at"}).
		AddRow(1, "Mechanical Keyboard", "Wireless", 120.5, 10, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, price, quantity, created_at, updated_at
		FROM products WHERE name ILIKE $1 OR description ILIKE $1 ORDER BY id ASC`)).
		WithArgs("%keyboard%").
		WillReturnRows(rows)

	repo := repositories.NewProductRepository(db)
	products, err := repo.GetAll(context.Background(), "keyboard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 1 || products[0].Name != "Mechanical Keyboard" {
		t.Fatalf("unexpected products: %+v", products)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProductRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM products WHERE id = $1`)).
		WithArgs(5).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo := repositories.NewProductRepository(db)
	err = repo.Delete(context.Background(), 5)
	if !errors.Is(err, repositories.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestProductRepository_Update_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "quantity", "created_at", "updated_at"}).
		AddRow(1, "Updated Keyboard", "Updated", 135.0, 15, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE products
		SET name = $1, description = $2, price = $3, quantity = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING id, name, description, price, quantity, created_at, updated_at`)).
		WithArgs("Updated Keyboard", "Updated", 135.0, 15, 1).
		WillReturnRows(rows)

	repo := repositories.NewProductRepository(db)
	product, err := repo.Update(context.Background(), 1, models.UpdateProductRequest{
		Name:        "Updated Keyboard",
		Description: "Updated",
		Price:       135.0,
		Quantity:    15,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if product.Name != "Updated Keyboard" || product.Quantity != 15 {
		t.Fatalf("unexpected product: %+v", product)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
