# 📘 PANDUAN UTS PBE — SIAKAD Mini API

> **Studi Kasus:** RESTful API SIAKAD Mini (Sistem Akademik Mini) untuk mengelola data mahasiswa, mata kuliah, dan KRS.
> **Stack:** Go + Fiber v2 + PostgreSQL (pgx v5) + JWT
> **Basis:** Build dari NOL (tidak reuse kode dari project sebelumnya), tinggal copy-paste.

---

## 📑 Daftar Isi

1. [Pemahaman Modul](#1-pemahaman-modul)
2. [Prasyarat](#2-prasyarat)
3. [Setup Database PostgreSQL](#3-setup-database-postgresql)
4. [Inisialisasi Proyek Go](#4-inisialisasi-proyek-go)
5. [Konfigurasi (.env, config, database)](#5-konfigurasi-env-config-database)
6. [Helper (Response, Pagination, Token)](#6-helper-response-pagination-token)
7. [Middleware](#7-middleware)
8. [Model](#8-model)
9. [Repository](#9-repository)
10. [Service & Business Rules](#10-service--business-rules)
11. [Routes](#11-routes)
12. [Main.go (Composition Root)](#12-maingo-composition-root)
13. [Testing Manual dengan curl/Postman](#13-testing-manual-dengan-curlpostman)
14. [Commit Bertahap ke GitHub](#14-commit-bertahap-ke-github)
15. [Checklist Kepatuhan Modul](#15-checklist-kepatuhan-modul)

---

## 1. Pemahaman Modul

API punya **10 endpoint** untuk 2 role (`admin` & `mahasiswa`):

| # | Method | Endpoint | Akses | Fungsi |
|---|--------|----------|-------|--------|
| 1 | POST | `/api/v1/auth/login` | Publik | Login, kembalikan access token |
| 2 | GET  | `/api/v1/auth/me` | Semua role | Profil user yang sedang login |
| 3 | GET  | `/api/v1/students` | Admin | Daftar mahasiswa + pagination/filter/search |
| 4 | POST | `/api/v1/students` | Admin | Buat mahasiswa + akun user (1 transaction) |
| 5 | GET  | `/api/v1/students/{id}` | Admin, Mahasiswa (data sendiri) | Detail + courses + total_sks + batas_sks |
| 6 | PUT  | `/api/v1/students/{id}` | Admin | Update data mahasiswa |
| 7 | DELETE | `/api/v1/students/{id}` | Admin | Soft delete |
| 8 | GET  | `/api/v1/courses` | Semua role | Daftar mata kuliah + sisa kuota |
| 9 | POST | `/api/v1/enrollments` | Mahasiswa | Ambil mata kuliah (KRS) |
| 10 | DELETE | `/api/v1/enrollments/{id}` | Mahasiswa (milik sendiri) | Batal KRS |

**Business rules penting:**

1. **Batas SKS by IPK** — IPK ≥ 3.00 → 24 SKS, IPK 2.50–2.99 → 21 SKS, IPK < 2.50 → 18 SKS.
2. **Tidak boleh ambil MK yang sama** pada tahun akademik yang sama (UNIQUE constraint).
3. **MK penuh** (kuota habis) tidak bisa diambil.
4. **Mahasiswa hanya akses KRS miliknya sendiri.**
5. **Login:** format email valid, password min 8 char, rate limit 5x/menit.

**Status code yang dipakai:** `200, 201, 204, 401, 403, 404, 409, 422, 429, 500`.

**Format response seragam:**

```json
// Sukses
{ "success": true, "message": "...", "data": {...}, "meta": {...} }

// Error
{ "success": false, "message": "...", "errors": { "field": ["..."] } }
```

---

## 2. Prasyarat

Pastikan sudah terinstall:

| Tools | Versi Minimal | Cek |
|---|---|---|
| Go | 1.22+ | `go version` |
| PostgreSQL | 14+ | `psql --version` |
| Git | 2.x | `git --version` |
| Postman / curl | - | untuk testing |

> 💡 **Kamu pakai PowerShell (Windows).** Semua command di panduan ini sudah disesuaikan:
> - Pakai baris baru atau `;` (PowerShell **TIDAK support** `&&`).
> - Path pakai `C:/...` (bukan `/tmp/...`).
>
> Kalau PostgreSQL-mu pakai **Supabase/Neon/Railway**, cukup catat connection string-nya. Step [Setup Database](#3-setup-database-postgresql) akan menyesuaikan.

---

## 3. Setup Database PostgreSQL

Karena kita **tidak pakai database project sebelumnya**, kita buat database baru `siakad_mini`.

### 3.1. Buat database

```bash
# Masuk sebagai superuser postgres
psql -U postgres

# Di prompt psql:
CREATE DATABASE siakad_mini;
\q
```

### 3.2. Jalankan migration

File migration sudah aku isi (lihat folder `migrations/`). Jalankan berurutan:

```powershell
cd "C:/Users/LENOVO/OneDrive/semester 5/Prak/Backend Prak/uts"

psql -U postgres -d siakad_mini -f migrations/001_create_users.sql
psql -U postgres -d siakad_mini -f migrations/002_create_students.sql
psql -U postgres -d siakad_mini -f migrations/003_create_courses.sql
psql -U postgres -d siakad_mini -f migrations/004_create_enrollments.sql
```

### 3.3. Generate bcrypt hash untuk seeder

> ✅ **Hash-nya sudah aku generate** dan placeholder di `005_seed_data.sql` sudah terisi.
> Lokasi hash di file SQL: cari `'$2a$10$...'`. Lewati step ini kalau tidak ingin ubah password default.

Folder `genhash/` di root project adalah generator sandbox. Cara pakainya (PowerShell):

```powershell
cd "C:/Users/LENOVO/OneDrive/semester 5/Prak/Backend Prak/uts/genhash"
go get golang.org/x/crypto/bcrypt
go run main.go admin123
go run main.go pass123
```

Output (contoh):
```
$2a$10$wNOMU5.hOd0qu1SFtVwNCORDb1Rbl39DaX4h.7AY1Mf.ff6bVrxku   # admin123
$2a$10$KjsiSv/GysgEeBprWejZ/OeuskJ7OfJy3cqm76eTB.WZsAPTn3.ZS   # pass123
```

Kalau kamu ganti password, replace manual di `migrations/005_seed_data.sql` (cari string `$2a$10$`).

> 💡 Folder `genhash/` **tidak** di-commit ke repo utama (sudah masuk `.gitignore`).
> Module name `genhash` berbeda dari `api-siakad` — tidak akan bentrok.

### 3.4. Jalankan seeder

```powershell
psql -U postgres -d siakad_mini -f migrations/005_seed_data.sql
```

Verifikasi harus muncul:

```
   tabel    | total
-----------+-------
 users     |    21
 students  |    20
 courses   |    10
 enrollments|    0
```

### ✅ Commit Step 3

```bash
git add migrations/
git commit -m "feat(db): schema SIAKAD + seeder 1 admin, 20 mhs, 10 mk"
```

---

## 4. Inisialisasi Proyek Go

> ⚠️ Pakai PowerShell, **JANGAN** pakai `&&`. Pisahkan command dengan baris baru atau `;`.

```powershell
cd "C:/Users/LENOVO/OneDrive/semester 5/Prak/Backend Prak/uts"

# Inisialisasi module (di root project, BUKAN di folder genhash/)
go mod init api-siakad

# Download dependencies (jalankan satu-satu atau pisah dengan `;`)
go get github.com/gofiber/fiber/v2
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go get github.com/joho/godotenv
```

`.gitignore` sudah aku sediakan di root project (isi: `genhash/`, `.env`, `*.exe`, dll).

Buat `.env` dari template:

```powershell
Copy-Item .env.example .env
notepad .env
# Isi DB_PASSWORD, JWT_SECRET, dll. Lalu save.
```

# Env
.env
.env.local

# Logs
logs/
*.log

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
```

Buat `.env.example` (template, jangan di-commit `.env` asli):

```env
APP_PORT=3000
APP_MODE=debug

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=siakad_mini
DB_SSLMODE=disable

JWT_SECRET=ganti-dengan-string-acak-minimal-32-karakter-yang-aman
JWT_ACCESS_TTL_MINUTES=60

LOGIN_MAX_FAILS=5
LOGIN_LOCKOUT_MINUTES=1
```

### ✅ Commit Step 4

```bash
git add go.mod go.sum .gitignore .env.example
git commit -m "chore: init go module + dependencies + .env.example"
```

---

## 5. Konfigurasi (`config/`, `database/`)

### 5.1. `config/env.go`

```go
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("peringatan: .env tidak ditemukan, pakai environment sistem")
	}
}

func GetEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("peringatan: %s bukan angka (%q), pakai bawaan %d", key, v, fallback)
		return fallback
	}
	return n
}
```

### 5.2. `database/postgres.go`

```go
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"api-siakad/config"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.GetEnv("DB_USER", "postgres"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "5432"),
		config.GetEnv("DB_NAME", "siakad_mini"),
		config.GetEnv("DB_SSLMODE", "disable"),
	)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("konfigurasi DB tidak valid: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal ping DB: %w", err)
	}
	return pool, nil
}
```

### ✅ Commit Step 5

```bash
git add config/ database/
git commit -m "feat(infra): env loader + pgxpool postgres"
```

---

## 6. Helper (`helper/`)

### 6.1. `helper/response.go` (envelope JSON)

```go
package helper

import "github.com/gofiber/fiber/v2"

type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

func Ok(c *fiber.Ctx, msg string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{Success: true, Message: msg, Data: data})
}

func Created(c *fiber.Ctx, msg string, data interface{}, location string) error {
	if location != "" {
		c.Set("Location", location)
	}
	return c.Status(fiber.StatusCreated).JSON(WebResponse{Success: true, Message: msg, Data: data})
}

func OkWithMeta(c *fiber.Ctx, msg string, data interface{}, m *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{Success: true, Message: msg, Data: data, Meta: m})
}

func NoContent(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNoContent).Send(nil)
}

func Fail(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(WebResponse{Success: false, Message: msg})
}

func FailValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false,
		Message: "Validasi gagal",
		Errors:  errs,
	})
}
```

### 6.2. `helper/pagination.go` (parse query + hitung meta)

```go
package helper

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func ParsePagination(c *fiber.Ctx, defaultPerPage, maxPerPage int) (page, perPage int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ = strconv.Atoi(c.Query("per_page", strconv.Itoa(defaultPerPage)))
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage
}

func BuildMeta(page, perPage, total int) *Meta {
	last := 0
	if perPage > 0 {
		last = (total + perPage - 1) / perPage
	}
	return &Meta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    last,
	}
}
```

### 6.3. `helper/token.go` (JWT issue/parse)

```go
package helper

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrTokenExpired = errors.New("token kedaluwarsa")

type Claims struct {
	UserID int    `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret), ttl: ttl}
}

func (t *TokenIssuer) Issue(userID int, email, role string) (string, time.Time, error) {
	exp := time.Now().Add(t.ttl)
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(t.secret)
	return signed, exp, err
}

func (t *TokenIssuer) Parse(raw string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing tidak didukung")
		}
		return t.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("token tidak valid")
	}
	return claims, nil
}
```

### ✅ Commit Step 6

```bash
git add helper/
git commit -m "feat(helper): response envelope, pagination, JWT issuer"
```

---

## 7. Middleware (`middleware/`)

### 7.1. `middleware/require_json.go`

```go
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"api-siakad/helper"
)

func RequireJSON() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ct := c.Get("Content-Type")
		if ct != "" && len(ct) >= 16 && ct[:16] != "application/json" {
			// Longgar: hanya cek prefix, beberapa client pakai charset
			if !startsWith(ct, "application/json") {
				return helper.Fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
			}
		}
		return c.Next()
	}
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
```

### 7.2. `middleware/require_auth.go`

```go
package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"api-siakad/helper"
)

const (
	LocalsUserID = "user_id"
	LocalsEmail  = "email"
	LocalsRole   = "role"
)

func RequireAuth(issuer *helper.TokenIssuer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			c.Set("WWW-Authenticate", `Bearer realm="siakad"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "Token wajib dikirim")
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			c.Set("WWW-Authenticate", `Bearer realm="siakad"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid")
		}
		claims, err := issuer.Parse(parts[1])
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="siakad"`)
			if errors.Is(err, helper.ErrTokenExpired) {
				return helper.Fail(c, fiber.StatusUnauthorized, "Token kedaluwarsa")
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid")
		}
		c.Locals(LocalsUserID, claims.UserID)
		c.Locals(LocalsEmail, claims.Email)
		c.Locals(LocalsRole, claims.Role)
		return c.Next()
	}
}
```

### 7.3. `middleware/require_role.go`

```go
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"api-siakad/helper"
)

