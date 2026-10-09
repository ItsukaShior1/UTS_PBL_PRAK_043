package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-siakad/app/model"
)

type EnrollmentRepository interface {
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

	var exist int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM enrollments
		WHERE student_id=$1 AND course_id=$2 AND tahun_akademik=$3`,
		studentID, courseID, tahunAkademik).Scan(&exist); err != nil {
		return model.Enrollment{}, err
	}
	if exist > 0 {
		return model.Enrollment{}, ErrAlreadyEnrolled
	}

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
