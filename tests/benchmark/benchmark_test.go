package benchmark

import (
	"testing"

	"github.com/mstgnz/sqlmapper/mysql"
	"github.com/mstgnz/sqlmapper/postgres"
	"github.com/mstgnz/sqlmapper/sqlite"
)

var complexMySQLSchema = `
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id INT,
    FOREIGN KEY (parent_id) REFERENCES categories(id)
);

CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category_id INT NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock INT DEFAULT 0,
    status ENUM('active', 'inactive', 'deleted') DEFAULT 'active',
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE INDEX idx_user_email ON users(email);
CREATE INDEX idx_category_parent ON categories(parent_id);
CREATE INDEX idx_product_category ON products(category_id);
CREATE INDEX idx_product_status ON products(status);
CREATE FULLTEXT INDEX idx_product_search ON products(name, description);

CREATE VIEW active_products AS
SELECT p.*, c.name as category_name
FROM products p
JOIN categories c ON p.category_id = c.id
WHERE p.status = 'active';

DELIMITER //
CREATE TRIGGER update_product_timestamp
BEFORE UPDATE ON products
FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END//
DELIMITER ;
`

func BenchmarkMySQLToPostgreSQL(b *testing.B) {
	mysqlParser := mysql.NewMySQL()
	pgParser := postgres.NewPostgreSQL()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema, err := mysqlParser.Parse(complexMySQLSchema)
		if err != nil {
			b.Fatal(err)
		}

		_, err = pgParser.Generate(schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMySQLToSQLite(b *testing.B) {
	mysqlParser := mysql.NewMySQL()
	sqliteParser := sqlite.NewSQLite()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema, err := mysqlParser.Parse(complexMySQLSchema)
		if err != nil {
			b.Fatal(err)
		}

		_, err = sqliteParser.Generate(schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPostgreSQLToMySQL(b *testing.B) {
	pgParser := postgres.NewPostgreSQL()
	mysqlParser := mysql.NewMySQL()

	// Convert MySQL schema to PostgreSQL first
	schema, err := mysql.NewMySQL().Parse(complexMySQLSchema)
	if err != nil {
		b.Fatal(err)
	}

	pgSQL, err := pgParser.Generate(schema)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema, err := pgParser.Parse(pgSQL)
		if err != nil {
			b.Fatal(err)
		}

		_, err = mysqlParser.Generate(schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSQLiteToMySQL(b *testing.B) {
	sqliteParser := sqlite.NewSQLite()
	mysqlParser := mysql.NewMySQL()

	// Use a simpler schema for SQLite
	simpleSQLiteSchema := `
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    price REAL NOT NULL,
    stock INTEGER DEFAULT 0
);

CREATE VIEW active_products AS
SELECT * FROM products WHERE stock > 0;
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema, err := sqliteParser.Parse(simpleSQLiteSchema)
		if err != nil {
			b.Fatal(err)
		}

		_, err = mysqlParser.Generate(schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}