func GetUser(c *fiber.Ctx) (userID int, email, role string, ok bool) {
	uid, ok1 := c.Locals(LocalsUserID).(int)
	e, ok2 := c.Locals(LocalsEmail).(string)
	r, ok3 := c.Locals(LocalsRole).(string)
	if !ok1 || !ok2 || !ok3 {
		return 0, "", "", false
	}
	return uid, e, r, true
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, _, role, ok := GetUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "Belum terautentikasi")
		}
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return helper.Fail(c, fiber.StatusForbidden, "Role Anda tidak berhak mengakses endpoint ini")
	}
}
```

### ✅ Commit Step 7

```bash
git add middleware/
git commit -m "feat(middleware): require auth, role guard, json guard"
```

---

## 8. Model (`model/`)

### 8.1. `model/user.go`

```go
package model

import "time"

const (
	RoleAdmin     = "admin"
	RoleMahasiswa = "mahasiswa"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"`
	User        UserResponse `json:"user"`
}

type MeResponse struct {
	User    UserResponse      `json:"user"`
	Student *StudentResponse `json:"student,omitempty"`
}
```

### 8.2. `model/student.go`

```go
package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type StudentResponse struct {
	ID          int     `json:"id"`
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type CreateStudentRequest struct {
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Email       string  `json:"email"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type UpdateStudentRequest struct {
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type ListStudentQuery struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan int
	Search   string
	Sort     string // "nama", "-nama", "ipk_terakhir", "-ipk_terakhir"
}

func (q ListStudentQuery) Offset() int { return (q.Page - 1) * q.PerPage }

type StudentDetailResponse struct {
	Student  StudentResponse  `json:"student"`
	Courses  []CourseResponse `json:"courses"`
	TotalSKS int              `json:"total_sks"`
	BatasSKS int              `json:"batas_sks"`
	SisaSKS  int              `json:"sisa_sks"`
}
```

### 8.3. `model/course.go`

```go
package model

import "time"

type Course struct {
	ID        int       `json:"id"`
	KodeMK    string    `json:"kode_mk"`
	NamaMK    string    `json:"nama_mk"`
	SKS       int       `json:"sks"`
	Semester  int       `json:"semester"`
	Kuota     int       `json:"kuota"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CourseResponse struct {
	ID        int    `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`
	SisaKuota int    `json:"sisa_kuota"`
}

type ListCourseQuery struct {
	Semester  int
	Search    string
	Available bool
}
```

### 8.4. `model/enrollment.go`

```go
package model

import "time"

type Enrollment struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id"`
	CourseID      int       `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateEnrollmentRequest struct {
	CourseID      int    `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}

type EnrollmentWithCourse struct {
	Enrollment Enrollment
	Course     Course
}
```

### ✅ Commit Step 8

```bash
git add model/
git commit -m "feat(model): DTO untuk user, student, course, enrollment"
```

---

## 9. Repository (`repository/`)

### 9.1. `repository/errors.go`

```go
package repository

import "errors"

var (
	ErrNotFound       = errors.New("data tidak ditemukan")
	ErrDuplicate      = errors.New("data sudah ada")
	ErrAlreadyEnrolled = errors.New("sudah pernah diambil pada tahun akademik ini")
	ErrKuotaPenuh     = errors.New("kuota mata kuliah penuh")
	ErrSKSExceed      = errors.New("total SKS melebihi batas")
)

// BatasSKS mengembalikan batas SKS berdasarkan IPK.
//   IPK >= 3.00 → 24 SKS
//   IPK 2.50..2.99 → 21 SKS
//   IPK < 2.50 → 18 SKS
func BatasSKS(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}
```

### 9.2. `repository/user_repository.go`

```go
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-siakad/app/model"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
}

