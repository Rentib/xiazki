package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"xiazki/internal/model"
	"xiazki/web/template/book"
	"xiazki/web/template/opinions"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

func (h *Handler) GetBookOpinions(c echo.Context) error {
	idStr := c.Param("id")
	bookID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	var b model.Book
	err = h.db.NewSelect().
		Model(&b).
		Where("id = ?", bookID).
		Column("id", "title", "cover_url").
		Relation("Authors").
		Limit(1).
		Scan(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to fetch book details: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book details")
	}

	user, err := h.currentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	ur := model.Review{BookID: b.ID} // set default to make it possible to add new reviews
	err = h.db.NewSelect().
		Model(&ur).
		Where("book_id = ? AND user_id = ?", bookID, user.ID).
		Limit(1).
		Scan(c.Request().Context())
	if err != nil && err != sql.ErrNoRows {
		c.Logger().Error("Failed to fetch user review: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch user review")
	}

	var or []*model.Review
	err = h.db.NewSelect().
		Model(&or).
		Where("book_id = ? AND user_id != ?", bookID, user.ID).
		Relation("User", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("id", "username")
		}).
		Order("review.created_at DESC").
		Scan(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to fetch reviews: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch reviews")
	}

	stats, err := h.db.FullReviewStats(c.Request().Context(), bookID, user.ID)
	if err != nil {
		c.Logger().Error("Failed to fetch book stats: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book stats")
	}

	return Render(c, opinions.Show(opinions.Data{
		Book:         &b,
		UserReview:   &ur,
		OtherReviews: or,
		Stats:        stats,
	}))
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
	if err != nil {
		c.Logger().Error("Failed to fetch book stats: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book stats")
	}
	return Render(c, book.Stats(id, stats))
}

func (h *Handler) PostBookReview(c echo.Context) error {
	ratingStr := c.FormValue("rating")
	rating, err := strconv.ParseInt(ratingStr, 10, 64)
	if err != nil || rating < 0 || rating > 10 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid rating value")
	}

	opinion := c.FormValue("opinion")

	idStr := c.Param("id")
	bookID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	user, err := h.currentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	if err := h.db.InsertOrUpdateReview(c.Request().Context(), &model.Review{
		UserID:  user.ID,
		BookID:  bookID,
		Rating:  rating,
		Opinion: opinion,
	}, true); err != nil {
		c.Logger().Error("Failed to submit review: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit review")
	}

	return HxRedirect(c, "/book/"+idStr+"/opinions")
}
