package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"api-siakad/middleware"
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