type userRepo struct{ pool *pgxpool.Pool }

func NewUserRepository(p *pgxpool.Pool) UserRepository { return &userRepo{p} }

func (r *userRepo) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx, `SELECT id, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE email = $1 AND is_active = TRUE`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	return u, err
}

func (r *userRepo) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx, `SELECT id, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	return u, err
}

// Helper untuk translate pgconn unique violation
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
```

### 9.3. `repository/student_repository.go`

```go
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-siakad/app/model"
)

type StudentRepository interface {
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindByUserID(ctx context.Context, userID int) (model.Student, error)
	CreateWithUser(ctx context.Context, u model.User, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
	List(ctx context.Context, q model.ListStudentQuery) ([]model.Student, int, error)
}

type studentRepo struct{ pool *pgxpool.Pool }

func NewStudentRepository(p *pgxpool.Pool) StudentRepository { return &studentRepo{p} }

const studentCols = `id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at`

func scanStudent(row pgx.Row) (model.Student, error) {
	var s model.Student
	err := row.Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan,
		&s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (r *studentRepo) FindByID(ctx context.Context, id int) (model.Student, error) {
	s, err := scanStudent(r.pool.QueryRow(ctx, `SELECT `+studentCols+` FROM students WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}
	return s, err
}

func (r *studentRepo) FindByUserID(ctx context.Context, userID int) (model.Student, error) {
	s, err := scanStudent(r.pool.QueryRow(ctx, `SELECT `+studentCols+` FROM students
		WHERE user_id = $1 AND deleted_at IS NULL`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}
	return s, err
}

// CreateWithUser: insert users + students dalam SATU TRANSACTION.
func (r *studentRepo) CreateWithUser(ctx context.Context, u model.User, s model.Student) (model.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Student{}, err
	}
	defer tx.Rollback(ctx)

	// 1) Insert user
	err = tx.QueryRow(ctx, `INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, 'mahasiswa') RETURNING id, is_active, created_at, updated_at`,
		u.Email, u.PasswordHash).
		Scan(&u.ID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if IsUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, err
	}

	// 2) Insert student
	err = tx.QueryRow(ctx, `INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+studentCols,
		u.ID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir).
		Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan,
			&s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if IsUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Student{}, err
	}
	return s, nil
}

func (r *studentRepo) Update(ctx context.Context, s model.Student) (model.Student, error) {
	row := r.pool.QueryRow(ctx, `UPDATE students SET nama=$1, prodi=$2, angkatan=$3, ipk_terakhir=$4
		WHERE id = $5 AND deleted_at IS NULL RETURNING `+studentCols,
		s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir, s.ID)
	res, err := scanStudent(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}
	return res, err
}

func (r *studentRepo) SoftDelete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *studentRepo) List(ctx context.Context, q model.ListStudentQuery) ([]model.Student, int, error) {
	conds := []string{"s.deleted_at IS NULL"}
	args := []any{}
	idx := 1

	if q.Prodi != "" {
		conds = append(conds, fmt.Sprintf("s.prodi = $%d", idx))
		args = append(args, q.Prodi)
		idx++
	}
	if q.Angkatan != 0 {
		conds = append(conds, fmt.Sprintf("s.angkatan = $%d", idx))
		args = append(args, q.Angkatan)
		idx++
	}
	if q.Search != "" {
		conds = append(conds, fmt.Sprintf("(s.nim ILIKE $%d OR s.nama ILIKE $%d)", idx, idx))
		args = append(args, "%"+q.Search+"%")
		idx++
	}
	where := " WHERE " + strings.Join(conds, " AND ")

	// total
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students s"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// sort
	orderBy := "s.nama ASC"
	switch q.Sort {
	case "nama":
		orderBy = "s.nama ASC"
	case "-nama":
		orderBy = "s.nama DESC"
	case "ipk_terakhir":
		orderBy = "s.ipk_terakhir ASC"
	case "-ipk_terakhir":
		orderBy = "s.ipk_terakhir DESC, s.nama ASC"
	}

	args = append(args, q.PerPage, q.Offset())
	limitIdx := idx
	offsetIdx := idx + 1
	sql := fmt.Sprintf(`SELECT %s FROM students s%s ORDER BY %s LIMIT $%d OFFSET $%d`,
		studentCols, where, orderBy, limitIdx, offsetIdx)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []model.Student{}
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, nil
}
```

### 9.4. `repository/course_repository.go`

```go
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-siakad/app/model"
)

type CourseRepository interface {
	FindByID(ctx context.Context, id int) (model.Course, error)
	List(ctx context.Context, q model.ListCourseQuery) ([]CourseWithStats, error)
}

type CourseWithStats struct {
	model.Course
	Terisi int
}

type courseRepo struct{ pool *pgxpool.Pool }

func NewCourseRepository(p *pgxpool.Pool) CourseRepository { return &courseRepo{p} }

func (r *courseRepo) FindByID(ctx context.Context, id int) (model.Course, error) {
	var c model.Course
	err := r.pool.QueryRow(ctx, `SELECT id, kode_mk, nama_mk, sks, semester, kuota, created_at, updated_at
		FROM courses WHERE id = $1`, id).
		Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Course{}, ErrNotFound
	}
	return c, err
}

func (r *courseRepo) List(ctx context.Context, q model.ListCourseQuery) ([]CourseWithStats, error) {
	conds := []string{"1=1"}
	args := []any{}
	idx := 1

	if q.Semester != 0 {
		conds = append(conds, fmt.Sprintf("c.semester = $%d", idx))
		args = append(args, q.Semester)
		idx++
	}
	if q.Search != "" {
		conds = append(conds, fmt.Sprintf("(c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", idx, idx))
		args = append(args, "%"+q.Search+"%")
		idx++
	}
	where := " WHERE " + strings.Join(conds, " AND ")

	sql := `SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, c.created_at, c.updated_at,
        COALESCE((SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id), 0) AS terisi
        FROM courses c` + where + ` ORDER BY c.semester, c.kode_mk`

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []CourseWithStats{}
	for rows.Next() {
		var c CourseWithStats
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota,
			&c.CreatedAt, &c.UpdatedAt, &c.Terisi); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}
```

### 9.5. `repository/enrollment_repository.go`

```go
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-siakad/app/model"
)

type EnrollmentRepository interface {
	// EnrollAtomic: cek duplikat, lock course row, hitung kuota, hitung SKS, INSERT — semua dalam 1 transaction.
	EnrollAtomic(ctx context.Context, studentID, courseID int, tahunAkademik string) (model.Enrollment, error)
	FindByID(ctx context.Context, id int) (model.Enrollment, error)
	Delete(ctx context.Context, id, studentID int) error
	ListByStudent(ctx context.Context, studentID int) ([]model.EnrollmentWithCourse, error)
	TotalSKSTa(ctx context.Context, studentID int, tahunAkademik string) (int, error)
}

type enrollRepo struct{ pool *pgxpool.Pool }

func NewEnrollmentRepository(p *pgxpool.Pool) EnrollmentRepository { return &enrollRepo{p} }

func (r *enrollRepo) EnrollAtomic(ctx context.Context, studentID, courseID int, tahunAkademik string) (model.Enrollment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Enrollment{}, err
	}
	defer tx.Rollback(ctx)

	// 1) Cek duplikat
	var exist int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM enrollments
		WHERE student_id=$1 AND course_id=$2 AND tahun_akademik=$3`,
		studentID, courseID, tahunAkademik).Scan(&exist); err != nil {
		return model.Enrollment{}, err
	}
	if exist > 0 {
		return model.Enrollment{}, ErrAlreadyEnrolled
	}

	// 2) Lock course row + hitung kuota
	var c model.Course
	var terisi int
	err = tx.QueryRow(ctx, `SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, c.created_at, c.updated_at,
		(SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id) AS terisi
		FROM courses c WHERE c.id = $1 FOR UPDATE`, courseID).
		Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt, &terisi)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Enrollment{}, ErrNotFound
	}
	if err != nil {
		return model.Enrollment{}, err
	}
	if terisi >= c.Kuota {
		return model.Enrollment{}, ErrKuotaPenuh
	}

	// 3) Ambil IPK & total SKS yang sudah diambil pada TA yang sama
	var ipk *float64
	if err := tx.QueryRow(ctx, `SELECT ipk_terakhir FROM students
		WHERE id = $1 AND deleted_at IS NULL`, studentID).Scan(&ipk); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, err
	}

	var totalSKS int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(c2.sks), 0)
		FROM enrollments e JOIN courses c2 ON c2.id = e.course_id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2`, studentID, tahunAkademik).Scan(&totalSKS); err != nil {
		return model.Enrollment{}, err
	}

	ipkVal := 0.0
	if ipk != nil {
		ipkVal = *ipk
	}
	batas := BatasSKS(ipkVal)
	if totalSKS+c.SKS > batas {
		return model.Enrollment{}, ErrSKSExceed
	}

	// 4) INSERT
	var en model.Enrollment
	err = tx.QueryRow(ctx, `INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		VALUES ($1, $2, $3) RETURNING id, student_id, course_id, tahun_akademik, created_at`,
		studentID, courseID, tahunAkademik).
		Scan(&en.ID, &en.StudentID, &en.CourseID, &en.TahunAkademik, &en.CreatedAt)
	if err != nil {
		if IsUniqueViolation(err) {
			return model.Enrollment{}, ErrAlreadyEnrolled
		}
		return model.Enrollment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Enrollment{}, err
	}
	return en, nil
}

func (r *enrollRepo) FindByID(ctx context.Context, id int) (model.Enrollment, error) {
	var e model.Enrollment
	err := r.pool.QueryRow(ctx, `SELECT id, student_id, course_id, tahun_akademik, created_at
		FROM enrollments WHERE id = $1`, id).
		Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Enrollment{}, ErrNotFound
	}
	return e, err
}

func (r *enrollRepo) Delete(ctx context.Context, id, studentID int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1 AND student_id = $2`, id, studentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *enrollRepo) ListByStudent(ctx context.Context, studentID int) ([]model.EnrollmentWithCourse, error) {
	rows, err := r.pool.Query(ctx, `SELECT e.id, e.student_id, e.course_id, e.tahun_akademik, e.created_at,
		c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, c.created_at, c.updated_at
		FROM enrollments e JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1 ORDER BY e.tahun_akademik DESC, c.kode_mk`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.EnrollmentWithCourse{}
	for rows.Next() {
		var ec model.EnrollmentWithCourse
		if err := rows.Scan(
			&ec.Enrollment.ID, &ec.Enrollment.StudentID, &ec.Enrollment.CourseID,
			&ec.Enrollment.TahunAkademik, &ec.Enrollment.CreatedAt,
			&ec.Course.ID, &ec.Course.KodeMK, &ec.Course.NamaMK, &ec.Course.SKS,
			&ec.Course.Semester, &ec.Course.Kuota, &ec.Course.CreatedAt, &ec.Course.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, ec)
	}
	return out, nil
}

func (r *enrollRepo) TotalSKSTa(ctx context.Context, studentID int, tahunAkademik string) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(c.sks), 0)
		FROM enrollments e JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2`, studentID, tahunAkademik).Scan(&total)
	return total, err
}
```

