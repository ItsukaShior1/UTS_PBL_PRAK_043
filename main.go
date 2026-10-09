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

	secret := config.GetEnv("JWT_SECRET", "")
	if secret == "" {
		log.Println("peringatan: JWT_SECRET kosong, server tetap jalan tapi token akan invalid untuk lintas restart")
		secret = "dev-only-insecure-secret-please-replace-in-production-env-32"
	}
	accessTTL := time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60)) * time.Minute
	tokens := helper.NewTokenIssuer(secret, accessTTL)

	maxFails := config.GetEnvInt("LOGIN_MAX_FAILS", 5)
	lockout := time.Duration(config.GetEnvInt("LOGIN_LOCKOUT_MINUTES", 1)) * time.Minute
	limiter := service.NewLoginRateLimiter(maxFails, lockout)
	limiter.Start(5 * time.Minute)
	defer limiter.Stop()

	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollRepo := repository.NewEnrollmentRepository(pool)

	authSvc := service.NewAuthService(userRepo, studentRepo, tokens, limiter)
	studentSvc := service.NewStudentService(studentRepo, enrollRepo)
	courseSvc := service.NewCourseService(courseRepo)
	enrollSvc := service.NewEnrollmentService(studentRepo, enrollRepo)

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
