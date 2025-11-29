package handler

import (
	"database/sql"
	"errors"
	"log"
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
			return q.Order("event.date ASC")
		}).
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
	} else if errors.Is(err, sql.ErrNoRows) {
		return Render(c, book.Stats(id, model.ReviewStats{}))
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book stats")
}

func (h *Handler) PostBookRate(c echo.Context) error {
	ratingStr := c.FormValue("rating")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 0 || rating > 10 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid rating value")
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event ID")
	}

	err = h.db.NewSelect().
		Model(&model.Book{}).
		Where("id = ?", id).
		Limit(1).
		Scan(c.Request().Context())
	if err != nil {
		log.Println("Error fetching book:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book: "+err.Error())
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
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit rating: "+err.Error())
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
	} else if errors.Is(err, sql.ErrNoRows) {
		return Render(c, book.Stats(id, model.ReviewStats{}))
	}
	log.Println("Error fetching book stats:", err)
	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book stats")
}
