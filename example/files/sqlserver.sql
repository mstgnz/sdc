-- User Status için custom tip
CREATE TYPE user_status AS TABLE (
    status NVARCHAR(20)
);
GO

-- Order Status için custom tip
CREATE TYPE order_status AS TABLE (
    status NVARCHAR(20)
);
GO

-- Users tablosu
CREATE TABLE users (
    id BIGINT IDENTITY(1,1) PRIMARY KEY,
    username NVARCHAR(50) NOT NULL,
    email NVARCHAR(100) NOT NULL UNIQUE,
    status NVARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE()
);

-- Products tablosu
CREATE TABLE products (
    id BIGINT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(100) NOT NULL,
    description NVARCHAR(MAX),
    price DECIMAL(10,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    metadata NVARCHAR(MAX), -- JSON verisi için
    is_active BIT DEFAULT 1,
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    CONSTRAINT chk_price CHECK (price >= 0),
    CONSTRAINT chk_stock CHECK (stock >= 0)
);

-- Orders tablosu
CREATE TABLE orders (
    id BIGINT IDENTITY(1,1) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    status NVARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
    total_amount DECIMAL(12,2) NOT NULL,
    shipping_address NVARCHAR(MAX) NOT NULL,
    order_date DATETIME2 DEFAULT GETDATE(),
    delivery_date DATETIME2 NULL,
    metadata NVARCHAR(MAX), -- JSON verisi için
    CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_total_amount CHECK (total_amount >= 0)
);

-- Order Items tablosu (çoka-çok ilişki)
CREATE TABLE order_items (
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price AS (quantity * unit_price) PERSISTED,
    PRIMARY KEY (order_id, product_id),
    CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) 
        REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) 
        REFERENCES products(id),
    CONSTRAINT chk_quantity CHECK (quantity > 0),
    CONSTRAINT chk_unit_price CHECK (unit_price >= 0)
);

-- Order History tablosu (partition örneği)
CREATE PARTITION FUNCTION OrderHistoryRangePF (DATETIME2)
AS RANGE RIGHT FOR VALUES ('2024-01-01');

CREATE PARTITION SCHEME OrderHistoryPS
AS PARTITION OrderHistoryRangePF ALL TO ([PRIMARY]);

CREATE TABLE order_history (
    id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    status NVARCHAR(20) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
    changed_at DATETIME2 DEFAULT GETDATE()
) ON OrderHistoryPS(changed_at);

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

-- Tablo ve kolon açıklamaları
EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Kullanıcı bilgilerinin tutulduğu tablo',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'users';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Kullanıcı email adresi',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'users',
    @level2type = N'COLUMN', @level2name = 'email';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Ürün bilgilerinin tutulduğu tablo',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'products';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Sipariş bilgilerinin tutulduğu tablo',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'orders';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Sipariş detaylarının tutulduğu tablo',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'order_items';

-- DDL Komut Örnekleri
-- CREATE DATABASE örneği
CREATE DATABASE ecommerce
    ON PRIMARY 
    (
        NAME = ecommerce_dat,
        FILENAME = 'C:\Program Files\Microsoft SQL Server\MSSQL15.SQLEXPRESS\MSSQL\DATA\ecommerce.mdf',
        SIZE = 10MB,
        MAXSIZE = UNLIMITED,
        FILEGROWTH = 5MB
    )
    LOG ON
    (
        NAME = ecommerce_log,
        FILENAME = 'C:\Program Files\Microsoft SQL Server\MSSQL15.SQLEXPRESS\MSSQL\DATA\ecommerce.ldf',
        SIZE = 5MB,
        MAXSIZE = UNLIMITED,
        FILEGROWTH = 5MB
    );

-- DROP DATABASE örneği (yorum satırı olarak)
-- DROP DATABASE IF EXISTS ecommerce;

-- ALTER DATABASE örneği
ALTER DATABASE ecommerce
    SET RECOVERY FULL;

-- USE DATABASE örneği
USE ecommerce;

-- CREATE SCHEMA örneği
CREATE SCHEMA app
    AUTHORIZATION dbo;

-- DROP SCHEMA örneği (yorum satırı olarak)
-- DROP SCHEMA IF EXISTS app;

-- ALTER SCHEMA örneği
ALTER SCHEMA app TRANSFER dbo.users;

-- CREATE LOGIN örneği
CREATE LOGIN app_login 
    WITH PASSWORD = 'password',
    DEFAULT_DATABASE = ecommerce,
    CHECK_EXPIRATION = ON,
    CHECK_POLICY = ON;

-- ALTER LOGIN örneği
ALTER LOGIN app_login WITH PASSWORD = 'new_password';

-- DROP LOGIN örneği (yorum satırı olarak)
-- DROP LOGIN app_login;

-- CREATE USER örneği
CREATE USER app_user 
    FOR LOGIN app_login
    WITH DEFAULT_SCHEMA = app;

-- ALTER USER örneği
ALTER USER app_user WITH DEFAULT_SCHEMA = dbo;

-- DROP USER örneği (yorum satırı olarak)
-- DROP USER IF EXISTS app_user;

