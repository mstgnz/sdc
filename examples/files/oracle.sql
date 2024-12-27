-- Create Tablespace
CREATE TABLESPACE example_data
DATAFILE 'example_data.dbf' SIZE 100M
AUTOEXTEND ON NEXT 50M MAXSIZE UNLIMITED;

-- Create User
CREATE USER example_user IDENTIFIED BY example_password
DEFAULT TABLESPACE example_data
QUOTA UNLIMITED ON example_data;

GRANT CREATE SESSION, CREATE TABLE, CREATE VIEW, CREATE SEQUENCE, CREATE TRIGGER TO example_user;

-- Connect as example_user
ALTER SESSION SET CURRENT_SCHEMA = example_user;

-- Create Sequences
CREATE SEQUENCE users_seq START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE posts_seq START WITH 1 INCREMENT BY 1;

-- Create Tables
CREATE TABLE users (
    id NUMBER DEFAULT users_seq.NEXTVAL PRIMARY KEY,
    username VARCHAR2(50) NOT NULL UNIQUE,
    email VARCHAR2(100) NOT NULL,
    password VARCHAR2(255) NOT NULL,
    status VARCHAR2(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_at TIMESTAMP DEFAULT SYSTIMESTAMP
);

CREATE TABLE posts (
    id NUMBER DEFAULT posts_seq.NEXTVAL PRIMARY KEY,
    user_id NUMBER NOT NULL,
    title VARCHAR2(255) NOT NULL,
    content CLOB,
    status VARCHAR2(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    created_at TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_at TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

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
GROUP BY u.id, u.username, u.email, u.password, u.status, u.created_at, u.updated_at;

-- Create Trigger for Updated Timestamp
CREATE OR REPLACE TRIGGER users_update_timestamp
BEFORE UPDATE ON users
FOR EACH ROW
BEGIN
    :NEW.updated_at := SYSTIMESTAMP;
END;
/

CREATE OR REPLACE TRIGGER posts_update_timestamp
BEFORE UPDATE ON posts
FOR EACH ROW
BEGIN
    :NEW.updated_at := SYSTIMESTAMP;
END;
/

-- Insert Data
INSERT INTO users (id, username, email, password) VALUES
(1, 'john_doe', 'john@example.com', 'hashed_password_1');
INSERT INTO users (id, username, email, password) VALUES
(2, 'jane_doe', 'jane@example.com', 'hashed_password_2');
INSERT INTO users (id, username, email, password) VALUES
(3, 'alice_smith', 'alice@example.com', 'hashed_password_3');
INSERT INTO users (id, username, email, password) VALUES
(4, 'bob_wilson', 'bob@example.com', 'hashed_password_4');

INSERT INTO posts (id, user_id, title, content, status) VALUES
(1, 1, 'First Post', 'This is my first post content', 'published');
INSERT INTO posts (id, user_id, title, content, status) VALUES
(2, 1, 'Second Post', 'This is a draft post', 'draft');
INSERT INTO posts (id, user_id, title, content, status) VALUES
(3, 2, 'Hello World', 'Post by Jane', 'published');
INSERT INTO posts (id, user_id, title, content, status) VALUES
(4, 2, 'Another Post', 'Another post by Jane', 'published');
INSERT INTO posts (id, user_id, title, content, status) VALUES
(5, 3, 'Tech News', 'Latest technology news', 'published');
INSERT INTO posts (id, user_id, title, content, status) VALUES
(6, 3, 'Draft Post', 'Work in progress', 'draft');
INSERT INTO posts (id, user_id, title, content, status) VALUES
(7, 4, 'Travel Blog', 'My travel experiences', 'published');
INSERT INTO posts (id, user_id, title, content, status) VALUES
(8, 4, 'Food Blog', 'Best recipes', 'published');

COMMIT;
