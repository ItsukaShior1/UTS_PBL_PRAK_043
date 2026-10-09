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
