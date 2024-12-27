-- Enable Foreign Keys
PRAGMA foreign_keys = ON;

-- Create Tables
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT,
    status TEXT DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_status ON posts(status);

-- Create View
CREATE VIEW active_users_view AS
SELECT 
    u.*,
    COUNT(p.id) as post_count,
    MAX(p.created_at) as last_post_date
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.status = 'active'
GROUP BY u.id;

-- Create Trigger for Updated Timestamp
CREATE TRIGGER users_update_timestamp
AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;

CREATE TRIGGER posts_update_timestamp
AFTER UPDATE ON posts
BEGIN
    UPDATE posts SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;

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
