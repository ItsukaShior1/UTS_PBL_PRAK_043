-- 001_create_users.sql
-- Tabel users: menyimpan akun login untuk admin dan mahasiswa.
-- Kolom: id, email (unik), password_hash (bcrypt), role, is_active, timestamps.

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL        PRIMARY KEY,
    email         VARCHAR(255)  NOT NULL,
    password_hash VARCHAR(255)  NOT NULL,
    role          VARCHAR(20)   NOT NULL,
    is_active     BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT users_role_check
        CHECK (role IN ('admin', 'mahasiswa'))
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_key
    ON users (email);
