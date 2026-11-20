CREATE TABLE permissions (
    key        VARCHAR(64) PRIMARY KEY,
    menu       VARCHAR(32) NOT NULL,
    menu_label VARCHAR(64) NOT NULL,
    action     VARCHAR(32) NOT NULL,
    label      VARCHAR(64) NOT NULL
);

CREATE TABLE roles (
    id         UUID PRIMARY KEY,
    name       VARCHAR(64) NOT NULL UNIQUE,
    is_system  BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_roles_default ON roles (is_default) WHERE is_default=TRUE;

CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_key VARCHAR(64) NOT NULL REFERENCES permissions(key) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_key)
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id),
    PRIMARY KEY (user_id, role_id)
);

INSERT INTO permissions (key, menu, menu_label, action, label) VALUES
    ('dashboard:view', 'dashboard', 'Dasbor', 'view', 'Lihat'),
    ('products:view', 'products', 'Produk', 'view', 'Lihat'),
    ('products:create', 'products', 'Produk', 'create', 'Tambah'),
    ('products:edit', 'products', 'Produk', 'edit', 'Ubah'),
    ('products:delete', 'products', 'Produk', 'delete', 'Hapus'),
    ('categories:view', 'categories', 'Kategori', 'view', 'Lihat'),
    ('categories:create', 'categories', 'Kategori', 'create', 'Tambah'),
    ('categories:edit', 'categories', 'Kategori', 'edit', 'Ubah'),
    ('categories:delete', 'categories', 'Kategori', 'delete', 'Hapus'),
    ('transactions:view', 'transactions', 'Transaksi', 'view', 'Lihat'),
    ('users:view', 'users', 'Pengguna admin', 'view', 'Lihat'),
    ('users:create', 'users', 'Pengguna admin', 'create', 'Tambah'),
    ('users:edit', 'users', 'Pengguna admin', 'edit', 'Ubah'),
    ('users:delete', 'users', 'Pengguna admin', 'delete', 'Hapus'),
    ('users:assign_roles', 'users', 'Pengguna admin', 'assign_roles', 'Atur role'),
    ('roles:view', 'roles', 'Role', 'view', 'Lihat'),
    ('roles:create', 'roles', 'Role', 'create', 'Tambah'),
    ('roles:edit', 'roles', 'Role', 'edit', 'Ubah'),
    ('roles:delete', 'roles', 'Role', 'delete', 'Hapus');

INSERT INTO roles (id, name, is_system, is_default) VALUES
    ('00000000-0000-4000-8000-000000000001', 'Super Admin', TRUE, FALSE),
    ('00000000-0000-4000-8000-000000000002', 'Catalog Manager', FALSE, FALSE),
    ('00000000-0000-4000-8000-000000000003', 'Viewer', FALSE, TRUE);

INSERT INTO role_permissions (role_id, permission_key)
SELECT '00000000-0000-4000-8000-000000000001', key FROM permissions;

INSERT INTO role_permissions (role_id, permission_key)
SELECT '00000000-0000-4000-8000-000000000002', key
FROM permissions
WHERE menu IN ('products', 'categories');

INSERT INTO role_permissions (role_id, permission_key)
SELECT '00000000-0000-4000-8000-000000000003', key
FROM permissions
WHERE action = 'view'
  AND menu IN ('dashboard', 'products', 'categories', 'transactions');

INSERT INTO user_roles (user_id, role_id)
SELECT id, '00000000-0000-4000-8000-000000000001'
FROM users;
