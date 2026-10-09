package service

import (
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
		c.Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
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

func (s *AuthService) Me(c *fiber.Ctx) error {
	userID, _, _, ok := middleware.GetUser(c)
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
