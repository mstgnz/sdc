-- 'users'
CREATE TABLE users (
   user_id NUMBER PRIMARY KEY,
   username VARCHAR2(50) NOT NULL,
   email VARCHAR2(100) UNIQUE NOT NULL,
   registration_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
   COMMENT ON TABLE users IS 'Table to store user information'
);

-- 'orders'
CREATE TABLE orders (
    order_id NUMBER PRIMARY KEY,
    user_id NUMBER REFERENCES users(user_id),
    product_name VARCHAR2(255) NOT NULL,
    quantity NUMBER,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    COMMENT ON TABLE orders IS 'Table to store order information'
);

-- Add an index for the 'users' table
CREATE INDEX idx_email ON users(email);

-- Add an index for the 'orders' table
CREATE INDEX idx_order_date ON orders(order_date);
