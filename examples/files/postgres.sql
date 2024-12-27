-- Create Database
CREATE DATABASE example_db;

\c example_db;

-- Create Schema
CREATE SCHEMA app;

-- Create Extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create Tables
CREATE TABLE app.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app.posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES app.users(id) ON DELETE CASCADE
);

-- Create Indexes
CREATE INDEX idx_users_email ON app.users(email);
CREATE INDEX idx_posts_user_id ON app.posts(user_id);
CREATE INDEX idx_posts_status ON app.posts(status);

-- Create View
CREATE OR REPLACE VIEW app.active_users_view AS
SELECT 
    u.*,
    COUNT(p.id) as post_count,
    MAX(p.created_at) as last_post_date
FROM app.users u
LEFT JOIN app.posts p ON u.id = p.user_id
WHERE u.status = 'active'
GROUP BY u.id;

-- Create Function for Updated Timestamp
CREATE OR REPLACE FUNCTION app.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create Triggers
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON app.users
    FOR EACH ROW
    EXECUTE FUNCTION app.update_updated_at_column();

CREATE TRIGGER update_posts_updated_at
    BEFORE UPDATE ON app.posts
    FOR EACH ROW
    EXECUTE FUNCTION app.update_updated_at_column();

-- Insert Data
INSERT INTO app.users (id, username, email, password) VALUES
(1, 'john_doe', 'john@example.com', 'hashed_password_1'),
(2, 'jane_doe', 'jane@example.com', 'hashed_password_2'),
(3, 'alice_smith', 'alice@example.com', 'hashed_password_3'),
(4, 'bob_wilson', 'bob@example.com', 'hashed_password_4');

INSERT INTO app.posts (id, user_id, title, content, status) VALUES
(1, 1, 'First Post', 'This is my first post content', 'published'),
(2, 1, 'Second Post', 'This is a draft post', 'draft'),
(3, 2, 'Hello World', 'Post by Jane', 'published'),
(4, 2, 'Another Post', 'Another post by Jane', 'published'),
(5, 3, 'Tech News', 'Latest technology news', 'published'),
(6, 3, 'Draft Post', 'Work in progress', 'draft'),
(7, 4, 'Travel Blog', 'My travel experiences', 'published'),
(8, 4, 'Food Blog', 'Best recipes', 'published');