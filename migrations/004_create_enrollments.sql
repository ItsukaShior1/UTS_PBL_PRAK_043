-- 004_create_enrollments.sql
-- Tabel enrollments: KRS (Kartu Rencana Studi).
-- Unique constraint mencegah mahasiswa mengambil MK yang sama pada tahun akademik yang sama.

CREATE TABLE IF NOT EXISTS enrollments (
    id              SERIAL        PRIMARY KEY,
    student_id      INTEGER       NOT NULL,
    course_id       INTEGER       NOT NULL,
    tahun_akademik  VARCHAR(20)   NOT NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT enrollments_unique
        UNIQUE (student_id, course_id, tahun_akademik),
    CONSTRAINT enrollments_student_id_fkey
        FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    CONSTRAINT enrollments_course_id_fkey
        FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE RESTRICT,
    CONSTRAINT enrollments_ta_check
        CHECK (tahun_akademik ~ '^[0-9]{4}/[0-9]{4}-(Ganjil|Genap)$')
);

CREATE INDEX IF NOT EXISTS enrollments_student_idx ON enrollments (student_id);
CREATE INDEX IF NOT EXISTS enrollments_course_idx  ON enrollments (course_id);
CREATE INDEX IF NOT EXISTS enrollments_ta_idx      ON enrollments (tahun_akademik);
