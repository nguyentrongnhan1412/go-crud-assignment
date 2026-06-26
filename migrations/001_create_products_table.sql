-- Create database (run manually if needed outside Docker)
-- CREATE DATABASE product_management;

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(12, 2) NOT NULL,
    quantity INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO products (name, description, price, quantity)
VALUES
    ('Mechanical Keyboard', 'Wireless mechanical keyboard', 120.50, 10),
    ('Gaming Mouse', 'Ergonomic gaming mouse with RGB lighting', 45.99, 25),
    ('USB-C Hub', '7-in-1 USB-C hub with HDMI and SD card reader', 32.00, 15);
