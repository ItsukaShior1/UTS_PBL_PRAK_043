-- 005_seed_data.sql
-- Seeder: 1 admin, 20 mahasiswa, 10 mata kuliah.
--
-- Hash bcrypt sudah di-generate via genhash/main.go (terpisah dari project).
-- Hash yang dipakai:
--   admin123      -> $2a$10$vaEDCPpe54zAQBQIZwK2cOKDPyhAq42vU35ocjNpZJktlK8FTucKW
--   mahasiswa123  -> $2a$10$ly3QSGwGjq77U2Op/n/3sOEJOgzAb38xQKxIRANHTkhPXO6INwnVa

BEGIN;

-- ============================================================
-- 1 ADMIN
-- ============================================================
-- email: admin@siakad.test  |  password: admin123
INSERT INTO users (email, password_hash, role)
VALUES ('admin@siakad.test', '$2a$10$vaEDCPpe54zAQBQIZwK2cOKDPyhAq42vU35ocjNpZJktlK8FTucKW', 'admin')
ON CONFLICT (email) DO NOTHING;

-- ============================================================
-- 10 MATA KULIAH
-- ============================================================
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES
    ('IF101', 'Algoritma dan Pemrograman',     3, 1, 40),
    ('IF102', 'Matematika Diskrit',             3, 1, 35),
    ('IF201', 'Struktur Data',                  4, 2, 35),
    ('IF202', 'Basis Data',                     3, 2, 40),
    ('IF301', 'Pemrograman Web',                3, 3, 40),
    ('IF302', 'Sistem Operasi',                 3, 3, 35),
    ('IF401', 'Jaringan Komputer',              3, 4, 30),
    ('IF402', 'Rekayasa Perangkat Lunak',       3, 4, 30),
    ('IF501', 'Kecerdasan Buatan',              3, 5, 30),
    ('IF502', 'Machine Learning',               3, 6, 25)
ON CONFLICT (kode_mk) DO NOTHING;

-- ============================================================
-- 20 MAHASISWA
-- ============================================================
-- Loop generate 20 mahasiswa:
--   - email: <nim>@siakad.test
--   - password awal semua sama: mahasiswa123
--   - nim: 2024XXXXXXXX (12 digit)
--   - prodi: Teknik Informatika (ganjil) atau Sistem Informasi (genap)
--   - angkatan: 2024
--   - ipk_terakhir: random 1.50 - 4.00
DO $$
DECLARE
    i INT;
    v_user_id INT;
    v_nim TEXT;
    v_nama TEXT;
    v_email TEXT;
    v_prodi TEXT;
    v_ipk NUMERIC(3,2);
    v_hash TEXT := '$2a$10$ly3QSGwGjq77U2Op/n/3sOEJOgzAb38xQKxIRANHTkhPXO6INwnVa';
BEGIN
    FOR i IN 1..20 LOOP
        v_nim   := '2024' || LPAD(i::TEXT, 8, '0');
        v_email := v_nim || '@siakad.test';
        v_nama  := 'Mahasiswa ' || i;
        v_prodi := CASE WHEN i % 2 = 0 THEN 'Sistem Informasi' ELSE 'Teknik Informatika' END;
        v_ipk   := ROUND((1.5 + random() * 2.5)::NUMERIC, 2);  -- 1.50 - 4.00
        IF v_ipk > 4.00 THEN v_ipk := 4.00; END IF;

        INSERT INTO users (email, password_hash, role)
        VALUES (v_email, v_hash, 'mahasiswa')
        ON CONFLICT (email) DO NOTHING
        RETURNING id INTO v_user_id;

        -- Jika user sudah ada (re-run), ambil id-nya
        IF v_user_id IS NULL THEN
            SELECT id INTO v_user_id FROM users WHERE email = v_email;
        END IF;

        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
        VALUES (v_user_id, v_nim, v_nama, v_prodi, 2024, v_ipk)
        ON CONFLICT (nim) DO NOTHING;
    END LOOP;
END $$;

COMMIT;

-- Verifikasi
SELECT 'users'      AS tabel, COUNT(*) AS total FROM users
UNION ALL
SELECT 'students',  COUNT(*) FROM students
UNION ALL
SELECT 'courses',   COUNT(*) FROM courses
UNION ALL
SELECT 'enrollments', COUNT(*) FROM enrollments;
