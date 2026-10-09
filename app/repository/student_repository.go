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

func (r *studentRepo) CreateWithUser(ctx context.Context, u model.User, s model.Student) (model.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Student{}, err
	}
	defer tx.Rollback(ctx)

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

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students s"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

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
	sql := fmt.Sprintf(`SELECT %s FROM students s%s ORDER BY %s LIMIT $%d OFFSET $%d`,
		studentCols, where, orderBy, idx, idx+1)

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
