# PostgreSQL Kullanım Kılavuzu

## İçindekiler
- [Bağlantı Yapılandırması](#bağlantı-yapılandırması)
- [Temel Kullanım](#temel-kullanım)
- [Veri Tipleri](#veri-tipleri)
- [PostgreSQL Özellikleri](#postgresql-özellikleri)
- [Örnek Senaryolar](#örnek-senaryolar)

## Bağlantı Yapılandırması

PostgreSQL bağlantısı için aşağıdaki yapılandırmayı kullanabilirsiniz:

```go
config := db.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "mydb",
    Username: "user",
    Password: "pass",
    SSLMode:  "disable", // veya "require", "verify-full"
}
```

## Temel Kullanım

PostgreSQL dump dosyasını dönüştürmek için:

```go
// PostgreSQL parser oluştur
parser := sqlporter.NewPostgresParser()

// PostgreSQL dump'ı parse et
entity, err := parser.Parse(pgDump)
if err != nil {
    log.Error("Parse hatası", err)
    return
}

// MySQL'e dönüştür
mysqlParser := sqlporter.NewMySQLParser()
mysqlSQL, err := mysqlParser.Convert(entity)
if err != nil {
    log.Error("Dönüştürme hatası", err)
    return
}
```

## Veri Tipleri

PostgreSQL'den diğer veritabanlarına dönüşüm yaparken veri tipi eşleştirmeleri:

| PostgreSQL Veri Tipi | MySQL | SQLite | Oracle | SQL Server |
|---------------------|-------|---------|---------|------------|
| SMALLINT           | SMALLINT | INTEGER | NUMBER | SMALLINT |
| INTEGER            | INT    | INTEGER | NUMBER | INT      |
| BIGINT             | BIGINT | INTEGER | NUMBER | BIGINT   |
| VARCHAR            | VARCHAR| TEXT    | VARCHAR2| VARCHAR  |
| TEXT               | TEXT   | TEXT    | CLOB   | TEXT     |
| TIMESTAMP          | DATETIME| TEXT   | TIMESTAMP| DATETIME |
| NUMERIC            | DECIMAL| REAL    | NUMBER | DECIMAL  |
| SERIAL             | AUTO_INCREMENT | AUTOINCREMENT | SEQUENCE | IDENTITY |

## PostgreSQL Özellikleri

### Schemas

PostgreSQL schema kullanımı:
```sql
CREATE SCHEMA myschema;
CREATE TABLE myschema.mytable (
    id SERIAL PRIMARY KEY,
    data TEXT
);
```

### İnheritance (Kalıtım)

PostgreSQL'e özgü tablo kalıtımı:
```sql
CREATE TABLE cities (
    name text,
    population real,
    elevation int
);

CREATE TABLE capitals (
    state char(2)
) INHERITS (cities);
```

### Özel Veri Tipleri

PostgreSQL'in desteklediği özel veri tipleri:
- JSONB
- Array types
- Geometric types
- Network address types
- UUID

## Örnek Senaryolar

### 1. JSONB Kullanımı

```go
pgDump := `CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    attributes JSONB
);`

// PostgreSQL'den MySQL'e dönüştürme
mysqlSQL, err := Convert(pgDump, "postgres", "mysql")
```

### 2. Array Tipleri

```go
pgDump := `CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    tags TEXT[],
    numbers INTEGER[]
);`

// PostgreSQL'den Oracle'a dönüştürme
oracleSQL, err := Convert(pgDump, "postgres", "oracle")
```

### 3. Composite Types

```go
pgDump := `
CREATE TYPE address AS (
    street VARCHAR(100),
    city VARCHAR(50),
    country VARCHAR(50)
);

CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    shipping_address address,
    billing_address address
);`

// PostgreSQL'den SQL Server'a dönüştürme
mssqlSQL, err := Convert(pgDump, "postgres", "sqlserver")
``` 