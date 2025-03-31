#!/bin/bash
set -e
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS TestObjTable (
        id SERIAL PRIMARY KEY,
        field1 INT,
        field2 TEXT
    );


EOSQL
