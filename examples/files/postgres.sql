-- Sequence ve tanımlı tipler
CREATE SEQUENCE IF NOT EXISTS users_id_seq;
CREATE SEQUENCE IF NOT EXISTS products_id_seq;
CREATE SEQUENCE IF NOT EXISTS orders_id_seq;

-- Enum tipi tanımlama
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE order_status AS ENUM ('pending', 'processing', 'completed', 'cancelled');

-- Users tablosu
CREATE TABLE "public"."users" (
    "id" bigint NOT NULL DEFAULT nextval('users_id_seq'::regclass),
    "username" varchar(50) NOT NULL,
    "email" varchar(100) NOT NULL UNIQUE,
    "status" user_status DEFAULT 'active',
    "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
);

-- Products tablosu
CREATE TABLE "public"."products" (
    "id" bigint NOT NULL DEFAULT nextval('products_id_seq'::regclass),
    "name" varchar(100) NOT NULL,
    "description" text,
    "price" decimal(10,2) NOT NULL CHECK (price >= 0),
    "stock" integer NOT NULL DEFAULT 0 CHECK (stock >= 0),
    "metadata" jsonb,
    "is_active" boolean DEFAULT true,
    "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "products_pkey" PRIMARY KEY ("id")
);

-- Orders tablosu
CREATE TABLE "public"."orders" (
    "id" bigint NOT NULL DEFAULT nextval('orders_id_seq'::regclass),
    "user_id" bigint NOT NULL,
    "status" order_status DEFAULT 'pending',
    "total_amount" decimal(12,2) NOT NULL CHECK (total_amount >= 0),
    "shipping_address" text NOT NULL,
    "order_date" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    "delivery_date" timestamp with time zone,
    "metadata" jsonb,
    CONSTRAINT "orders_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "orders_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE
);

