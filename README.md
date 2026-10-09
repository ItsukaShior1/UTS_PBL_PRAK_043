# API SIAKAD Mini

RESTful API untuk SIAKAD Mini — layanan akademik sederhana yang mengelola data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS). Dibangun dengan **Go + Fiber v2 + PostgreSQL** dengan autentikasi JWT.

## Stack

- **Bahasa**: Go 1.22+
- **Framework HTTP**: Fiber v2
- **Database**: PostgreSQL 14+ (driver pgx v5)
- **Auth**: JWT (HS256)
- **Password**: bcrypt (cost 10)
- **Pattern**: Clean architecture (model / repository / service / route)

## Struktur Folder

```
api-siakad/
├── app/
│   ├── model/         # DTO & entity
│   ├── repository/    # query ke PostgreSQL
│   └── service/       # business rules + handler HTTP
├── config/            # env loader
├── database/          # pgxpool
├── helper/            # response envelope, pagination, token JWT
├── middleware/        # require auth, require role, require JSON
├── migrations/        # SQL schema & seeder
├── route/             # pendaftaran endpoint
├── main.go            # composition root
├── go.mod
└── .env
```

## Menjalankan

### Prasyarat

- Go 1.22+ (tested di 1.26.5)
- PostgreSQL 14+ (tested di 18.4)
- psql CLI

### Langkah

1. **Buat database**
   ```powershell
   psql -U postgres
   CREATE DATABASE siakad_mini;
   \q
   ```

2. **Jalankan migration + seeder**
   ```powershell
   cd "C:\Users\LENOVO\OneDrive\semester 5\Prak\Backend Prak\uts"
   psql -U postgres -d siakad_mini -f migrations/001_create_users.sql
   psql -U postgres -d siakad_mini -f migrations/002_create_students.sql
   psql -U postgres -d siakad_mini -f migrations/003_create_courses.sql
   psql -U postgres -d siakad_mini -f migrations/004_create_enrollments.sql
   psql -U postgres -d siakad_mini -f migrations/005_seed_data.sql
   ```

3. **Salin & edit `.env`**
   ```powershell
   Copy-Item .env.example .env
   notepad .env
   ```
   Isi `DB_PASSWORD` dan `JWT_SECRET` (string acak min 32 karakter).

4. **Build & jalankan**
   ```powershell
   go mod tidy
   go run .
   ```
   Server jalan di `http://localhost:3000`.

## Kredensial Default (dari seeder)

| Role | Email | Password |
|---|---|---|
| Admin | `admin@siakad.test` | `admin123` |
| Mahasiswa 1–20 | `20240000000X@siakad.test` (X=1..20) | `mahasiswa123` |

Catatan: Password awal setiap mahasiswa sama dengan NIM-nya saat akun dibuat via endpoint `POST /api/v1/students`.

## Endpoint

| # | Method | Path | Akses | Fungsi |
|---|---|---|---|---|
| 1 | POST | `/api/v1/auth/login` | Publik | Login |
| 2 | GET | `/api/v1/auth/me` | Semua role | Profil sendiri |
| 3 | GET | `/api/v1/students` | Admin | List + filter + search + sort |
| 4 | POST | `/api/v1/students` | Admin | Buat mahasiswa + akun |
| 5 | GET | `/api/v1/students/:id` | Admin, Mhs (data sendiri) | Detail + KRS + total/batas/sisa SKS |
| 6 | PUT | `/api/v1/students/:id` | Admin | Update data |
| 7 | DELETE | `/api/v1/students/:id` | Admin | Soft delete |
| 8 | GET | `/api/v1/courses` | Semua role | List + sisa kuota |
| 9 | POST | `/api/v1/enrollments` | Mahasiswa | Ambil mata kuliah |
| 10 | DELETE | `/api/v1/enrollments/:id` | Mahasiswa (miliknya) | Batal KRS |

## Business Rules

1. **Batas SKS per semester** (berdasarkan IPK terakhir):
   - IPK ≥ 3.00 → max 24 SKS
   - IPK 2.50–2.99 → max 21 SKS
   - IPK < 2.50 → max 18 SKS

2. **Satu mata kuliah hanya boleh diambil satu kali per tahun akademik** (UNIQUE constraint + pengecekan di transaction).

3. **Mata kuliah penuh (kuota habis) tidak bisa diambil** — dicek dengan `SELECT ... FOR UPDATE` di transaction.

4. **Mahasiswa hanya boleh akses KRS miliknya sendiri.**

5. **Login dilindungi rate limit**: 5x gagal per IP+email → kunci 1 menit.

## Status Code

| Code | Penggunaan |
|---|---|
| 200 | Sukses baca / update |
| 201 | Sukses create |
| 204 | Sukses delete (no body) |
| 401 | Token tidak ada / tidak valid / kedaluwarsa |
| 403 | Role tidak berhak / akses data orang lain |
| 404 | Data tidak ditemukan / sudah dihapus |
| 409 | Duplikat (NIM/email/MK di TA yang sama) |
| 422 | Validasi gagal / kuota penuh / SKS melebihi |
| 429 | Rate limit login |
| 500 | Error server tak terduga |

## Testing

Lihat `postman/PANDUAN-POSTMAN.md` (di folder lokal) untuk skenario uji lengkap.

## Lisensi

UTS Praktikum Backend — © 2026
