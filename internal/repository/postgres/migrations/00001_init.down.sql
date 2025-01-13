BEGIN TRANSACTION;

DROP INDEX IF EXISTS user_login;
DROP INDEX IF EXISTS order_num;
DROP INDEX IF EXISTS order_user_id;
DROP INDEX IF EXISTS accounts_user_id;
DROP INDEX IF EXISTS history_order_num;

DROP TABLE history;
DROP TABLE orders;
DROP TABLE accounts;
DROP TABLE users;

COMMIT; 