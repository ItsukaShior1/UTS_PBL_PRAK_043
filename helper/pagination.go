package helper

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func ParsePagination(c *fiber.Ctx, defaultPerPage, maxPerPage int) (page, perPage int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ = strconv.Atoi(c.Query("per_page", strconv.Itoa(defaultPerPage)))
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage
}

func BuildMeta(page, perPage, total int) *Meta {
	last := 0
	if perPage > 0 {
		last = (total + perPage - 1) / perPage
	}
	return &Meta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    last,
	}
}
