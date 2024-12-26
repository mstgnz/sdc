# MySQL Kullanım Kılavuzu

## İçindekiler
- [Bağlantı Yapılandırması](#bağlantı-yapılandırması)
- [Temel Kullanım](#temel-kullanım)
- [Veri Tipleri](#veri-tipleri)
- [Özel Durumlar](#özel-durumlar)
- [Örnek Senaryolar](#örnek-senaryolar)

## Bağlantı Yapılandırması

MySQL bağlantısı için aşağıdaki yapılandırmayı kullanabilirsiniz:

```go
config := db.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "mydb",
    Username: "user",
    Password: "pass",
}
```

## Temel Kullanım

MySQL dump dosyasını dönüştürmek için:

```go
// MySQL parser oluştur
parser := sdc.NewMySQLParser()

// MySQL dump'ı parse et
entity, err := parser.Parse(mysqlDump)
if err != nil {
    log.Error("Parse hatası", err)
    return
}

// PostgreSQL'e dönüştür
pgParser := sdc.NewPostgresParser()
pgSQL, err := pgParser.Convert(entity)
if err != nil {
    log.Error("Dönüştürme hatası", err)
    return
}
```

## Veri Tipleri

MySQL'den diğer veritabanlarına dönüşüm yaparken veri tipi eşleştirmeleri:

| MySQL Veri Tipi | PostgreSQL | SQLite | Oracle | SQL Server |
|-----------------|------------|---------|---------|------------|
| TINYINT        | SMALLINT   | INTEGER | NUMBER  | TINYINT    |
| INT            | INTEGER    | INTEGER | NUMBER  | INT        |
| BIGINT         | BIGINT     | INTEGER | NUMBER  | BIGINT     |
| VARCHAR        | VARCHAR    | TEXT    | VARCHAR2| VARCHAR    |
| TEXT           | TEXT       | TEXT    | CLOB    | TEXT       |
| DATETIME       | TIMESTAMP  | TEXT    | DATE    | DATETIME   |
| DECIMAL        | DECIMAL    | REAL    | NUMBER  | DECIMAL    |

## Özel Durumlar

### AUTO_INCREMENT

MySQL'in AUTO_INCREMENT özelliği diğer veritabanlarında şu şekilde karşılık bulur:
- PostgreSQL: SERIAL
- SQLite: AUTOINCREMENT
- Oracle: SEQUENCE
- SQL Server: IDENTITY

### Karakter Seti ve Collation

MySQL'de karakter seti ve collation belirtimi:
```sql
CREATE TABLE users (
    name VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
);
```

## Örnek Senaryolar

### 1. Tablo Oluşturma ve Dönüştürme

```go
mysqlDump := `CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

// MySQL'den PostgreSQL'e dönüştürme
pgSQL, err := Convert(mysqlDump, "mysql", "postgres")
```

### 2. İndeks ve Kısıtlamalar

```go
mysqlDump := `CREATE TABLE orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    total DECIMAL(10,2) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_user_id (user_id)
);`

// MySQL'den Oracle'a dönüştürme
oracleSQL, err := Convert(mysqlDump, "mysql", "oracle")
``` 