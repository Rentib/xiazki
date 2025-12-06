package handler

import (
	"net/http"
	"strconv"

	"xiazki/internal/model"
	"xiazki/web/template/book"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetBookAddEvent(c echo.Context) error {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	return Render(c, book.AddEventModal(id))
}

func (h *Handler) PostBookAddEvent(c echo.Context) error {
	var efv book.EventFormValues
	if err := c.Bind(&efv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	if err := efv.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	u, err := h.currentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get current user: "+err.Error())
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	var b model.Book
	err = h.db.NewSelect().
		Model(&b).
		Column("id").
		Where("id = ?", id).
		Scan(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book: "+err.Error())
	}

	// TODO: notify user if event insertion is rejected due to existing conflicting events
	if err := h.db.InsertEvent(c.Request().Context(), &b, u, efv.ToEvent()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add event: "+err.Error())
	}

	return HxRedirect(c, "/book/"+idStr)
}

func (h *Handler) DeleteEvent(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event ID")
	}

	_, err = h.db.NewDelete().
		Model((*model.Event)(nil)).
		Where("id = ?", id).
		Exec(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete event: "+err.Error())
	}

	return HxRedirect(c, c.Request().Referer())
}
