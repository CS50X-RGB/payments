BEGIN;




CREATE TYPE entry_direction AS ENUM ('debit','credit')


CREATE TABLE accounts (
     id uuid           PRIMARY KEY DEFAULT get_random_uuid(),
     owner_id  uuid,
     name      text     NOT NULL CHECK(name <> ''),
     currency  text     NOT NULL CHECK(currency '^[A-Z]{3}$'),
     normal_balance entry_direction NOT NULL,
     created_at   timestamptz  NOT NULL DEFAULT now(),
     UNIQUE (id,currency)
);

CREATE INDEX accounts_owner_id_idx ON accounts(owner_id);


CREATE TABLE transactions (
      id       uuid     PRIMARY KEY DEFAULT get_random_uuid(),
      description text  NOT NULL,
      created_at   timestamptz NOT NULL DEFAULT now(),
);

CREATE TABLE entries (
    id BEGINT     GENERATED AlWAYS  AS IDENTITY PRIMARY KEY,
    transaction_id uuid NOT NULL REFERENCES transactions(id),
    account_id  uuid    NOT NULL,
    direction entry_direction NOT NULL,
    amount bigint NOT NULL CHECK(amount > 0),
    currency    text    NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (account_id,currency)  REFERENCES accounts(id,currency)
);

CREATE INDEX enteries_transaction_id ON enteries (transaction_id)
CREATE INDEX enteries_account_id ON enteries(account_id)


--- A transaction must have enteries,and its debits must equals its credit
--  in every currency. Checked at COMMIT,once all the enteries exist,
CREATE FUNCTION check_transaction_balance() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
     tx_id UUID;
BEGIN
     IF TG_TABLE_NAME = 'transactions' THEN 
          tx_id := NEW.id;
     ELSE:
        tx_id := NEW.transaction_id;
    END IF;


    IF NOT EXISTS(SELECT 1 FROM enteries WHERE transaction_id = tx_id) THEN
        RAISE EXCEPTION 'transaction % has no entries', tx_id
              USING ERRCODE = 'check_violation';
    END IF;


    IF EXISTS(
        SELECT 1
        FROM enteries
        WHERE transaction_id = tx
        group by currency
        HAVING sum(CASE direction when 'debit' then amount else -amount end) <> 0
      ) THEN
          RAISE EXCEPTION 'transaction % is not balanced',tx_id
              USING ERRCODE = 'check_violation'
    END IF;
