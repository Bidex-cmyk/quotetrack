-- 0001_init.sql
-- QuoteTrack initial schema.

CREATE EXTENSION IF NOT EXISTS pgcrypto; -- provides gen_random_uuid()

-- Accounts / business profiles.
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    business_name TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Customers belong to a single user.
CREATE TABLE customers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL CHECK (length(trim(name)) > 0),
    phone      TEXT NOT NULL DEFAULT '',
    email      TEXT NOT NULL DEFAULT '',
    company    TEXT NOT NULL DEFAULT '',
    notes      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_customers_user_id ON customers (user_id);
CREATE INDEX idx_customers_user_search ON customers (user_id, lower(name));
CREATE INDEX idx_customers_user_email ON customers (user_id, lower(email)) WHERE email <> '';

-- Quotes.
CREATE TABLE quotes (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    customer_id    UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    public_id      TEXT NOT NULL UNIQUE,
    quote_number   TEXT NOT NULL,
    title          TEXT NOT NULL,
    notes          TEXT NOT NULL DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'draft'
                   CHECK (status IN ('draft', 'sent', 'waiting', 'won', 'lost', 'expired')),
    quote_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    expiry_date    DATE,
    follow_up_date DATE,
    subtotal       BIGINT NOT NULL DEFAULT 0 CHECK (subtotal >= 0), -- cents
    total          BIGINT NOT NULL DEFAULT 0 CHECK (total >= 0),    -- cents
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_quotes_user_id ON quotes (user_id);
CREATE INDEX idx_quotes_customer_id ON quotes (customer_id);
CREATE INDEX idx_quotes_follow_up ON quotes (user_id, follow_up_date)
    WHERE follow_up_date IS NOT NULL AND status NOT IN ('won', 'lost');

-- Quote line items.
CREATE TABLE quote_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_id    UUID NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    position    INT  NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    quantity    NUMERIC(12,2) NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price  BIGINT NOT NULL DEFAULT 0 CHECK (unit_price >= 0), -- cents
    total       BIGINT NOT NULL DEFAULT 0 CHECK (total >= 0),      -- cents
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_quote_items_quote_id ON quote_items (quote_id);