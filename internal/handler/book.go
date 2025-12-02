package handler

import (
	"net/http"
	"strconv"

	"xiazki/internal/db"
	"xiazki/internal/model"
	"xiazki/web/template/book"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
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
		Relation("Events", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.OrderExpr("CASE WHEN type = ? THEN 3 WHEN type = ? THEN 2 WHEN type = ? THEN 1 ELSE 0 END ASC", model.EventFinished, model.EventDropped, model.EventReading).OrderExpr("date ASC")
		}).
		Scan(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to fetch book details: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book details")
	}

	csrf, _ := c.Get("csrf").(string)
	return Render(c, book.Show(
		book.Data{
			CSRF: csrf,
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
		c.Logger().Error("Failed to delete book: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete book")
	}

	return HxRedirect(c, "/books")
}

func (h *Handler) GetBookStats(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	user, err := h.currentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var stats model.ReviewStats
	err = h.db.NewSelect().
		ColumnExpr("COALESCE(MAX(CASE WHEN user_id = ? THEN rating END), 0) AS user_rating", user.ID).
		ColumnExpr("COALESCE(AVG(CAST(rating AS REAL)), 0.0) AS average_rating").
		ColumnExpr("COUNT(CASE WHEN rating != 0 THEN 1 END) AS ratings_count").
		ColumnExpr("COUNT(CASE WHEN opinion != '' THEN 1 END) AS opinions_count").
		Table("reviews").
		Where("book_id = ?", id).
		Scan(c.Request().Context(), &stats)
	if err == nil {
		return Render(c, book.Stats(id, stats))
	}
	c.Logger().Error("Failed to fetch book stats: ", err)
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book stats")
}

func (h *Handler) PostBookRate(c echo.Context) error {
	ratingStr := c.FormValue("rating")
	rating, err := strconv.ParseInt(ratingStr, 10, 64)
	if err != nil || rating < 0 || rating > 10 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid rating value")
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	err = h.db.NewSelect().
		Model(&model.Book{}).
		Where("id = ?", id).
		Limit(1).
		Scan(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to fetch book: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book")
	}

	user, err := h.currentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	if err := db.InsertOrUpdateReview(h.db, c.Request().Context(), &model.Review{
		UserID: user.ID,
		BookID: id,
		Rating: rating,
	}); err != nil {
		c.Logger().Error("Failed to submit rating: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit rating")
	}

	var stats model.ReviewStats
	err = h.db.NewSelect().
		ColumnExpr("COALESCE(MAX(CASE WHEN user_id = ? THEN rating END), 0) AS user_rating", user.ID).
		ColumnExpr("COALESCE(AVG(CAST(rating AS REAL)), 0.0) AS average_rating").
		ColumnExpr("COUNT(CASE WHEN rating != 0 THEN 1 END) AS ratings_count").
		ColumnExpr("COUNT(CASE WHEN opinion != '' THEN 1 END) AS opinions_count").
		Table("reviews").
		Where("book_id = ?", id).
		Scan(c.Request().Context(), &stats)
	if err == nil {
		return Render(c, book.Stats(id, stats))
	}
	c.Logger().Error("Failed to fetch book stats: ", err)
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book stats")
}
