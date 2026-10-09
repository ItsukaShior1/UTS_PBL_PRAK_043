-- 002_create_students.sql
-- Tabel students: data mahasiswa, terikat 1:1 ke users via user_id.
-- Mendukung soft delete (deleted_at).

CREATE TABLE IF NOT EXISTS students (
    id            SERIAL        PRIMARY KEY,
    user_id       INTEGER       NOT NULL,
    nim           VARCHAR(20)   NOT NULL,
    nama          VARCHAR(255)  NOT NULL,
    prodi         VARCHAR(100)  NOT NULL,
    angkatan      INTEGER       NOT NULL,
    ipk_terakhir  NUMERIC(3,2),
    deleted_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT students_user_id_key UNIQUE (user_id),
    CONSTRAINT students_nim_key     UNIQUE (nim),
    CONSTRAINT students_angkatan_check
        CHECK (angkatan BETWEEN 2000 AND 2100),
    CONSTRAINT students_ipk_check
        CHECK (ipk_terakhir IS NULL OR (ipk_terakhir >= 0 AND ipk_terakhir <= 4.00)),
    CONSTRAINT students_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS students_prodi_idx    ON students (prodi);
CREATE INDEX IF NOT EXISTS students_angkatan_idx ON students (angkatan);
CREATE INDEX IF NOT EXISTS students_deleted_at_idx ON students (deleted_at);
