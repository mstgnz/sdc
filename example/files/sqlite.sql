-- SQLite sürüm ve ayarlar
PRAGMA foreign_keys = ON;

-- Users tablosu
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    status TEXT DEFAULT 'active' 
        CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Products tablosu
CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    metadata TEXT, -- JSON verisi için
    is_active INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders tablosu
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    status TEXT DEFAULT 'pending' 
        CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
    total_amount DECIMAL(12,2) NOT NULL CHECK (total_amount >= 0),
    shipping_address TEXT NOT NULL,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivery_date TIMESTAMP,
    metadata TEXT, -- JSON verisi için
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Order Items tablosu (çoka-çok ilişki)
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

-- Order History tablosu
CREATE TABLE order_history (
    id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    status TEXT NOT NULL 
        CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- İndeksler
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_product ON order_items(product_id);

-- View örneği
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

-- Trigger örneği (updated_at alanını güncellemek için)
CREATE TRIGGER users_update_trigger
AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;

CREATE TRIGGER products_update_trigger
AFTER UPDATE ON products
BEGIN
    UPDATE products SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;

-- DDL Komut Örnekleri
-- SQLite'da veritabanı dosya tabanlıdır, CREATE DATABASE komutu yoktur
-- Veritabanı oluşturmak için sqlite3 komutu kullanılır:
-- $ sqlite3 ecommerce.db

-- ATTACH DATABASE örneği
ATTACH DATABASE 'archive.db' AS archive;

-- DETACH DATABASE örneği
DETACH DATABASE archive;

-- ALTER TABLE örnekleri
ALTER TABLE users 
    ADD COLUMN phone TEXT;

ALTER TABLE users 
    ADD COLUMN address TEXT;

ALTER TABLE users
    RENAME TO users_new;

ALTER TABLE products 
    RENAME COLUMN name TO product_name;

-- DROP TABLE örnekleri (yorum satırı olarak)
-- DROP TABLE IF EXISTS order_items;
-- DROP TABLE IF EXISTS orders;
-- DROP TABLE IF EXISTS products;
-- DROP TABLE IF EXISTS users;

-- CREATE INDEX örnekleri
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_orders_created ON orders(created_at);

-- DROP INDEX örnekleri (yorum satırı olarak)
-- DROP INDEX IF EXISTS idx_users_username;
-- DROP INDEX IF EXISTS idx_products_price;
-- DROP INDEX IF EXISTS idx_orders_created;

-- CREATE VIEW örneği
CREATE VIEW view_order_summary AS
    SELECT u.username, COUNT(o.id) as total_orders, SUM(o.total_amount) as total_spent
    FROM users u
    LEFT JOIN orders o ON u.id = o.user_id
    GROUP BY u.username;

-- DROP VIEW örneği (yorum satırı olarak)
-- DROP VIEW IF EXISTS view_order_summary;

-- CREATE TRIGGER örneği
CREATE TRIGGER before_product_update
    BEFORE UPDATE ON products
    FOR EACH ROW
    WHEN NEW.price < 0
BEGIN
    SELECT RAISE(ROLLBACK, 'Price cannot be negative');
END;

-- DROP TRIGGER örneği (yorum satırı olarak)
-- DROP TRIGGER IF EXISTS before_product_update;

-- CREATE VIRTUAL TABLE örneği (FTS5)
CREATE VIRTUAL TABLE products_fts USING fts5(
    name,
    description,
    content='products',
    content_rowid='id'
);

-- DROP VIRTUAL TABLE örneği (yorum satırı olarak)
-- DROP TABLE IF EXISTS products_fts;

-- CREATE TEMP TABLE örneği
CREATE TEMP TABLE temp_orders (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    total_amount DECIMAL(12,2),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- DROP TEMP TABLE örneği (yorum satırı olarak)
-- DROP TABLE IF EXISTS temp_orders;

-- PRAGMA örnekleri
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -2000;
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 30000000000;
PRAGMA page_size = 4096;

-- VACUUM örneği
VACUUM;

-- ANALYZE örneği
ANALYZE;

-- REINDEX örneği
REINDEX;
