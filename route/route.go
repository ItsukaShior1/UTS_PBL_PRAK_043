package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-siakad/middleware"
	"api-siakad/app/model"
	"api-siakad/app/service"
	"api-siakad/helper"
)

type Dependencies struct {
	Pool       *pgxpool.Pool
	JWT        *helper.TokenIssuer
	AuthSvc    *service.AuthService
	StudentSvc *service.StudentService
	CourseSvc  *service.CourseService
	EnrollSvc  *service.EnrollmentService
}

func Register(app *fiber.App, deps Dependencies) {
	app.Use(requestid.New())
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("SIAKAD Mini API — running")
	})

	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(deps.Pool))

	api.Post("/auth/login", deps.AuthSvc.Login)

	auth := api.Group("", middleware.RequireAuth(deps.JWT))

	auth.Get("/auth/me", deps.AuthSvc.Me)

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

	auth.Get("/courses", deps.CourseSvc.List)

	enrolls := auth.Group("/enrollments",
		middleware.RequireRole(model.RoleMahasiswa))
	enrolls.Post("/", deps.EnrollSvc.Create)
	enrolls.Delete("/:id", deps.EnrollSvc.Delete)

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
