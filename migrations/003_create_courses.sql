-- 003_create_courses.sql
-- Tabel courses: master mata kuliah.
-- Kuota akan dibandingkan dengan COUNT(enrollments.course_id) untuk slot tersisa.

CREATE TABLE IF NOT EXISTS courses (
    id            SERIAL        PRIMARY KEY,
    kode_mk       VARCHAR(20)   NOT NULL,
    nama_mk       VARCHAR(255)  NOT NULL,
    sks           INTEGER       NOT NULL,
    semester      INTEGER       NOT NULL,
    kuota         INTEGER       NOT NULL,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT courses_kode_mk_key UNIQUE (kode_mk),
    CONSTRAINT courses_sks_check
        CHECK (sks BETWEEN 1 AND 6),
    CONSTRAINT courses_semester_check
        CHECK (semester BETWEEN 1 AND 8),
    CONSTRAINT courses_kuota_check
        CHECK (kuota > 0)
);

CREATE INDEX IF NOT EXISTS courses_semester_idx ON courses (semester);
