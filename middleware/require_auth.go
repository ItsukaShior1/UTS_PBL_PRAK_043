package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-siakad/helper"
)

const (
	LocalsUserID = "user_id"
	LocalsEmail  = "email"
	LocalsRole   = "role"
)

func RequireAuth(issuer *helper.TokenIssuer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			c.Set("WWW-Authenticate", `Bearer realm="siakad"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "Token wajib dikirim")
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			c.Set("WWW-Authenticate", `Bearer realm="siakad"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid")
		}
		claims, err := issuer.Parse(parts[1])
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="siakad"`)
			if errors.Is(err, helper.ErrTokenExpired) {
				return helper.Fail(c, fiber.StatusUnauthorized, "Token kedaluwarsa")
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid")
		}
		c.Locals(LocalsUserID, claims.UserID)
		c.Locals(LocalsEmail, claims.Email)
		c.Locals(LocalsRole, claims.Role)
		return c.Next()
	}
}
