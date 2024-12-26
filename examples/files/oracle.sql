-- Sequence declarations
CREATE SEQUENCE users_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE products_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE orders_seq START WITH 1 INCREMENT BY 1;

-- Users table
CREATE TABLE users (
    id NUMBER DEFAULT users_seq.NEXTVAL PRIMARY KEY,
    username VARCHAR2(50) NOT NULL,
    email VARCHAR2(100) NOT NULL UNIQUE,
    status VARCHAR2(20) DEFAULT 'active' 
        CONSTRAINT users_status_chk CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Products table
CREATE TABLE products (
    id NUMBER DEFAULT products_seq.NEXTVAL PRIMARY KEY,
    name VARCHAR2(100) NOT NULL,
    description CLOB,
    price NUMBER(10,2) NOT NULL 
        CONSTRAINT products_price_chk CHECK (price >= 0),
    stock NUMBER DEFAULT 0 
        CONSTRAINT products_stock_chk CHECK (stock >= 0),
    metadata CLOB, -- JSON data
    is_active NUMBER(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Orders table
CREATE TABLE orders (
    id NUMBER DEFAULT orders_seq.NEXTVAL PRIMARY KEY,
    user_id NUMBER NOT NULL,
    status VARCHAR2(20) DEFAULT 'pending' 
        CONSTRAINT orders_status_chk CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
    total_amount NUMBER(12,2) NOT NULL 
        CONSTRAINT orders_total_amount_chk CHECK (total_amount >= 0),
    shipping_address CLOB NOT NULL,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivery_date TIMESTAMP,
    metadata CLOB, -- JSON data
    CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE
);

-- Order Items table (many-to-many relationship)
CREATE TABLE order_items (
    order_id NUMBER NOT NULL,
    product_id NUMBER NOT NULL,
    quantity NUMBER NOT NULL 
        CONSTRAINT order_items_quantity_chk CHECK (quantity > 0),
    unit_price NUMBER(10,2) NOT NULL 
        CONSTRAINT order_items_unit_price_chk CHECK (unit_price >= 0),
    total_price NUMBER(10,2) GENERATED ALWAYS AS (quantity * unit_price) VIRTUAL,
    CONSTRAINT order_items_pk PRIMARY KEY (order_id, product_id),
    CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) 
        REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) 
        REFERENCES products(id)
);

-- Order History table (partition example)
CREATE TABLE order_history (
    id NUMBER NOT NULL,
    order_id NUMBER NOT NULL,
    status VARCHAR2(20) NOT NULL 
        CONSTRAINT order_history_status_chk CHECK (status IN ('pending', 'processing', 'completed', 'cancelled')),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
PARTITION BY RANGE (changed_at) (
    PARTITION order_history_2023 VALUES LESS THAN (TIMESTAMP '2024-01-01 00:00:00'),
    PARTITION order_history_2024 VALUES LESS THAN (TIMESTAMP '2025-01-01 00:00:00')
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_product ON order_items(product_id);

-- View example
CREATE OR REPLACE VIEW order_summary AS
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
COMMENT ON TABLE users IS 'Table storing user information';
COMMENT ON COLUMN users.email IS 'User email address';
COMMENT ON TABLE products IS 'Table storing product information';
COMMENT ON TABLE orders IS 'Table storing order information';
COMMENT ON TABLE order_items IS 'Table storing order item details';

-- Trigger example (for updating updated_at field)
CREATE OR REPLACE TRIGGER users_update_trigger
    BEFORE UPDATE ON users
    FOR EACH ROW
BEGIN
    :NEW.updated_at := CURRENT_TIMESTAMP;
END;
/

CREATE OR REPLACE TRIGGER products_update_trigger
    BEFORE UPDATE ON products
    FOR EACH ROW
BEGIN
    :NEW.updated_at := CURRENT_TIMESTAMP;
END;
/

-- DDL Command Examples
-- CREATE DATABASE (as SYSDBA)
-- CREATE DATABASE ecommerce
--     USER SYS IDENTIFIED BY password
--     USER SYSTEM IDENTIFIED BY password
--     LOGFILE GROUP 1 ('path/ecommerce_log1a.rdo', 'path/ecommerce_log1b.rdo') SIZE 100M,
--             GROUP 2 ('path/ecommerce_log2a.rdo', 'path/ecommerce_log2b.rdo') SIZE 100M
--     MAXLOGFILES 5
--     MAXLOGMEMBERS 5
--     MAXLOGHISTORY 1
--     MAXDATAFILES 100
--     CHARACTER SET AL32UTF8
--     NATIONAL CHARACTER SET AL16UTF16
--     EXTENT MANAGEMENT LOCAL
--     DATAFILE 'path/ecommerce_system.dbf' SIZE 500M REUSE AUTOEXTEND ON NEXT 10M MAXSIZE UNLIMITED
--     SYSAUX DATAFILE 'path/ecommerce_sysaux.dbf' SIZE 500M REUSE AUTOEXTEND ON NEXT 10M MAXSIZE UNLIMITED
--     DEFAULT TABLESPACE users
--     DEFAULT TEMPORARY TABLESPACE temp
--     UNDO TABLESPACE undotbs1;

-- CREATE TABLESPACE example
CREATE TABLESPACE ecommerce_data
    DATAFILE 'path/ecommerce_data01.dbf' SIZE 100M REUSE
    AUTOEXTEND ON NEXT 100M MAXSIZE UNLIMITED
    LOGGING
    ONLINE
    EXTENT MANAGEMENT LOCAL
    SEGMENT SPACE MANAGEMENT AUTO;

-- DROP TABLESPACE example (commented)
-- DROP TABLESPACE ecommerce_data INCLUDING CONTENTS AND DATAFILES;

-- ALTER TABLESPACE example
ALTER TABLESPACE ecommerce_data
    ADD DATAFILE 'path/ecommerce_data02.dbf' SIZE 100M
    AUTOEXTEND ON NEXT 100M MAXSIZE UNLIMITED;

-- CREATE USER example
CREATE USER app_user IDENTIFIED BY password
    DEFAULT TABLESPACE ecommerce_data
    TEMPORARY TABLESPACE temp
    QUOTA UNLIMITED ON ecommerce_data;

-- ALTER USER example
ALTER USER app_user
    IDENTIFIED BY new_password
    ACCOUNT UNLOCK;

-- DROP USER example (commented)
-- DROP USER app_user CASCADE;

-- GRANT examples
GRANT CREATE SESSION TO app_user;
GRANT CREATE TABLE, CREATE VIEW, CREATE SEQUENCE TO app_user;
GRANT SELECT ON users TO app_user;
GRANT INSERT, UPDATE ON products TO app_user;

-- REVOKE examples
REVOKE INSERT, UPDATE ON products FROM app_user;

-- CREATE ROLE example
CREATE ROLE app_read_role;
GRANT SELECT ANY TABLE TO app_read_role;
GRANT app_read_role TO app_user;

-- DROP ROLE example (commented)
-- DROP ROLE app_read_role;

-- ALTER TABLE examples
ALTER TABLE users 
    ADD (
        phone VARCHAR2(20),
        address CLOB
    );

-- CREATE MATERIALIZED VIEW example
CREATE MATERIALIZED VIEW mv_order_summary
    BUILD IMMEDIATE
    REFRESH COMPLETE ON DEMAND
    ENABLE QUERY REWRITE
AS
    SELECT u.username, COUNT(o.id) as total_orders, SUM(o.total_amount) as total_spent
    FROM users u
    LEFT JOIN orders o ON u.id = o.user_id
    GROUP BY u.username;

-- DROP MATERIALIZED VIEW example (commented)
-- DROP MATERIALIZED VIEW mv_order_summary;

-- CREATE SYNONYM example
CREATE PUBLIC SYNONYM orders_syn FOR orders;

-- DROP SYNONYM example (commented)
-- DROP PUBLIC SYNONYM orders_syn;

-- CREATE DIRECTORY example
CREATE OR REPLACE DIRECTORY data_pump_dir AS 'path/to/directory';

-- DROP DIRECTORY example (commented)
-- DROP DIRECTORY data_pump_dir;
