-- Copyright (c) 2026 Dedy Setiawan. All rights reserved.
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tambahan dari TDD 6.1 (CategoryID) — belum ada di ERD 3.5.
CREATE TABLE IF NOT EXISTS categories (
    id         UUID PRIMARY KEY,
    name       VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255)  NOT NULL,
    sku         VARCHAR(64)   NOT NULL,              -- TDD 6.1
    description TEXT          NOT NULL DEFAULT '',
    price       DECIMAL(12,2) NOT NULL CHECK (price >= 0),
    stock       INT           NOT NULL DEFAULT 0 CHECK (stock >= 0),
    category_id UUID REFERENCES categories(id),      -- TDD 6.1
    image_url   VARCHAR(512)  NOT NULL DEFAULT '',
    deleted_at  TIMESTAMPTZ,                         -- soft delete (TDD 6.1)
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);
-- SKU unik hanya di antara produk yang belum dihapus.
CREATE UNIQUE INDEX IF NOT EXISTS ux_products_sku_active ON products (sku) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_products_category ON products (category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_products_created  ON products (created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS transactions (
    id               UUID PRIMARY KEY,
    invoice_number   VARCHAR(64)   NOT NULL UNIQUE,
    customer_name    VARCHAR(255)  NOT NULL,
    customer_email   VARCHAR(255)  NOT NULL,
    customer_phone   VARCHAR(32)   NOT NULL DEFAULT '',   -- form checkout (TDD 2.2.B)
    shipping_address TEXT          NOT NULL DEFAULT '',   -- form checkout (TDD 2.2.B)
    total_amount     DECIMAL(12,2) NOT NULL CHECK (total_amount >= 0),
    status           VARCHAR(32)   NOT NULL DEFAULT 'PENDING'
                     CHECK (status IN ('PENDING','PAID','EXPIRED','CANCELLED')),
    payment_url      VARCHAR(512)  NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS ix_transactions_status_created ON transactions (status, created_at DESC);

CREATE TABLE IF NOT EXISTS transaction_items (
    id                UUID PRIMARY KEY,
    transaction_id    UUID          NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    product_id        UUID          NOT NULL REFERENCES products(id),
    quantity          INT           NOT NULL CHECK (quantity > 0),
    price_at_purchase DECIMAL(12,2) NOT NULL CHECK (price_at_purchase >= 0)
);
CREATE INDEX IF NOT EXISTS ix_transaction_items_trx ON transaction_items (transaction_id);
