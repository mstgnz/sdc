-- Sequence and defined type
CREATE TABLE IF NOT
    EXISTS blogs_id_seq(id INT AUTO_INCREMENT PRIMARY KEY);

-- Table Definition
CREATE TABLE blogs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255),
    user_id INT NOT NULL,
    short_text VARCHAR(255) NOT NULL,
    long_text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT blogs_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Column Comment
COMMENT ON COLUMN

    blogs.slug IS 'Title Slug';
