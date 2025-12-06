package handler

import (
	"net/http"
	"strconv"

	"xiazki/internal/model"
	"xiazki/web/template/add_book"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetAddBook(c echo.Context) error {
	return Render(c, add_book.Show(add_book.Data{Op: add_book.Add}))
}

func (h *Handler) PostAddBook(c echo.Context) error {
	var bfv add_book.BookFormValues
	if err := c.Bind(&bfv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	if errors := bfv.Validate(); len(errors) > 0 {
		return Render(c, add_book.FormAdd(add_book.Data{
			Op:     add_book.Add,
			Values: bfv,
			Errors: errors,
		}))
	}

	if err := h.db.InsertBook(c, bfv.ToBook()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add book: "+err.Error())
	}

	return HxRedirect(c, "/books")
}

func (h *Handler) GetBookEdit(c echo.Context) error {
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

	return Render(c, add_book.Show(add_book.Data{
		Op:     add_book.Edit,
		Values: add_book.BookToBookFormValues(b),
		BookID: b.ID,
	},
	))
}

func (h *Handler) PutBookEdit(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	var bfv add_book.BookFormValues
	if err := c.Bind(&bfv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	if errors := bfv.Validate(); len(errors) > 0 {
		return Render(c, add_book.FormEdit(add_book.Data{
			Op:     add_book.Edit,
			BookID: id,
			Errors: errors,
			Values: bfv,
		},
		))
	}

	if err := h.db.UpdateBook(c, id, bfv.ToBook()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update book: "+err.Error())
	}

	return HxRedirect(c, "/book/"+idStr)
}
