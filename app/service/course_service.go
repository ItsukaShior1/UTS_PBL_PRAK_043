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
