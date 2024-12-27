-- Example MySQL database schema with common features
CREATE DATABASE IF NOT EXISTS example_db;
USE example_db;

-- User management tables
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    date_of_birth DATE,
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login TIMESTAMP NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    metadata JSON,
    INDEX idx_user_status (status),
    INDEX idx_user_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Product catalog
CREATE TABLE categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id INT NULL,
    level INT GENERATED ALWAYS AS (
        CASE
            WHEN parent_id IS NULL THEN 0
            ELSE 1 + (SELECT level FROM categories p WHERE p.id = categories.parent_id)
        END
    ) STORED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category_id INT NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock_quantity INT UNSIGNED DEFAULT 0,
    status ENUM('draft', 'published', 'archived') DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    metadata JSON,
    FOREIGN KEY (category_id) REFERENCES categories(id),
    FULLTEXT INDEX idx_product_search (name, description)
) ENGINE=InnoDB;

-- Order management
CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    status ENUM('pending', 'processing', 'shipped', 'delivered', 'cancelled') DEFAULT 'pending',
    total_amount DECIMAL(12,2) NOT NULL,
    shipping_address TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB;

CREATE TABLE order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(12,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
) ENGINE=InnoDB;

-- Reviews and ratings
CREATE TABLE reviews (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL,
    user_id INT NOT NULL,
    rating TINYINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title VARCHAR(200),
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_verified BOOLEAN DEFAULT FALSE,
    UNIQUE KEY unique_review (product_id, user_id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB;

-- Stored procedures
DELIMITER //

CREATE PROCEDURE calculate_product_rating(IN product_id_param INT, OUT avg_rating DECIMAL(3,2))
BEGIN
    SELECT AVG(rating) INTO avg_rating
    FROM reviews
    WHERE product_id = product_id_param;
END //

CREATE PROCEDURE update_stock(
    IN product_id_param INT,
    IN quantity_param INT,
    OUT new_stock INT
)
BEGIN
    UPDATE products
    SET stock_quantity = stock_quantity + quantity_param
    WHERE id = product_id_param;
    
    SELECT stock_quantity INTO new_stock
    FROM products
    WHERE id = product_id_param;
END //

DELIMITER ;

-- Triggers
DELIMITER //

CREATE TRIGGER before_order_insert
BEFORE INSERT ON orders
FOR EACH ROW
BEGIN
    SET NEW.order_number = CONCAT('ORD', DATE_FORMAT(NOW(), '%Y%m%d'), LPAD(FLOOR(RAND() * 10000), 4, '0'));
END //

CREATE TRIGGER after_order_item_insert
AFTER INSERT ON order_items
FOR EACH ROW
BEGIN
    UPDATE products
    SET stock_quantity = stock_quantity - NEW.quantity
    WHERE id = NEW.product_id;
END //

DELIMITER ;

-- Sample data
INSERT INTO categories (name, description, parent_id) VALUES
    ('Electronics', 'Electronic devices and accessories', NULL),
    ('Computers', 'Desktop and laptop computers', 1),
    ('Smartphones', 'Mobile phones and accessories', 1),
    ('Books', 'Physical and digital books', NULL),
    ('Programming', 'Programming and technical books', 4);

INSERT INTO users (username, email, password_hash, first_name, last_name, status) VALUES
    ('john_doe', 'john@example.com', SHA2('password123', 256), 'John', 'Doe', 'active'),
    ('jane_smith', 'jane@example.com', SHA2('password456', 256), 'Jane', 'Smith', 'active');

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
    AVG(r.rating) AS avg_rating
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN reviews r ON p.id = r.product_id
GROUP BY p.id;
