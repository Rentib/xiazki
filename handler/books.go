package handler

import (
	"strconv"

	"xiazki/model"
	"xiazki/view/author"
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

func (h *Handler) GetAuthor(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	var a model.Author
	err = h.db.NewSelect().
		Model(&a).
		Where("id = ?", id).
		Relation("Books").
		Order("publish_date DESC").
		Scan(c.Request().Context())
	if err != nil {
		return err
	}

	return Render(c, author.Show(author.Data{
		CSRF:   c.Get("csrf").(string),
		Author: a,
	}))
}