### ✅ Commit Step 9

```bash
git add repository/
git commit -m "feat(repository): user/student/course/enrollment + row locking"
```

---

## 10. Service & Business Rules (`service/`)

### 10.1. `service/login_rate_limiter.go`

```go
package service

import (
	"sync"
	"time"
)

// LoginRateLimiter: max N gagal dalam window, lalu kunci selama M menit.
type LoginRateLimiter struct {
	maxFails   int
	lockoutDur time.Duration

	mu     sync.Mutex
	fails  map[string]*failRecord
}

type failRecord struct {
	count    int
	lockUntil time.Time
}

func NewLoginRateLimiter(maxFails int, lockout time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		maxFails:   maxFails,
		lockoutDur: lockout,
		fails:      make(map[string]*failRecord),
	}
}

func (l *LoginRateLimiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.fails[key]
	if !ok {
		return true, 0
	}
	if time.Now().Before(r.lockUntil) {
		return false, time.Until(r.lockUntil)
	}
	if r.lockUntil.IsZero() == false && time.Now().After(r.lockUntil) {
		// lockout selesai, reset
		delete(l.fails, key)
	}
	return true, 0
}

func (l *LoginRateLimiter) RegisterFailure(key string) (int, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.fails[key]
	if !ok {
		r = &failRecord{}
		l.fails[key] = r
	}
	r.count++
	if r.count >= l.maxFails {
		r.lockUntil = time.Now().Add(l.lockoutDur)
		return r.count, l.lockoutDur
	}
	return r.count, 0
}

func (l *LoginRateLimiter) RegisterSuccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, key)
}

// Start background cleanup (opsional, panggil dengan ticker)
func (l *LoginRateLimiter) Start(interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			l.mu.Lock()
			now := time.Now()
			for k, r := range l.fails {
				if !r.lockUntil.IsZero() && now.After(r.lockUntil) {
					delete(l.fails, k)
				}
			}
			l.mu.Unlock()
		}
	}()
}
```