-- GRANT örnekleri
GRANT SELECT ON SCHEMA::app TO app_user;
GRANT SELECT ON users TO app_user;
GRANT INSERT, UPDATE ON products TO app_user;

-- REVOKE örnekleri
REVOKE INSERT, UPDATE ON products FROM app_user;

-- CREATE ROLE örneği
CREATE ROLE app_read_role;
GRANT SELECT ON SCHEMA::app TO app_read_role;
ALTER ROLE app_read_role ADD MEMBER app_user;

-- DROP ROLE örneği (yorum satırı olarak)
-- DROP ROLE IF EXISTS app_read_role;

-- ALTER TABLE örnekleri
ALTER TABLE users 
    ADD phone VARCHAR(20),
    address NVARCHAR(MAX);

ALTER TABLE users
    ADD CONSTRAINT users_phone_unique UNIQUE (phone);

ALTER TABLE products 
    ALTER COLUMN price DECIMAL(12,2) NOT NULL;

ALTER TABLE products
    ADD CONSTRAINT products_name_unique UNIQUE (name),
    CONSTRAINT df_stock DEFAULT 100 FOR stock;

-- DROP TABLE örnekleri (yorum satırı olarak)
-- DROP TABLE IF EXISTS order_items;
-- DROP TABLE IF EXISTS orders;
-- DROP TABLE IF EXISTS products;
-- DROP TABLE IF EXISTS users;

-- TRUNCATE örneği (yorum satırı olarak)
-- TRUNCATE TABLE order_items;
-- TRUNCATE TABLE orders;
-- TRUNCATE TABLE products;
-- TRUNCATE TABLE users;

-- CREATE INDEX örnekleri
CREATE INDEX idx_users_username 
    ON users(username)
    WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ONLINE = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON);

CREATE INDEX idx_products_price 
    ON products(price)
    WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ONLINE = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON);

CREATE INDEX idx_orders_created 
    ON orders(created_at)
    WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ONLINE = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON);

-- DROP INDEX örnekleri (yorum satırı olarak)
-- DROP INDEX IF EXISTS idx_users_username ON users;
-- DROP INDEX IF EXISTS idx_products_price ON products;
-- DROP INDEX IF EXISTS idx_orders_created ON orders;

-- CREATE VIEW örneği
CREATE OR ALTER VIEW view_order_summary AS
    SELECT u.username, COUNT(o.id) as total_orders, SUM(o.total_amount) as total_spent
    FROM users u
    LEFT JOIN orders o ON u.id = o.user_id
    GROUP BY u.username;

-- DROP VIEW örneği (yorum satırı olarak)
-- DROP VIEW IF EXISTS view_order_summary;

-- CREATE TRIGGER örneği
CREATE OR ALTER TRIGGER before_product_update
    ON products
    AFTER UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF EXISTS (SELECT 1 FROM inserted WHERE price < 0)
    BEGIN
        UPDATE p
        SET price = 0
        FROM products p
        INNER JOIN inserted i ON p.id = i.id
        WHERE i.price < 0;
    END
END;

-- DROP TRIGGER örneği (yorum satırı olarak)
-- DROP TRIGGER IF EXISTS before_product_update;

-- CREATE PROCEDURE örneği
CREATE OR ALTER PROCEDURE get_user_orders
    @user_id INT
AS
BEGIN
    SET NOCOUNT ON;
    SELECT * FROM orders WHERE user_id = @user_id;
END;

-- DROP PROCEDURE örneği (yorum satırı olarak)
-- DROP PROCEDURE IF EXISTS get_user_orders;

-- CREATE FUNCTION örneği
CREATE OR ALTER FUNCTION calculate_discount
(
    @price DECIMAL(12,2),
    @discount_percent INT
)
RETURNS DECIMAL(12,2)
AS
BEGIN
    RETURN @price - (@price * @discount_percent / 100);
END;

-- DROP FUNCTION örneği (yorum satırı olarak)
-- DROP FUNCTION IF EXISTS calculate_discount;

-- CREATE TYPE örneği
CREATE TYPE order_status FROM VARCHAR(20) NOT NULL;

-- DROP TYPE örneği (yorum satırı olarak)
-- DROP TYPE IF EXISTS order_status;

-- CREATE SEQUENCE örneği
CREATE SEQUENCE order_number_seq
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 999999
    CYCLE;

-- DROP SEQUENCE örneği (yorum satırı olarak)
-- DROP SEQUENCE IF EXISTS order_number_seq;

-- CREATE FILEGROUP örneği
ALTER DATABASE ecommerce
    ADD FILEGROUP ecommerce_archive;

ALTER DATABASE ecommerce
    ADD FILE 
    (
        NAME = ecommerce_archive,
        FILENAME = 'C:\Program Files\Microsoft SQL Server\MSSQL15.SQLEXPRESS\MSSQL\DATA\ecommerce_archive.ndf',
        SIZE = 5MB,
        MAXSIZE = UNLIMITED,
        FILEGROWTH = 5MB
    )
    TO FILEGROUP ecommerce_archive;

-- Mevcut sequence ve tablo tanımlamaları...
// ... existing code ...