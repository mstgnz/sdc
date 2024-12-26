-- Custom type for User Status
CREATE TYPE user_status AS TABLE (
    status NVARCHAR(20)
);
GO

-- Custom type for Order Status
CREATE TYPE order_status AS TABLE (
    status NVARCHAR(20)
);
GO

-- Users table
CREATE TABLE users (
    id BIGINT IDENTITY(1,1) PRIMARY KEY,
    username NVARCHAR(50) NOT NULL,
    email NVARCHAR(100) NOT NULL UNIQUE,
    status NVARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE()
);

-- Products table
CREATE TABLE products (
    id BIGINT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(100) NOT NULL,
    description NVARCHAR(MAX),
    price DECIMAL(10,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    metadata NVARCHAR(MAX), -- For JSON data
    is_active BIT DEFAULT 1,
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    CONSTRAINT chk_price CHECK (price >= 0),
    CONSTRAINT chk_stock CHECK (stock >= 0)
);

-- Orders table
CREATE TABLE orders (
    id BIGINT IDENTITY(1,1) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    status NVARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
    total_amount DECIMAL(12,2) NOT NULL,
    shipping_address NVARCHAR(MAX) NOT NULL,
    order_date DATETIME2 DEFAULT GETDATE(),
    delivery_date DATETIME2 NULL,
    metadata NVARCHAR(MAX), -- For JSON data
    CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_total_amount CHECK (total_amount >= 0)
);

-- Order Items table (many-to-many relationship)
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

-- Order History table (partition example)
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
EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Table storing user information',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'users';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'User email address',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'users',
    @level2type = N'COLUMN', @level2name = 'email';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Table storing product information',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'products';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Table storing order information',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'orders';

EXEC sp_addextendedproperty 
    @name = N'MS_Description', 
    @value = N'Table storing order item details',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE',  @level1name = 'order_items';

-- DDL Command Examples
-- CREATE DATABASE example
CREATE DATABASE ecommerce
    ON PRIMARY (
        NAME = ecommerce_data,
        FILENAME = 'C:\Program Files\Microsoft SQL Server\MSSQL15.SQLEXPRESS\MSSQL\DATA\ecommerce_data.mdf',
        SIZE = 100MB,
        MAXSIZE = UNLIMITED,
        FILEGROWTH = 10MB
    )
    LOG ON (
        NAME = ecommerce_log,
        FILENAME = 'C:\Program Files\Microsoft SQL Server\MSSQL15.SQLEXPRESS\MSSQL\DATA\ecommerce_log.ldf',
        SIZE = 50MB,
        MAXSIZE = UNLIMITED,
        FILEGROWTH = 5MB
    );

-- DROP DATABASE example (commented)
-- DROP DATABASE IF EXISTS ecommerce;

-- ALTER DATABASE example
ALTER DATABASE ecommerce
    SET RECOVERY SIMPLE;

-- CREATE SCHEMA example
CREATE SCHEMA app AUTHORIZATION dbo;

-- DROP SCHEMA example (commented)
-- DROP SCHEMA IF EXISTS app;

-- ALTER SCHEMA example
ALTER SCHEMA app TRANSFER dbo.users;

-- CREATE LOGIN example
CREATE LOGIN app_login 
    WITH PASSWORD = 'password',
    DEFAULT_DATABASE = ecommerce,
    CHECK_EXPIRATION = ON,
    CHECK_POLICY = ON;

-- ALTER LOGIN example
ALTER LOGIN app_login WITH PASSWORD = 'new_password';

-- DROP LOGIN example (commented)
-- DROP LOGIN app_login;

-- CREATE USER example
CREATE USER app_user 
    FOR LOGIN app_login
    WITH DEFAULT_SCHEMA = app;

-- ALTER USER example
ALTER USER app_user WITH DEFAULT_SCHEMA = dbo;

-- DROP USER example (commented)
-- DROP USER IF EXISTS app_user;

-- GRANT examples
GRANT SELECT ON SCHEMA::app TO app_user;
GRANT SELECT ON users TO app_user;
GRANT INSERT, UPDATE ON products TO app_user;

-- REVOKE examples
REVOKE INSERT, UPDATE ON products FROM app_user;

-- CREATE ROLE example
CREATE ROLE app_read_role;
GRANT SELECT ON SCHEMA::app TO app_read_role;
ALTER ROLE app_read_role ADD MEMBER app_user;

-- DROP ROLE example (commented)
-- DROP ROLE IF EXISTS app_read_role;

-- ALTER TABLE examples
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
CREATE INDEX idx_users_username 
    ON users(username)
    WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ONLINE = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON);

CREATE INDEX idx_products_price 
    ON products(price)
    WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ONLINE = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON);

CREATE INDEX idx_orders_created 
    ON orders(created_at)
    WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ONLINE = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON);

-- DROP INDEX examples (commented)
-- DROP INDEX IF EXISTS idx_users_username ON users;
-- DROP INDEX IF EXISTS idx_products_price ON products;
-- DROP INDEX IF EXISTS idx_orders_created ON orders;

-- CREATE VIEW example
CREATE OR ALTER VIEW view_order_summary AS
    SELECT u.username, COUNT(o.id) as total_orders, SUM(o.total_amount) as total_spent
    FROM users u
    LEFT JOIN orders o ON u.id = o.user_id
    GROUP BY u.username;

-- DROP VIEW example (commented)
-- DROP VIEW IF EXISTS view_order_summary;

-- CREATE TRIGGER example
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

-- DROP TRIGGER example (commented)
-- DROP TRIGGER IF EXISTS before_product_update;

-- CREATE PROCEDURE example
CREATE OR ALTER PROCEDURE get_user_orders
    @user_id INT
AS
BEGIN
    SET NOCOUNT ON;
    SELECT * FROM orders WHERE user_id = @user_id;
END;

-- DROP PROCEDURE example (commented)
-- DROP PROCEDURE IF EXISTS get_user_orders;

-- CREATE FUNCTION example
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

-- DROP FUNCTION example (commented)
-- DROP FUNCTION IF EXISTS calculate_discount;

-- CREATE TYPE example
CREATE TYPE order_status FROM VARCHAR(20) NOT NULL;

-- DROP TYPE example (commented)
-- DROP TYPE IF EXISTS order_status;

-- CREATE SEQUENCE example
CREATE SEQUENCE order_number_seq
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 999999
    CYCLE;

-- DROP SEQUENCE example (commented)
-- DROP SEQUENCE IF EXISTS order_number_seq;

-- CREATE FILEGROUP example
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