### 10.2. `service/validation.go` (semua validasi input UTS)

```go
package service

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"api-siakad/app/model"
)

var (
	emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	nimRe   = regexp.MustCompile(`^\d{12}$`)
	taRe    = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)
)

func ValidateLogin(r model.LoginRequest) map[string]string {
	errs := map[string]string{}
	if !emailRe.MatchString(r.Email) {
		errs["email"] = "Format email tidak valid"
	}
	if len(r.Password) < 8 {
		errs["password"] = "Password minimal 8 karakter"
	}
	return errs
}

func ValidateCreateStudent(r model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}
	if !nimRe.MatchString(r.NIM) {
		errs["nim"] = "NIM harus 12 digit angka"
	}
	if strings.TrimSpace(r.Nama) == "" {
		errs["nama"] = "wajib diisi"
	}
	if !emailRe.MatchString(r.Email) {
		errs["email"] = "Format email tidak valid"
	}
	if strings.TrimSpace(r.Prodi) == "" {
		errs["prodi"] = "wajib diisi"
	}
	tahunNow := time.Now().Year()
	if r.Angkatan < 2000 || r.Angkatan > tahunNow {
		errs["angkatan"] = "Angkatan harus 4 digit dan <= tahun berjalan (" + strconv.Itoa(tahunNow) + ")"
	}
	if r.IPKTerakhir < 0 || r.IPKTerakhir > 4.0 {
		errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
	}
	return errs
}

func ValidateUpdateStudent(r model.UpdateStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(r.Nama) == "" {
		errs["nama"] = "wajib diisi"
	}
	if strings.TrimSpace(r.Prodi) == "" {
		errs["prodi"] = "wajib diisi"
	}
	tahunNow := time.Now().Year()
	if r.Angkatan < 2000 || r.Angkatan > tahunNow {
		errs["angkatan"] = "Angkatan harus 4 digit dan <= tahun berjalan (" + strconv.Itoa(tahunNow) + ")"
	}
	if r.IPKTerakhir < 0 || r.IPKTerakhir > 4.0 {
		errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
	}
	return errs
}

func ValidateCreateEnrollment(r model.CreateEnrollmentRequest) map[string]string {
	errs := map[string]string{}
	if r.CourseID <= 0 {
		errs["course_id"] = "wajib diisi dan harus > 0"
	}
	if !taRe.MatchString(r.TahunAkademik) {
		errs["tahun_akademik"] = `Format harus "YYYY/YYYY-Ganjil" atau "YYYY/YYYY-Genap"`
	}
	return errs
}
```

### 10.3. `service/auth_service.go` (Login + Me)

```go
package service

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"api-siakad/app/middleware"
	"api-siakad/app/model"
	"api-siakad/app/repository"
	"api-siakad/helper"
)

type AuthService struct {
	users    repository.UserRepository
	students repository.StudentRepository
	tokens   *helper.TokenIssuer
	limiter  *LoginRateLimiter
}

func NewAuthService(
	users repository.UserRepository,
	students repository.StudentRepository,
	tokens *helper.TokenIssuer,
	limiter *LoginRateLimiter,
) *AuthService {
	return &AuthService{users: users, students: students, tokens: tokens, limiter: limiter}
}

// Login — endpoint #1
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus JSON valid")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if errs := ValidateLogin(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	key := c.IP() + "|" + req.Email
	if allowed, retry := s.limiter.Allow(key); !allowed {
		c.Set("Retry-After", strconvI(int(retry.Seconds())))
		return helper.Fail(c, fiber.StatusTooManyRequests, "Terlalu banyak percobaan login, coba lagi nanti")
	}

	u, err := s.users.FindByEmail(c.UserContext(), req.Email)
	if err != nil || !u.IsActive {
		s.limiter.RegisterFailure(key)
		return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		s.limiter.RegisterFailure(key)
		return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah")
	}
	s.limiter.RegisterSuccess(key)

	access, exp, err := s.tokens.Issue(u.ID, u.Email, u.Role)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menerbitkan token")
	}
	return helper.Ok(c, "Login berhasil", model.LoginResponse{
		AccessToken: access,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Until(exp).Seconds()),
		User:        model.UserResponse{ID: u.ID, Email: u.Email, Role: u.Role},
	})
}

// Me — endpoint #2
func (s *AuthService) Me(c *fiber.Ctx) error {
	userID, email, role, ok := middleware.GetUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "Belum terautentikasi")
	}

	u, err := s.users.FindByID(c.UserContext(), userID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "User tidak ditemukan")
	}

	resp := model.MeResponse{
		User: model.UserResponse{ID: u.ID, Email: u.Email, Role: u.Role},
	}
	_ = email
	_ = role

	if u.Role == model.RoleMahasiswa {
		if st, err := s.students.FindByUserID(c.UserContext(), u.ID); err == nil {
			resp.Student = &model.StudentResponse{
				ID: st.ID, NIM: st.NIM, Nama: st.Nama,
				Prodi: st.Prodi, Angkatan: st.Angkatan, IPKTerakhir: st.IPKTerakhir,
			}
		}
	}
	return helper.Ok(c, "Profil pengguna", resp)
}

func strconvI(n int) string {
	// helper kecil agar tidak import strconv di atas banyak tempat
	if n < 0 {
		n = 0
	}
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
```

### 10.4. `service/student_service.go`

