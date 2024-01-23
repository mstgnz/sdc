-- 'users' tablosunu oluşturalım
CREATE TABLE users (
    user_id INT PRIMARY KEY,
    username NVARCHAR(50) NOT NULL,
    email NVARCHAR(100) UNIQUE NOT NULL,
    registration_date DATETIME DEFAULT GETDATE(),
    -- EXEC sp_addextendedproperty 'MS_Description', 'Table to store user information', 'SCHEMA', 'dbo', 'TABLE', 'users';
);

-- 'orders' tablosunu oluşturalım
CREATE TABLE orders (
    order_id INT PRIMARY KEY,
    user_id INT FOREIGN KEY REFERENCES users(user_id),
    product_name NVARCHAR(255) NOT NULL,
    quantity INT,
    order_date DATETIME DEFAULT GETDATE(),
    -- EXEC sp_addextendedproperty 'MS_Description', 'Table to store order information', 'SCHEMA', 'dbo', 'TABLE', 'orders';
);

-- Add an index for the 'users' table
CREATE INDEX idx_email ON users(email);

-- Add an index for the 'orders' table
CREATE INDEX idx_order_date ON orders(order_date);