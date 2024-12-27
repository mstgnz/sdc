-- Example SQLite database schema with common features

-- Enable foreign key support
PRAGMA foreign_keys = ON;

-- User management tables
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE CHECK (
        email LIKE '%_@_%.__%' AND
        email REGEXP '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
    ),
    password_hash TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    status TEXT CHECK (status IN ('active', 'inactive', 'suspended')) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    is_admin INTEGER DEFAULT 0,
    metadata TEXT CHECK (json_valid(metadata))
);

CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_email ON users(email);

-- Product catalog
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    parent_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE CASCADE
);

-- Create a trigger to calculate category level
CREATE TRIGGER calculate_category_level
AFTER INSERT ON categories
BEGIN
    UPDATE categories
    SET level = CASE
        WHEN parent_id IS NULL THEN 0
        ELSE (SELECT level + 1 FROM categories WHERE id = NEW.parent_id)
    END
    WHERE id = NEW.id;
END;

CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock_quantity INTEGER DEFAULT 0 CHECK (stock_quantity >= 0),
    status TEXT CHECK (status IN ('draft', 'published', 'archived')) DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT CHECK (json_valid(metadata)),
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_search ON products(name, description);

-- Order management
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    order_number TEXT NOT NULL UNIQUE,
    status TEXT CHECK (status IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled')) DEFAULT 'pending',
    total_amount DECIMAL(12,2) NOT NULL,
    shipping_address TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE order_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Create a trigger to calculate total_price for order_items
CREATE TRIGGER calculate_order_item_total
AFTER INSERT ON order_items
BEGIN
    UPDATE order_items
    SET total_price = quantity * unit_price
    WHERE id = NEW.id;
END;

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_product ON order_items(product_id);

-- Reviews and ratings
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title TEXT,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_verified INTEGER DEFAULT 0,
    UNIQUE (product_id, user_id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_reviews_product ON reviews(product_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);

-- Triggers for updated_at timestamps
CREATE TRIGGER update_users_timestamp
AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_products_timestamp
AFTER UPDATE ON products
BEGIN
    UPDATE products SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_orders_timestamp
AFTER UPDATE ON orders
BEGIN
    UPDATE orders SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger for generating order numbers
CREATE TRIGGER generate_order_number
AFTER INSERT ON orders
BEGIN
    UPDATE orders
    SET order_number = 'ORD' || strftime('%Y%m%d', 'now') || substr('0000' || NEW.id, -4)
    WHERE id = NEW.id;
END;

-- Trigger for updating product stock
CREATE TRIGGER update_product_stock
AFTER INSERT ON order_items
BEGIN
    UPDATE products
    SET stock_quantity = stock_quantity - NEW.quantity
    WHERE id = NEW.product_id;
END;

-- Sample data
INSERT INTO categories (name, description, parent_id) VALUES
    ('Electronics', 'Electronic devices and accessories', NULL),
    ('Computers', 'Desktop and laptop computers', 1),
    ('Smartphones', 'Mobile phones and accessories', 1),
    ('Books', 'Physical and digital books', NULL),
    ('Programming', 'Programming and technical books', 4);

INSERT INTO users (username, email, password_hash, first_name, last_name, status) VALUES
    ('john_doe', 'john@example.com', 'hashed_password_123', 'John', 'Doe', 'active'),
    ('jane_smith', 'jane@example.com', 'hashed_password_456', 'Jane', 'Smith', 'active');

INSERT INTO products (category_id, name, description, price, stock_quantity, status) VALUES
    (2, 'Gaming Laptop', '15" Gaming Laptop with RTX 3080', 1499.99, 10, 'published'),
    (3, 'Smartphone X', 'Latest smartphone with 5G support', 999.99, 20, 'published'),
    (5, 'Python Programming', 'Complete guide to Python programming', 49.99, 100, 'published');

-- Views
CREATE VIEW product_summary AS
SELECT 
    p.id,
    p.name,
    c.name AS category,
    p.price,
    p.stock_quantity,
    COUNT(r.id) AS review_count,
    COALESCE(AVG(CAST(r.rating AS FLOAT)), 0.00) AS avg_rating
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN reviews r ON p.id = r.product_id
GROUP BY p.id, p.name, c.name, p.price, p.stock_quantity;
