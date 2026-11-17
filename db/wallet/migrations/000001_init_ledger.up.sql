BEGIN;




CREATE TYPE entry_direction AS ENUM ('debit','credit')


CREATE TABLE accounts (
     id uuid           PRIMARY KEY DEFAULT get_random_uuid()
)
