DELETE FROM role_permissions
WHERE permission_key IN ('contact_messages:view', 'contact_messages:reply', 'activity_logs:view');
DELETE FROM permissions
WHERE key IN ('contact_messages:view', 'contact_messages:reply', 'activity_logs:view');

DROP TABLE IF EXISTS activity_log;

ALTER TABLE contact_message
    DROP COLUMN IF EXISTS replied_by,
    DROP COLUMN IF EXISTS replied_at,
    DROP COLUMN IF EXISTS reply_message;