```go
package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"api-siakad/app/middleware"
	"api-siakad/app/model"
	"api-siakad/app/repository"
	"api-siakad/helper"
)

type StudentService struct {
	students    repository.StudentRepository
	enrollments repository.EnrollmentRepository
}

func NewStudentService(s repository.StudentRepository, e repository.EnrollmentRepository) *StudentService {
	return &StudentService{students: s, enrollments: e}
}

func ctx5(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	return id, err == nil && id > 0
}

// List — endpoint #3 (admin only)
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := ctx5(c)
	defer cancel()

	page, perPage := helper.ParsePagination(c, 10, 50)
	q := model.ListStudentQuery{
		Page:    page,
		PerPage: perPage,
		Prodi:   strings.TrimSpace(c.Query("prodi")),
		Search:  strings.TrimSpace(c.Query("search")),
		Sort:    c.Query("sort", "nama"),
	}
	if angkatan, err := strconv.Atoi(c.Query("angkatan", "0")); err == nil {
		q.Angkatan = angkatan
	}

	data, total, err := s.students.List(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa")
	}
	return helper.OkWithMeta(c, "Data mahasiswa berhasil diambil", data, helper.BuildMeta(page, perPage, total))
}

// Create — endpoint #4 (admin only)
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := ctx5(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus JSON valid")
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Nama = strings.TrimSpace(req.Nama)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Prodi = strings.TrimSpace(req.Prodi)

	if errs := ValidateCreateStudent(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Password awal = NIM (sesuai spek: "password awal = nim yang di-hash")
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NIM), 10)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memproses password")
	}

	created, err := s.students.CreateWithUser(ctx, model.User{
		Email:        req.Email,
		PasswordHash: string(hash),
	}, model.Student{
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicate):
			return helper.Fail(c, fiber.StatusConflict, "NIM atau email sudah terdaftar")
		default:
			return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menyimpan mahasiswa")
		}
	}

	resp := model.StudentResponse{
		ID: created.ID, NIM: created.NIM, Nama: created.Nama,
		Prodi: created.Prodi, Angkatan: created.Angkatan, IPKTerakhir: created.IPKTerakhir,
	}
	return helper.Created(c, "Mahasiswa berhasil dibuat", resp,
		"/api/v1/students/"+strconv.Itoa(created.ID))
}

// Get — endpoint #5 (admin atau mahasiswa data sendiri)
func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := ctx5(c)
	defer cancel()

	userID, _, role, _ := middleware.GetUser(c)
	id, ok := paramID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus angka positif")
	}

	st, err := s.students.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan atau sudah dihapus")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil mahasiswa")
	}

	// Mahasiswa hanya boleh lihat data sendiri
	if role == model.RoleMahasiswa && st.UserID != userID {
		return helper.Fail(c, fiber.StatusForbidden, "Tidak boleh melihat data mahasiswa lain")
	}

	enrolls, _ := s.enrollments.ListByStudent(ctx, id)
	courses := make([]model.CourseResponse, 0, len(enrolls))
	totalSKS := 0
	for _, e := range enrolls {
		courses = append(courses, model.CourseResponse{
			ID: e.Course.ID, KodeMK: e.Course.KodeMK, NamaMK: e.Course.NamaMK,
			SKS: e.Course.SKS, Semester: e.Course.Semester, Kuota: e.Course.Kuota,
		})
		totalSKS += e.Course.SKS
	}
	batas := repository.BatasSKS(st.IPKTerakhir)
	return helper.Ok(c, "Detail mahasiswa", model.StudentDetailResponse{
		Student: model.StudentResponse{
			ID: st.ID, NIM: st.NIM, Nama: st.Nama,
			Prodi: st.Prodi, Angkatan: st.Angkatan, IPKTerakhir: st.IPKTerakhir,
		},
		Courses:  courses,
		TotalSKS: totalSKS,
		BatasSKS: batas,
		SisaSKS:  batas - totalSKS,
	})
}

// Replace — endpoint #6 (admin only)
func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := ctx5(c)
	defer cancel()

	id, ok := paramID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus angka positif")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus JSON valid")
	}
	req.Nama = strings.TrimSpace(req.Nama)
	req.Prodi = strings.TrimSpace(req.Prodi)
	if errs := ValidateUpdateStudent(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	updated, err := s.students.Update(ctx, model.Student{
		ID: id, Nama: req.Nama, Prodi: req.Prodi,
		Angkatan: req.Angkatan, IPKTerakhir: req.IPKTerakhir,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan atau sudah dihapus")
		default:
			return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memperbarui mahasiswa")
		}
	}

	return helper.Ok(c, "Mahasiswa berhasil diperbarui", model.StudentResponse{
		ID: updated.ID, NIM: updated.NIM, Nama: updated.Nama,
		Prodi: updated.Prodi, Angkatan: updated.Angkatan, IPKTerakhir: updated.IPKTerakhir,
	})
}

// Delete — endpoint #7 (admin only, soft delete)
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := ctx5(c)
	defer cancel()

	id, ok := paramID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus angka positif")
	}
	if err := s.students.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menghapus mahasiswa")
	}
	return helper.NoContent(c)
}
```

### 10.5. `service/course_service.go` (endpoint #8)

```go
package service

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-siakad/app/model"
	"api-siakad/app/repository"
	"api-siakad/helper"
)

type CourseService struct {
	courses repository.CourseRepository
}

func NewCourseService(c repository.CourseRepository) *CourseService {
	return &CourseService{courses: c}
}

func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	q := model.ListCourseQuery{
		Semester:  c.QueryInt("semester", 0),
		Search:    strings.TrimSpace(c.Query("search")),
		Available: c.Query("available") == "true",
	}

	list, err := s.courses.List(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data mata kuliah")
	}

	out := make([]model.CourseResponse, 0, len(list))
	for _, x := range list {
		if q.Available && x.Terisi >= x.Kuota {
			continue
		}
		out = append(out, model.CourseResponse{
			ID: x.ID, KodeMK: x.KodeMK, NamaMK: x.NamaMK, SKS: x.SKS,
			Semester: x.Semester, Kuota: x.Kuota,
			Terisi:    x.Terisi,
			SisaKuota: x.Kuota - x.Terisi,
		})
	}
	return helper.Ok(c, "Daftar mata kuliah", out)
}
```

### 10.6. `service/enrollment_service.go` (endpoint #9 & #10)

```go
package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-siakad/app/middleware"
	"api-siakad/app/model"
	"api-siakad/app/repository"
	"api-siakad/helper"
)

type EnrollmentService struct {
	students    repository.StudentRepository
	enrollments repository.EnrollmentRepository
}

func NewEnrollmentService(s repository.StudentRepository, e repository.EnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{students: s, enrollments: e}
}

// Create — endpoint #9 (mahasiswa only)
func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	userID, _, role, _ := middleware.GetUser(c)
	if role != model.RoleMahasiswa {
		return helper.Fail(c, fiber.StatusForbidden, "Hanya mahasiswa yang dapat mengambil mata kuliah")
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus JSON valid")
	}
	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)
	if errs := ValidateCreateEnrollment(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	st, err := s.students.FindByUserID(ctx, userID)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan")
	}

	en, err := s.enrollments.EnrollAtomic(ctx, st.ID, req.CourseID, req.TahunAkademik)
	switch {
	case errors.Is(err, repository.ErrAlreadyEnrolled):
		return helper.Fail(c, fiber.StatusConflict, "Mata kuliah sudah pernah diambil pada tahun akademik ini")
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "Mata kuliah tidak ditemukan")
	case errors.Is(err, repository.ErrKuotaPenuh):
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Kuota mata kuliah sudah penuh")
	case errors.Is(err, repository.ErrSKSExceed):
		total, _ := s.enrollments.TotalSKSTa(ctx, st.ID, req.TahunAkademik)
		batas := repository.BatasSKS(st.IPKTerakhir)
		return helper.Fail(c, fiber.StatusUnprocessableEntity,
			"Total SKS melebihi batas. Sisa SKS Anda: "+strconv.Itoa(batas-total))
	case err != nil:
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil mata kuliah")
	}
	return helper.Created(c, "Berhasil mengambil mata kuliah", en,
		"/api/v1/enrollments/"+strconv.Itoa(en.ID))
}

// Delete — endpoint #10 (mahasiswa, hanya miliknya)
func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()

	userID, _, role, _ := middleware.GetUser(c)
	if role != model.RoleMahasiswa {
		return helper.Fail(c, fiber.StatusForbidden, "Hanya mahasiswa yang dapat membatalkan KRS")
	}

	id, ok := paramID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus angka positif")
	}

	en, err := s.enrollments.FindByID(ctx, id)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "Enrollment tidak ditemukan")
	}

	st, err := s.students.FindByUserID(ctx, userID)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan")
	}
	if en.StudentID != st.ID {
		return helper.Fail(c, fiber.StatusForbidden, "Bukan enrollment milik Anda")
	}

	if err := s.enrollments.Delete(ctx, id, st.ID); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal membatalkan KRS")
	}
	return helper.NoContent(c)
}
```

