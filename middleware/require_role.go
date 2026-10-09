package middleware

import (
	"github.com/gofiber/fiber/v2"
	"api-siakad/helper"
)

func GetUser(c *fiber.Ctx) (userID int, email, role string, ok bool) {
	uid, ok1 := c.Locals(LocalsUserID).(int)
	e, ok2 := c.Locals(LocalsEmail).(string)
	r, ok3 := c.Locals(LocalsRole).(string)
	if !ok1 || !ok2 || !ok3 {
		return 0, "", "", false
	}
	return uid, e, r, true
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, _, role, ok := GetUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "Belum terautentikasi")
		}
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return helper.Fail(c, fiber.StatusForbidden, "Role Anda tidak berhak mengakses endpoint ini")
	}
}