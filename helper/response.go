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
