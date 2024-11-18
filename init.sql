-- kata_test/init.sql
\c kata_test;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS cities (
    id bigserial PRIMARY KEY,
    name citext NOT NULL,
    state citext NOT NULL
);


