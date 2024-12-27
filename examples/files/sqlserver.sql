-- Create Database
CREATE DATABASE example_db;
GO

USE example_db;
GO

-- Create Schema
CREATE SCHEMA app;
GO

-- Create Tables
CREATE TABLE app.users (
    id INT IDENTITY(1,1) PRIMARY KEY,
    username NVARCHAR(50) NOT NULL UNIQUE,
    email NVARCHAR(100) NOT NULL,
    password NVARCHAR(255) NOT NULL,
    status NVARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE()
);

CREATE TABLE app.posts (
    id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT NOT NULL,
    title NVARCHAR(255) NOT NULL,
    content NVARCHAR(MAX),
    status NVARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    CONSTRAINT FK_posts_users FOREIGN KEY (user_id) REFERENCES app.users(id) ON DELETE CASCADE
);

-- Create Indexes
CREATE INDEX idx_users_email ON app.users(email);
CREATE INDEX idx_posts_user_id ON app.posts(user_id);
CREATE INDEX idx_posts_status ON app.posts(status);

-- Create View
CREATE VIEW app.active_users_view
AS
SELECT 
    u.*,
    COUNT(p.id) as post_count,
    MAX(p.created_at) as last_post_date
FROM app.users u
LEFT JOIN app.posts p ON u.id = p.user_id
WHERE u.status = 'active'
GROUP BY u.id, u.username, u.email, u.password, u.status, u.created_at, u.updated_at;
GO

-- Create Trigger for Updated Timestamp
CREATE TRIGGER app.users_update_timestamp
ON app.users
AFTER UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    UPDATE app.users
    SET updated_at = GETDATE()
    FROM app.users u
    INNER JOIN inserted i ON u.id = i.id;
END;
GO

CREATE TRIGGER app.posts_update_timestamp
ON app.posts
AFTER UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    UPDATE app.posts
    SET updated_at = GETDATE()
    FROM app.posts p
    INNER JOIN inserted i ON p.id = i.id;
END;
GO

-- Insert Data
SET IDENTITY_INSERT app.users ON;
INSERT INTO app.users (id, username, email, password) VALUES
(1, 'john_doe', 'john@example.com', 'hashed_password_1'),
(2, 'jane_doe', 'jane@example.com', 'hashed_password_2'),
(3, 'alice_smith', 'alice@example.com', 'hashed_password_3'),
(4, 'bob_wilson', 'bob@example.com', 'hashed_password_4');
SET IDENTITY_INSERT app.users OFF;

SET IDENTITY_INSERT app.posts ON;
INSERT INTO app.posts (id, user_id, title, content, status) VALUES
(1, 1, 'First Post', 'This is my first post content', 'published'),
(2, 1, 'Second Post', 'This is a draft post', 'draft'),
(3, 2, 'Hello World', 'Post by Jane', 'published'),
(4, 2, 'Another Post', 'Another post by Jane', 'published'),
(5, 3, 'Tech News', 'Latest technology news', 'published'),
(6, 3, 'Draft Post', 'Work in progress', 'draft'),
(7, 4, 'Travel Blog', 'My travel experiences', 'published'),
(8, 4, 'Food Blog', 'Best recipes', 'published');
SET IDENTITY_INSERT app.posts OFF;
