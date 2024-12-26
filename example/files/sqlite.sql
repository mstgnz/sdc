-- SQLite version and settings
PRAGMA foreign_keys = ON;

-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    status TEXT DEFAULT 'active' 
        CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Products table
CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    metadata TEXT, -- JSON data
    is_active INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders table
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    status TEXT DEFAULT 'pending' 
        CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
    total_amount DECIMAL(12,2) NOT NULL CHECK (total_amount >= 0),
    shipping_address TEXT NOT NULL,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivery_date TIMESTAMP,
    metadata TEXT, -- JSON data
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Order Items table (many-to-many relationship)
CREATE TABLE order_items (
    order_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL CHECK (unit_price >= 0),
    total_price DECIMAL(10,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    PRIMARY KEY (order_id, product_id),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Order History table
CREATE TABLE order_history (
    id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    status TEXT NOT NULL 
        CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
-- Note: SQLite does not support table and column comments directly
-- We can use a separate table to store comments if needed

-- DDL Command Examples
-- Note: SQLite does not support many DDL commands that other databases have
-- For example, it does not support ALTER TABLE ADD COLUMN with constraints
-- or DROP COLUMN

-- CREATE INDEX examples
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_orders_created ON orders(created_at);

-- DROP INDEX examples (commented)
-- DROP INDEX IF EXISTS idx_users_username;
-- DROP INDEX IF EXISTS idx_products_price;
-- DROP INDEX IF EXISTS idx_orders_created;

-- CREATE VIEW example
CREATE VIEW view_order_summary AS
    SELECT u.username, COUNT(o.id) as total_orders, SUM(o.total_amount) as total_spent
    FROM users u
    LEFT JOIN orders o ON u.id = o.user_id
    GROUP BY u.username;

-- DROP VIEW example (commented)
-- DROP VIEW IF EXISTS view_order_summary;

-- CREATE TRIGGER example
CREATE TRIGGER before_product_update
    BEFORE UPDATE ON products
    FOR EACH ROW
    WHEN NEW.price < 0
BEGIN
    SELECT RAISE(ROLLBACK, 'Price cannot be negative');
END;

-- DROP TRIGGER example (commented)
-- DROP TRIGGER IF EXISTS before_product_update;

-- CREATE VIRTUAL TABLE example (FTS5)
CREATE VIRTUAL TABLE products_fts USING fts5(
    name,
    description,
    content='products',
    content_rowid='id'
);

-- DROP VIRTUAL TABLE example (commented)
-- DROP TABLE IF EXISTS products_fts;

-- CREATE TEMP TABLE example
CREATE TEMP TABLE temp_orders (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    total_amount DECIMAL(12,2),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- DROP TEMP TABLE example (commented)
-- DROP TABLE IF EXISTS temp_orders;

-- PRAGMA examples
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -2000;
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 30000000000;
PRAGMA page_size = 4096;

-- VACUUM example
VACUUM;

-- ANALYZE example
ANALYZE;

-- REINDEX example
REINDEX;
