-- DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    password VARCHAR(256) NOT NULL,
    config_yaml TEXT NOT NULL
);
--
-- INSERT INTO users(username, password, config_yaml)
-- VALUES ('admin', 'admin', '#
-- # Configuration file for API Gateway
-- #
--
-- connections:
--   routes:
--     - path: /products
--       url: http://services:9001
--       rate-limit:
--         zone: 10
--         rate: 5
--       auth: false
--
--     - path: /orders
--       url: http://services:9002
--       rate-limit:
--         zone: 10
--         rate: 5
--       auth: false
--
--     - path: /protected
--       url: http://services:9003
--       rate-limit:
--         zone: 10
--         rate: 5
--       auth: true
--
--     - path: /external-weather
--       url: https://api.open-meteo.com
--       rate-limit:
--         zone: 10
--         rate: 5
--       auth: false');

-- DROP TABLE IF EXISTS sessions;
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    expires TIMESTAMP NOT NULL
);