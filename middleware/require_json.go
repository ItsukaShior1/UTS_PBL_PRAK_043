package middleware

import (
	"github.com/gofiber/fiber/v2"
	"api-siakad/helper"
)

func RequireJSON() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ct := c.Get("Content-Type")
		if ct != "" && len(ct) >= 16 && ct[:16] != "application/json" {
			if !startsWith(ct, "application/json") {
				return helper.Fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
			}
		}
		return c.Next()
	}
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}