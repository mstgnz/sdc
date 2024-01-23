-- 'users'
CREATE TABLE users (
   user_id INTEGER PRIMARY KEY,
   username TEXT NOT NULL,
   email TEXT UNIQUE NOT NULL,
   registration_date DATETIME DEFAULT CURRENT_TIMESTAMP,
   COMMENT ON TABLE users IS 'Table to store user information',
   COMMENT ON COLUMN users.email IS 'Description of the email'
);


-- 'orders'
CREATE TABLE orders (
    order_id INTEGER PRIMARY KEY,
    user_id INTEGER REFERENCES users(user_id),
    product_name TEXT NOT NULL,
    quantity INTEGER,
    order_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    COMMENT ON TABLE orders IS 'Table to store order information',
    COMMENT ON COLUMN orders.product_name IS 'Description of the product'
);

-- Add an index for the 'users' table
CREATE INDEX idx_email ON users(email);

-- Add an index for the 'orders' table
CREATE INDEX idx_order_date ON orders(order_date);