### ✅ Commit Step 10

```bash
git add service/
git commit -m "feat(service): auth/student/course/enrollment + business rules"
```

---

## 11. Routes (`route/`)

### 11.1. `route/route.go`

```go
package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-siakad/app/middleware"
	"api-siakad/app/model"
	"api-siakad/app/repository"
	"api-siakad/app/service"
	"api-siakad/config"
	"api-siakad/helper"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.TokenIssuer
	AuthSvc     *service.AuthService
	StudentSvc  *service.StudentService
	CourseSvc   *service.CourseService
	EnrollSvc   *service.EnrollmentService
}

func Register(app *fiber.App, deps Dependencies) {
	app.Use(requestid.New())
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("SIAKAD Mini API — running")
	})

	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(deps.Pool))

	// 1) Login (publik)
	api.Post("/auth/login", deps.AuthSvc.Login)

	// Semua endpoint di bawah butuh auth
	auth := api.Group("", middleware.RequireAuth(deps.JWT))

	// 2) Me
	auth.Get("/auth/me", deps.AuthSvc.Me)

	// 3) GET /students (admin)
	// 4) POST /students (admin)
	// 5) GET /students/:id (admin atau mahasiswa data sendiri — dicek di service)
	// 6) PUT /students/:id (admin)
	// 7) DELETE /students/:id (admin)
	students := auth.Group("/students")
	students.Get("/",
		middleware.RequireRole(model.RoleAdmin),
		deps.StudentSvc.List)
	students.Post("/",
		middleware.RequireRole(model.RoleAdmin),
		deps.StudentSvc.Create)
	students.Get("/:id", deps.StudentSvc.Get)
	students.Put("/:id",
		middleware.RequireRole(model.RoleAdmin),
		deps.StudentSvc.Replace)
	students.Delete("/:id",
		middleware.RequireRole(model.RoleAdmin),
		deps.StudentSvc.Delete)

	// 8) GET /courses (semua role yang login)
	auth.Get("/courses", deps.CourseSvc.List)

	// 9) POST /enrollments (mahasiswa)
	// 10) DELETE /enrollments/:id (mahasiswa, miliknya)
	enrolls := auth.Group("/enrollments",
		middleware.RequireRole(model.RoleMahasiswa))
	enrolls.Post("/", deps.EnrollSvc.Create)
	enrolls.Delete("/:id", deps.EnrollSvc.Delete)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "Endpoint tidak ditemukan")
	})
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "Database tidak dapat dihubungi")
		}
		return helper.Ok(c, "Server dan database berjalan", nil)
	}
}

// init — agar kompatibilitas bila package ini di-import terpisah (opsional)
var _ = config.GetEnv // suppress unused import
```

### ✅ Commit Step 11

```bash
git add route/
git commit -m "feat(route): daftarkan 10 endpoint SIAKAD"
```

---

## 12. Main.go (Composition Root)

### 12.1. `main.go`

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-siakad/app/repository"
	"api-siakad/app/service"
	"api-siakad/config"
	"api-siakad/database"
	"api-siakad/helper"
	"api-siakad/route"
)

func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	app := fiber.New(fiber.Config{
		AppName: "SIAKAD Mini API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			pesan := "Terjadi kesalahan pada server"
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				pesan = e.Message
			}
			return c.Status(status).JSON(fiber.Map{
				"success": false,
				"message": pesan,
			})
		},
	})

	// JWT
	secret := config.GetEnv("JWT_SECRET", "")
	if secret == "" {
		log.Fatal("JWT_SECRET wajib diisi di .env")
	}
	accessTTL := time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60)) * time.Minute
	tokens := helper.NewTokenIssuer(secret, accessTTL)

	// Rate limiter login
	maxFails := config.GetEnvInt("LOGIN_MAX_FAILS", 5)
	lockout := time.Duration(config.GetEnvInt("LOGIN_LOCKOUT_MINUTES", 1)) * time.Minute
	limiter := service.NewLoginRateLimiter(maxFails, lockout)
	limiter.Start(5 * time.Minute)
	defer limiter.Stop()

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollRepo := repository.NewEnrollmentRepository(pool)

	// Services
	authSvc := service.NewAuthService(userRepo, studentRepo, tokens, limiter)
	studentSvc := service.NewStudentService(studentRepo, enrollRepo)
	courseSvc := service.NewCourseService(courseRepo)
	enrollSvc := service.NewEnrollmentService(studentRepo, enrollRepo)

	// Routes
	route.Register(app, route.Dependencies{
		Pool:       pool,
		JWT:        tokens,
		AuthSvc:    authSvc,
		StudentSvc: studentSvc,
		CourseSvc:  courseSvc,
		EnrollSvc:  enrollSvc,
	})

	addr := ":" + config.GetEnv("APP_PORT", "3000")
	log.Printf("Server mendengarkan %s", addr)
	log.Fatal(app.Listen(addr))
}
```

### 12.2. Build & Run

```powershell
cd "C:/Users/LENOVO/OneDrive/semester 5/Prak/Backend Prak/uts"
go mod tidy
go build ./...
go run .
```

Output sukses:
```
Server mendengarkan :3000
```

### ✅ Commit Step 12

```bash
git add main.go
git commit -m "feat: composition root, hubungkan semua service & repo"
```

---

## 13. Testing Manual dengan curl/Postman

### 13.1. Login admin

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@siakad.test","password":"admin123"}'
```

Response:
```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": { "id": 1, "email": "admin@siakad.test", "role": "admin" }
  }
}
```

Simpan token:
```powershell
# PowerShell
$response = Invoke-RestMethod -Method Post -Uri http://localhost:3000/api/v1/auth/login `
  -ContentType "application/json" `
  -Body '{"email":"admin@siakad.test","password":"admin123"}'
$TOKEN = $response.data.access_token

# Test endpoint
$headers = @{ Authorization = "Bearer $TOKEN" }
Invoke-RestMethod -Uri "http://localhost:3000/api/v1/students?per_page=5" -Headers $headers
```

