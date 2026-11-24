ALTER TABLE contact_message
    ADD COLUMN reply_message TEXT,
    ADD COLUMN replied_at TIMESTAMPTZ,
    ADD COLUMN replied_by UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE TABLE activity_log (
    id                  UUID PRIMARY KEY,
    datetime            TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id             UUID REFERENCES users(id) ON DELETE SET NULL,
    action              VARCHAR(32) NOT NULL,
    entity              VARCHAR(64) NOT NULL,
    data_before_updated JSONB,
    data_after_updated  JSONB
);
CREATE INDEX ix_activity_log_datetime ON activity_log (datetime DESC);
CREATE INDEX ix_activity_log_entity_datetime ON activity_log (entity, datetime DESC);

INSERT INTO permissions (key, menu, menu_label, action, label) VALUES
    ('contact_messages:view', 'contact_messages', 'Pesan kontak', 'view', 'Lihat'),
    ('contact_messages:reply', 'contact_messages', 'Pesan kontak', 'reply', 'Balas'),
    ('activity_logs:view', 'activity_logs', 'Log aktivitas', 'view', 'Lihat');

INSERT INTO role_permissions (role_id, permission_key)
SELECT '00000000-0000-4000-8000-000000000001', key
FROM permissions
WHERE key IN ('contact_messages:view', 'contact_messages:reply', 'activity_logs:view');
