CREATE TABLE users (
    id INTEGER AUTO_INCREMENT PRIMARY KEY DEFAULT 'nextval('id_seq'::regclass)',
    name VARCHAR(100)
);