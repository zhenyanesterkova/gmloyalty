BEGIN TRANSACTION;

CREATE TABLE users(
    id SERIAL UNIQUE NOT NULL PRIMARY KEY,
    user_login VARCHAR(200) UNIQUE NOT NULL,
    hashed_password  VARCHAR(200) NOT NULL
);

CREATE TABLE accounts(
    id SERIAL UNIQUE NOT NULL PRIMARY KEY,
    user_id INT NOT NULL,
    balance DOUBLE PRECISION NOT NULL,
    withdrawn DOUBLE PRECISION
);

CREATE TABLE orders(
    order_num VARCHAR(200) UNIQUE NOT NULL PRIMARY KEY,
    order_status VARCHAR(200) NOT NULL,
    upload_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id INT NOT NULL
);

CREATE TABLE history(
    id SERIAL UNIQUE NOT NULL PRIMARY KEY,
    sum DOUBLE PRECISION NOT NULL,
    item_type VARCHAR(200) NOT NULL,
    item_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    order_num VARCHAR(200) UNIQUE NOT NULL
);

CREATE INDEX user_login ON users (user_login);
CREATE INDEX order_num ON orders (order_num);
CREATE INDEX order_user_id ON orders (user_id);
CREATE INDEX accounts_user_id ON accounts (user_id);
CREATE INDEX history_order_num ON history (order_num);

COMMIT;