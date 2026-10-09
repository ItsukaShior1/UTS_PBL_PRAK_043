package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-siakad/middleware"
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
