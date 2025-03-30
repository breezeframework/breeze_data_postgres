#!/bin/bash
set -e
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS MyObjTable (
        id SERIAL PRIMARY KEY,
        field1 INT,
        field2 TEXT
    );

    INSERT INTO MyObjTable  (id,field1,field2) values(1,5,'str');
EOSQL
