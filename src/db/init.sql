DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    password VARCHAR(256) NOT NULL,
    config_yaml TEXT NOT NULL
);

INSERT INTO users(username, password, config_yaml)
VALUES ('admin', 'admin', '#
# Configuration file for API Gateway
#

connections:
  - host: localhost
    port: 8080
    routes:
      - path: /products
        upstream:
          name: product_service
          port: 9001
        rate-limit:
          zone: 10
          rate: 5
        auth: false

      - path: /orders
        upstream:
          name: order_service
          port: 9002
        rate-limit:
          zone: 10
          rate: 5
        auth: false

      - path: /protected
        upstream:
          name: protected_service
          port: 9003
        rate-limit:
          zone: 10
          rate: 5
        auth: true');

DROP TABLE IF EXISTS sessions;
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    expires TIMESTAMP NOT NULL
);