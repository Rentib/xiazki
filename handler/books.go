package handler

import (
	"xiazki/model"
	"xiazki/view/books"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetBooks(c echo.Context) error {
	var b []model.Book
	err := h.db.NewSelect().
		Model(&b).
		Relation("Authors").
		// Relation("Tags").
		// Relation("Translators").
		// Relation("Narrators").
		Order("created_at DESC").
		Scan(c.Request().Context())
	if err != nil {
		return err
	}

	return Render(c, books.Show(books.Data{CSRF: c.Get("csrf").(string), Books: b}))
}
