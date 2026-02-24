DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    password VARCHAR(256) NOT NULL,
    config_yaml TEXT NOT NULL
);

INSERT INTO users(username, password, config_yaml)
VALUES ('admin', 'admin', 'test');

DROP TABLE IF EXISTS sessions;
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    expires TIMESTAMP NOT NULL
);