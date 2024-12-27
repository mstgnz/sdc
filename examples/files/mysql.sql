-- Create Database
CREATE DATABASE IF NOT EXISTS example_db;
USE example_db;

-- Create Tables
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,
    status ENUM('active', 'inactive') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    status ENUM('draft', 'published', 'archived') DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_status ON posts(status);

-- Create View
CREATE OR REPLACE VIEW active_users_view AS
SELECT 
    u.*,
    COUNT(p.id) as post_count,
    MAX(p.created_at) as last_post_date
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.status = 'active'
GROUP BY u.id;

-- Create Trigger for Logging
DELIMITER //
CREATE TRIGGER users_after_delete
AFTER DELETE ON users
FOR EACH ROW
BEGIN
    INSERT INTO user_logs (user_id, action, action_date)
    VALUES (OLD.id, 'DELETE', NOW());
END //
DELIMITER ;

-- Insert Data
INSERT INTO users (id, username, email, password) VALUES
(1, 'john_doe', 'john@example.com', 'hashed_password_1'),
(2, 'jane_doe', 'jane@example.com', 'hashed_password_2'),
(3, 'alice_smith', 'alice@example.com', 'hashed_password_3'),
(4, 'bob_wilson', 'bob@example.com', 'hashed_password_4');

INSERT INTO posts (id, user_id, title, content, status) VALUES
(1, 1, 'First Post', 'This is my first post content', 'published'),
(2, 1, 'Second Post', 'This is a draft post', 'draft'),
(3, 2, 'Hello World', 'Post by Jane', 'published'),
(4, 2, 'Another Post', 'Another post by Jane', 'published'),
(5, 3, 'Tech News', 'Latest technology news', 'published'),
(6, 3, 'Draft Post', 'Work in progress', 'draft'),
(7, 4, 'Travel Blog', 'My travel experiences', 'published'),
(8, 4, 'Food Blog', 'Best recipes', 'published');
