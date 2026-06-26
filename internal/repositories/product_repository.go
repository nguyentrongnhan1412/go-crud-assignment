package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"app/internal/models"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
	query := `
		INSERT INTO products (name, description, price, quantity)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, price, quantity, created_at, updated_at`

	var product models.Product
	err := r.db.QueryRowContext(ctx, query, req.Name, req.Description, req.Price, req.Quantity).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Quantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetAll(ctx context.Context, keyword string) ([]models.Product, error) {
	query := `
		SELECT id, name, description, price, quantity, created_at, updated_at
		FROM products`

	args := []any{}
	if keyword != "" {
		query += ` WHERE name ILIKE $1 OR description ILIKE $1`
		args = append(args, "%"+strings.TrimSpace(keyword)+"%")
	}
	query += ` ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Quantity,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int) (*models.Product, error) {
	query := `
		SELECT id, name, description, price, quantity, created_at, updated_at
		FROM products
		WHERE id = $1`

	var product models.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Quantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) Update(ctx context.Context, id int, req models.UpdateProductRequest) (*models.Product, error) {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, quantity = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING id, name, description, price, quantity, created_at, updated_at`

	var product models.Product
	err := r.db.QueryRowContext(ctx, query, req.Name, req.Description, req.Price, req.Quantity, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Quantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}