### 13.2. Endpoint penting yang harus di-test

| # | Aksi | curl |
|---|------|------|
| 1 | Login admin | `POST /api/v1/auth/login` |
| 2 | Get profile | `GET /api/v1/auth/me` (Bearer) |
| 3 | List students (admin) | `GET /api/v1/students?page=1&per_page=5&sort=-ipk_terakhir` |
| 4 | Create student | `POST /api/v1/students` body `{nim,nama,email,prodi,angkatan,ipk_terakhir}` |
| 5 | Get student detail | `GET /api/v1/students/1` |
| 6 | Update student | `PUT /api/v1/students/1` body `{nama,prodi,angkatan,ipk_terakhir}` |
| 7 | Soft delete | `DELETE /api/v1/students/1` (no body) → 204 |
| 8 | List courses | `GET /api/v1/courses?semester=1&available=true` |
| 9 | Enroll course (mhs) | `POST /api/v1/enrollments` body `{course_id, tahun_akademik}` |
| 10 | Cancel enrollment (mhs) | `DELETE /api/v1/enrollments/1` |

### 13.3. Skenario uji penting (WAJIB)

| Skenario | Ekspektasi |
|----------|------------|
| Login 5x salah | 6x percobaan → 429 Too Many Requests |
| Login tanpa email / format salah | 422 dengan error per field |
| Mahasiswa coba GET `/students` (bukan /me) | 403 |
| Mahasiswa A GET `/students/<id B>` | 403 |
| GET `/students/<id dihapus>` | 404 |
| POST `/enrollments` MK yang sama 2x di TA sama | 409 |
| POST `/enrollments` MK yang kuotanya penuh | 422 |
| POST `/enrollments` hingga SKS > batas (IPK 2.0 → max 18) | 422 dengan pesan "Sisa SKS Anda: 0" |
| DELETE `/enrollments/<id orang lain>` sebagai mahasiswa | 403 |
| PUT `/students/<id>` field NIM di body | NIM tidak berubah (NIM immutable) |

### 13.4. Ekspor Postman Collection

Setelah semua endpoint di-test via Postman, **Export** collection ke `postman/uts-siakad.postman_collection.json` lalu commit.

### ✅ Commit Step 13

```bash
git add postman/ README.md
git commit -m "test: postman collection + skenario uji lengkap"
```

---

## 14. Commit Bertahap ke GitHub

Riwayat commit yang ideal (sekitar 12 commit):

```
ab08ca1 setup migration
<new>  feat(db): schema SIAKAD + seeder 1 admin, 20 mhs, 10 mk
<new>  chore: init go module + dependencies + .env.example
<new>  feat(infra): env loader + pgxpool postgres
<new>  feat(helper): response envelope, pagination, JWT issuer
<new>  feat(middleware): require auth, role guard, json guard
<new>  feat(model): DTO untuk user, student, course, enrollment
<new>  feat(repository): user/student/course/enrollment + row locking
<new>  feat(service): auth/student/course/enrollment + business rules
<new>  feat(route): daftarkan 10 endpoint SIAKAD
<new>  feat: composition root, hubungkan semua service & repo
<new>  test: postman collection + skenario uji lengkap
<new>  docs: README + laporan UTS
```

Push:

```bash
git remote add origin https://github.com/ItsukaShior1/UTS_PBL_PRAK_043.git   # kalau belum
git branch -M main
git push -u origin main
```

> ⚠️ Repo `https://github.com/ItsukaShior1/UTS_PBL_PRAK_043.git` sudah ada di remote. Jadi cukup `git push` setelah langkah-langkah di atas.

---

## 15. Checklist Kepatuhan Modul

| Syarat | Status | Keterangan |
|---|---|---|
| Go + Fiber + PostgreSQL | ✅ | Step 4 |
| Migration + seeder (1 admin, 20 mhs, 10 mk) | ✅ | Step 3 |
| Password di-hash (bcrypt) | ✅ | `bcrypt.GenerateFromPassword`, `auth_service.Login` |
| Token di semua endpoint kecuali login | ✅ | `RequireAuth` di group `/api/v1` setelah login |
| Riwayat commit bertahap di GitHub | ✅ | Step 14 |
| Batas SKS by IPK (24/21/18) | ✅ | `repository.BatasSKS()` |
| Anti duplikat KRS | ✅ | UNIQUE constraint + `ErrAlreadyEnrolled` |
| Kuota penuh ditolak | ✅ | `ErrKuotaPenuh` setelah `FOR UPDATE` |
| Akses KRS hanya milik sendiri | ✅ | `enrollment_service.Delete` cek `StudentID` |
| 10 endpoint sesuai spek | ✅ | Step 11 |
| Format response uniform (success/error envelope) | ✅ | `helper.Ok` / `helper.Fail` |
| Status code 200/201/204/401/403/404/409/422/429/500 | ✅ | Dipakai sesuai konteks |
| Soft delete student | ✅ | `deleted_at` di filter `List` dan dicek di `Get`/`Update` |
| Row locking untuk enrollment | ✅ | `FOR UPDATE` di `EnrollAtomic` |
| Batas SKS by IPK logic | ✅ | `BatasSKS()` dengan switch 3 level |

---

## 🆘 Troubleshooting

| Masalah | Solusi |
|---------|--------|
| `dial tcp ...:5432: connectex: No connection could be made` | Postgres belum jalan atau port salah. Cek `DB_HOST`, `DB_PORT` di `.env`. |
| `pq: password authentication failed` | `DB_USER` / `DB_PASSWORD` salah di `.env`. |
| `bcrypt: ...` error saat generate hash | Pastikan `go get golang.org/x/crypto/bcrypt` sudah dijalankan. |
| Seeder jalan tapi hash masih `<BCRYPT_HASH_admin123>` | Replace placeholder di `005_seed_data.sql` dulu. |
| `pq: duplicate key value violates unique constraint` saat POST student | NIM/email sudah ada, pakai yang lain. |
| Token selalu 401 | Cek header: `Authorization: Bearer <token>` (dengan spasi). |
| Login rate limit aktif terus | Tunggu `LOGIN_LOCKOUT_MINUTES` atau restart server (memory reset). |
| Kuota penuh padahal baru 1 enroll | Cek `courses.kuota` di DB, dan ada data enrollments sebelumnya. |

---

## 🎯 Ringkasan Urutan Eksekusi

1. **Step 3** → buat DB + migration + seeder → commit
2. **Step 4** → `go mod init` + deps + `.env` → commit
3. **Step 5** → `config/` + `database/` → commit
4. **Step 6** → `helper/` → commit
5. **Step 7** → `middleware/` → commit
6. **Step 8** → `model/` → commit
7. **Step 9** → `repository/` → commit
8. **Step 10** → `service/` → commit
9. **Step 11** → `route/` → commit
10. **Step 12** → `main.go` + `go run .` → commit
11. **Step 13** → testing Postman + export collection → commit
12. **Step 14** → push + tag

Selamat mengerjakan! 🚀
