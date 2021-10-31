-- +migrate Up

CREATE TABLE IF NOT EXISTS users (
    Id SERIAL PRIMARY KEY,
    login varchar(255) not null,
    password varchar(255) not null,
    name varchar(255),
    phone varchar(255)
    );

-- +migrate Down
DROP TABLE IF EXISTS users;