-- Order Items tablosu (çoka-çok ilişki)
CREATE TABLE "public"."order_items" (
    "order_id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "quantity" integer NOT NULL CHECK (quantity > 0),
    "unit_price" decimal(10,2) NOT NULL CHECK (unit_price >= 0),
    "total_price" decimal(10,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    CONSTRAINT "order_items_pkey" PRIMARY KEY ("order_id", "product_id"),
    CONSTRAINT "order_items_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "public"."orders"("id") ON DELETE CASCADE,
    CONSTRAINT "order_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products"("id") ON DELETE RESTRICT
);

-- İndeksler
CREATE INDEX idx_users_email ON "public"."users" USING btree ("email");
CREATE INDEX idx_products_name ON "public"."products" USING btree ("name");
CREATE INDEX idx_orders_user_id ON "public"."orders" USING btree ("user_id");
CREATE INDEX idx_orders_status ON "public"."orders" USING btree ("status");
CREATE INDEX idx_order_items_product ON "public"."order_items" USING btree ("product_id");

-- Partition örneği
CREATE TABLE "public"."order_history" (
    "id" bigint NOT NULL,
    "order_id" bigint NOT NULL,
    "status" order_status NOT NULL,
    "changed_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (changed_at);

-- Partition tabloları
CREATE TABLE order_history_2023 PARTITION OF order_history
    FOR VALUES FROM ('2023-01-01') TO ('2024-01-01');
CREATE TABLE order_history_2024 PARTITION OF order_history
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

-- View örneği
CREATE VIEW "public"."order_summary" AS
SELECT 
    o.id as order_id,
    u.username,
    o.total_amount,
    o.status,
    o.order_date,
    COUNT(oi.product_id) as total_items
FROM "public"."orders" o
JOIN "public"."users" u ON o.user_id = u.id
JOIN "public"."order_items" oi ON o.id = oi.order_id
GROUP BY o.id, u.username, o.total_amount, o.status, o.order_date;

-- Table and column comments
COMMENT ON TABLE "public"."users" IS 'Table storing user information';
COMMENT ON COLUMN "public"."users"."email" IS 'User email address';
COMMENT ON TABLE "public"."products" IS 'Table storing product information';
COMMENT ON TABLE "public"."orders" IS 'Table storing order information';
COMMENT ON TABLE "public"."order_items" IS 'Table storing order item details';

-- DDL Command Examples
-- CREATE DATABASE example
CREATE DATABASE ecommerce
    WITH 
    OWNER = postgres
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TABLESPACE = pg_default
    CONNECTION LIMIT = -1;

-- DROP DATABASE example (commented)
-- DROP DATABASE IF EXISTS ecommerce;

-- ALTER DATABASE example
ALTER DATABASE ecommerce
    SET search_path = public, extensions;

-- CREATE SCHEMA example
CREATE SCHEMA IF NOT EXISTS app
    AUTHORIZATION postgres;

-- DROP SCHEMA example (commented)
-- DROP SCHEMA IF EXISTS app CASCADE;

-- ALTER SCHEMA example
ALTER SCHEMA app OWNER TO postgres;

-- CREATE EXTENSION example
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- DROP EXTENSION example (commented)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
-- DROP EXTENSION IF EXISTS "pgcrypto";

-- CREATE USER example
CREATE USER app_user WITH 
    LOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    INHERIT
    NOREPLICATION
    CONNECTION LIMIT -1
    PASSWORD 'password';

-- ALTER USER example
ALTER USER app_user WITH PASSWORD 'new_password';

-- DROP USER example (commented)
-- DROP USER IF EXISTS app_user;

-- GRANT examples
GRANT CONNECT ON DATABASE ecommerce TO app_user;
GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_user;
GRANT INSERT, UPDATE ON products TO app_user;

-- REVOKE examples
REVOKE INSERT, UPDATE ON products FROM app_user;

-- CREATE ROLE example
CREATE ROLE app_read_role;
GRANT CONNECT ON DATABASE ecommerce TO app_read_role;
GRANT USAGE ON SCHEMA public TO app_read_role;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_read_role;
GRANT app_read_role TO app_user;

-- DROP ROLE example (commented)
-- DROP ROLE IF EXISTS app_read_role;

-- ALTER TABLE examples
ALTER TABLE users 
    ADD COLUMN phone VARCHAR(20),
    ADD COLUMN address TEXT;

ALTER TABLE users
    ADD CONSTRAINT users_phone_unique UNIQUE (phone);

ALTER TABLE products 
    ALTER COLUMN price TYPE DECIMAL(12,2),
    ALTER COLUMN price SET NOT NULL,
    ALTER COLUMN stock SET DEFAULT 100;

ALTER TABLE products
    ADD CONSTRAINT products_name_unique UNIQUE (name);

-- DROP TABLE examples (commented)
-- DROP TABLE IF EXISTS order_items CASCADE;
-- DROP TABLE IF EXISTS orders CASCADE;
-- DROP TABLE IF EXISTS products CASCADE;
-- DROP TABLE IF EXISTS users CASCADE;

-- TRUNCATE example (commented)
-- TRUNCATE TABLE order_items CASCADE;
-- TRUNCATE TABLE orders CASCADE;
-- TRUNCATE TABLE products CASCADE;
-- TRUNCATE TABLE users CASCADE;

-- CREATE INDEX examples
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_orders_created ON orders(created_at);

-- DROP INDEX examples (commented)
-- DROP INDEX IF EXISTS idx_users_username;
-- DROP INDEX IF EXISTS idx_products_price;
-- DROP INDEX IF EXISTS idx_orders_created;

-- ALTER INDEX example
ALTER INDEX idx_users_email RENAME TO idx_users_email_new;

-- CREATE VIEW example
CREATE OR REPLACE VIEW view_order_summary AS
    SELECT u.username, COUNT(o.id) as total_orders, SUM(o.total_amount) as total_spent
    FROM users u
    LEFT JOIN orders o ON u.id = o.user_id
    GROUP BY u.username;

-- DROP VIEW example (commented)
-- DROP VIEW IF EXISTS view_order_summary;

-- CREATE TRIGGER example
CREATE OR REPLACE FUNCTION update_product_price()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.price < 0 THEN
        NEW.price := 0;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER before_product_update
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_product_price();

-- DROP TRIGGER example (commented)
-- DROP TRIGGER IF EXISTS before_product_update ON products;
-- DROP FUNCTION IF EXISTS update_product_price();

-- CREATE FUNCTION example
CREATE OR REPLACE FUNCTION get_user_orders(user_id INTEGER)
    RETURNS TABLE (
        order_id INTEGER,
        order_date TIMESTAMP,
        total_amount DECIMAL
    ) AS $$
BEGIN
    RETURN QUERY
    SELECT id, created_at, total_amount
    FROM orders
    WHERE orders.user_id = get_user_orders.user_id;
END;
$$ LANGUAGE plpgsql;

-- DROP FUNCTION example (commented)
-- DROP FUNCTION IF EXISTS get_user_orders(INTEGER);

-- CREATE TYPE example
CREATE TYPE order_status AS ENUM ('pending', 'processing', 'completed', 'cancelled');

-- DROP TYPE example (commented)
-- DROP TYPE IF EXISTS order_status;

-- CREATE DOMAIN example
CREATE DOMAIN positive_price AS DECIMAL(12,2)
    CHECK (VALUE >= 0);

-- DROP DOMAIN example (commented)
-- DROP DOMAIN IF EXISTS positive_price;

-- Sequence examples
CREATE SEQUENCE IF NOT EXISTS custom_seq
    INCREMENT 1
    START 1000
    MINVALUE 1000
    MAXVALUE 9999999999
    CACHE 1;

ALTER SEQUENCE custom_seq
    INCREMENT 5
    RESTART WITH 2000;

-- DROP SEQUENCE example (commented)
-- DROP SEQUENCE IF EXISTS custom_seq;