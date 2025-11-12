package handler

import (
	"net/http"
	"strconv"
	"xiazki/model"
	"xiazki/view/book"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetBook(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	var b model.Book
	err = h.db.NewSelect().
		Model(&b).
		Where("id = ?", id).
		Relation("Authors").
		Relation("Tags").
		Relation("Translators").
		Relation("Narrators").
		Scan(c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book details")
	}

	return Render(c, book.Show(
		book.Data{
			CSRF: c.Get("csrf").(string),
			Book: b,
		},
	))
}

func (h *Handler) DeleteBook(c echo.Context) error {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	_, err = h.db.NewDelete().
		Model((*model.Book)(nil)).
		Where("id = ?", id).
		Exec(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete book: "+err.Error())
	}

	return HxRedirect(c, "/books")
}
