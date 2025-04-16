#!/bin/bash
set -euo pipefail

# Validate required environment variables
: "${POSTGRES_USER:?Postgres username must be set}"
: "${POSTGRES_DB:?Postgres database name must be set}"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS test_plain_entity_table (
        id SERIAL PRIMARY KEY,
        field1 INTEGER,
        field2 TEXT
    );

    CREATE TABLE IF NOT EXISTS test_parent_entity_table (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS test_child1_table (
        id SERIAL PRIMARY KEY,
        type TEXT NOT NULL,
        parent_id INTEGER NOT NULL,
        FOREIGN KEY (parent_id) REFERENCES test_parent_entity_table(id)
            ON DELETE CASCADE
            ON UPDATE RESTRICT
    );

    CREATE TABLE IF NOT EXISTS test_child2_table (
        id SERIAL PRIMARY KEY,
        size FLOAT NOT NULL,
        parent_id INTEGER NOT NULL,
        FOREIGN KEY (parent_id) REFERENCES test_parent_entity_table(id)
            ON DELETE CASCADE
            ON UPDATE RESTRICT
    );
EOSQL
