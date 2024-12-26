-- Users table
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB;

-- Products table
CREATE TABLE products (
    id BIGINT AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    metadata JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    CONSTRAINT chk_price CHECK (price >= 0),
    CONSTRAINT chk_stock CHECK (stock >= 0)
) ENGINE=InnoDB;

-- Orders table
CREATE TABLE orders (
    id BIGINT AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    status ENUM('pending', 'processing', 'completed', 'cancelled') DEFAULT 'pending',
    total_amount DECIMAL(12,2) NOT NULL,
    shipping_address TEXT NOT NULL,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivery_date TIMESTAMP NULL,
    metadata JSON,
    PRIMARY KEY (id),
    CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_total_amount CHECK (total_amount >= 0)
) ENGINE=InnoDB;

-- Order Items table (many-to-many relationship)
CREATE TABLE order_items (
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    PRIMARY KEY (order_id, product_id),
    CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) 
        REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) 
        REFERENCES products(id) ON DELETE RESTRICT,
    CONSTRAINT chk_quantity CHECK (quantity > 0),
    CONSTRAINT chk_unit_price CHECK (unit_price >= 0)
) ENGINE=InnoDB;

-- Order History table (partition instance)
CREATE TABLE order_history (
    id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    status ENUM('pending', 'processing', 'completed', 'cancelled') NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB
PARTITION BY RANGE (UNIX_TIMESTAMP(changed_at)) (
    PARTITION p_2023 VALUES LESS THAN (UNIX_TIMESTAMP('2024-01-01 00:00:00')),
    PARTITION p_2024 VALUES LESS THAN (UNIX_TIMESTAMP('2025-01-01 00:00:00'))
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_product ON order_items(product_id);

-- View example
CREATE VIEW order_summary AS
SELECT 
    o.id as order_id,
    u.username,
    o.total_amount,
    o.status,
    o.order_date,
    COUNT(oi.product_id) as total_items
FROM orders o
JOIN users u ON o.user_id = u.id
JOIN order_items oi ON o.id = oi.order_id
GROUP BY o.id, u.username, o.total_amount, o.status, o.order_date;

-- Table and column comments
ALTER TABLE users COMMENT = 'Table storing user information';
ALTER TABLE users MODIFY COLUMN email VARCHAR(100) COMMENT 'User email address';
ALTER TABLE products COMMENT = 'Table storing product information';
ALTER TABLE orders COMMENT = 'Table storing order information';
ALTER TABLE order_items COMMENT = 'Table storing order item details';

-- DDL Command Examples
-- CREATE DATABASE example
CREATE DATABASE IF NOT EXISTS ecommerce
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

-- DROP DATABASE example (commented)
-- DROP DATABASE IF EXISTS ecommerce;

-- USE DATABASE example
USE ecommerce;

-- ALTER DATABASE example
ALTER DATABASE ecommerce
    CHARACTER SET = utf8mb4
    COLLATE = utf8mb4_unicode_ci;

-- CREATE USER example
CREATE USER 'app_user'@'localhost' IDENTIFIED BY 'password';
CREATE USER 'app_user'@'%' IDENTIFIED BY 'password';

-- ALTER USER example
ALTER USER 'app_user'@'localhost'
    IDENTIFIED BY 'new_password'
    ACCOUNT UNLOCK;

-- DROP USER example (commented)
-- DROP USER 'app_user'@'localhost';
-- DROP USER 'app_user'@'%';

-- GRANT examples
GRANT ALL PRIVILEGES ON ecommerce.* TO 'app_user'@'localhost';
GRANT SELECT, INSERT, UPDATE ON ecommerce.* TO 'app_user'@'%';
GRANT SELECT ON ecommerce.users TO 'app_user'@'localhost';
GRANT INSERT, UPDATE ON ecommerce.products TO 'app_user'@'localhost';

-- REVOKE examples
REVOKE INSERT, UPDATE ON ecommerce.products FROM 'app_user'@'localhost';

-- FLUSH PRIVILEGES example
FLUSH PRIVILEGES;

-- ALTER TABLE examples
ALTER TABLE users 
    ADD COLUMN phone VARCHAR(20),
    ADD COLUMN address TEXT;

ALTER TABLE users
    ADD UNIQUE INDEX users_phone_unique (phone);

ALTER TABLE products 
    MODIFY COLUMN price DECIMAL(12,2) NOT NULL,
    MODIFY COLUMN stock INT DEFAULT 100;

ALTER TABLE products
    ADD UNIQUE INDEX products_name_unique (name);

-- DROP TABLE examples (commented)
-- DROP TABLE IF EXISTS order_items;
-- DROP TABLE IF EXISTS orders;
-- DROP TABLE IF EXISTS products;
-- DROP TABLE IF EXISTS users;

-- TRUNCATE example (commented)
-- TRUNCATE TABLE order_items;
-- TRUNCATE TABLE orders;
-- TRUNCATE TABLE products;
-- TRUNCATE TABLE users;

-- CREATE INDEX examples
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_orders_created ON orders(created_at);

-- DROP INDEX examples (commented)
-- DROP INDEX idx_users_username ON users;
-- DROP INDEX idx_products_price ON products;
-- DROP INDEX idx_orders_created ON orders;

-- ALTER INDEX example
ALTER TABLE users 
    RENAME INDEX idx_users_email TO idx_users_email_new;

-- CREATE VIEW example
CREATE OR REPLACE VIEW view_order_summary AS
    SELECT u.username, COUNT(o.id) as total_orders, SUM(o.total_amount) as total_spent
    FROM users u
    LEFT JOIN orders o ON u.id = o.user_id
    GROUP BY u.username;

-- DROP VIEW example (commented)
-- DROP VIEW IF EXISTS view_order_summary;

-- CREATE TRIGGER example
DELIMITER //
CREATE TRIGGER before_product_update
    BEFORE UPDATE ON products
    FOR EACH ROW
BEGIN
    IF NEW.price < 0 THEN
        SET NEW.price = 0;
    END IF;
END//
DELIMITER ;

-- DROP TRIGGER example (commented)
-- DROP TRIGGER IF EXISTS before_product_update;

-- CREATE EVENT example
DELIMITER //
CREATE EVENT clean_old_orders
    ON SCHEDULE EVERY 1 DAY
    DO
BEGIN
    DELETE FROM orders WHERE created_at < DATE_SUB(NOW(), INTERVAL 1 YEAR);
END//
DELIMITER ;

-- DROP EVENT example (commented)
-- DROP EVENT IF EXISTS clean_old_orders;

-- CREATE PROCEDURE example
DELIMITER //
CREATE PROCEDURE get_user_orders(IN user_id INT)
BEGIN
    SELECT * FROM orders WHERE user_id = user_id;
END//
DELIMITER ;

-- DROP PROCEDURE example (commented)
-- DROP PROCEDURE IF EXISTS get_user_orders;

-- CREATE FUNCTION example
DELIMITER //
CREATE FUNCTION calculate_discount(price DECIMAL(12,2), discount_percent INT)
    RETURNS DECIMAL(12,2)
    DETERMINISTIC
BEGIN
    RETURN price - (price * discount_percent / 100);
END//
DELIMITER ;

-- DROP FUNCTION example (commented)
-- DROP FUNCTION IF EXISTS calculate_discount;

-- Current sequence and table definitions...
