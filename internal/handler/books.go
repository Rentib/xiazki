package handler

import (
	"net/http"
	"strconv"

	"xiazki/internal/model"
	"xiazki/web/template/author"
	"xiazki/web/template/books"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

func (h *Handler) GetBooks(c echo.Context) error {
	var b []*model.Book
	err := h.db.NewSelect().
		Model(&b).
		Relation("Authors").
		Relation("Events", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.OrderExpr("CASE WHEN type = ? THEN 3 WHEN type = ? THEN 2 WHEN type = ? THEN 1 ELSE 0 END ASC", model.EventFinished, model.EventDropped, model.EventReading).OrderExpr("date ASC")
		}).
		OrderExpr("created_at DESC").
		Scan(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to fetch books: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch books")
	}

	csrf, ok := c.Get("csrf").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "CSRF token not found")
	}
	return Render(c, books.Show(books.Data{CSRF: csrf, Books: b}))
}

func (h *Handler) GetAuthor(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return err
	}

	// TOOD: change logic to add events
	// Author doesn't have events relation, so the sql query would need to be
	// different and start with books rather then authors.
	var a model.Author
	err = h.db.NewSelect().
		Model(&a).
		Where("id = ?", id).
		Relation("Books").
		Scan(c.Request().Context())
	if err != nil {
		return err
	}

	return Render(c, author.Show(author.Data{
		CSRF:   c.Get("csrf").(string),
		Author: &a,
	}))
}
