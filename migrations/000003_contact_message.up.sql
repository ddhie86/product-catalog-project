CREATE TABLE contact_message (
    id         UUID PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    type       VARCHAR(32) NOT NULL CHECK (type IN ('QUESTION', 'COMPLAINT')),
    message    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ix_contact_message_created_at ON contact_message (created_at DESC);
