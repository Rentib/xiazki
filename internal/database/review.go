package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"xiazki/internal/model"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (db *DB) InsertOrUpdateReview(ctx context.Context, review *model.Review, updateOpinion bool) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var oldReview model.Review

		if err := tx.NewSelect().
			Model(&oldReview).
			Where("user_id = ? AND book_id = ?", review.UserID, review.BookID).
			Limit(1).
			Scan(ctx); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_, err := tx.NewInsert().
					Model(review).
					Exec(ctx)
				return err
			}
			return err
		}

		if 0 <= review.Rating && review.Rating <= 10 {
			oldReview.Rating = review.Rating
		}
		if updateOpinion {
			oldReview.Opinion = review.Opinion
		}

		oldReview.UpdatedAt = time.Now()

		_, err := tx.NewUpdate().
			Model(&oldReview).
			WherePK().
			Exec(ctx)

		return err
	})
}

func (db *DB) FullReviewStats(ctx context.Context, bookID int64, userID uuid.UUID) (*model.ReviewStats, error) {
	var stats model.ReviewStats
	ratingsSpreadCTE := db.NewSelect().
		Table("reviews").
		ColumnExpr("rating").
		ColumnExpr("COUNT(*) AS count").
		Where("book_id = ?", bookID).
		Where("rating BETWEEN 1 AND 10").
		Group("rating")
	err := db.NewSelect().
		ColumnExpr("COALESCE(MAX(CASE WHEN user_id = ? THEN rating END), 0) AS user_rating", userID).
		ColumnExpr("COALESCE(AVG(CASE WHEN rating != 0 THEN CAST(rating AS REAL) END), 0.0) AS average_rating").
		ColumnExpr("COUNT(CASE WHEN rating != 0 THEN 1 END) AS ratings_count").
		ColumnExpr("COUNT(CASE WHEN opinion != '' THEN 1 END) AS opinions_count").
		ColumnExpr(`
			COALESCE(
				(SELECT json_object(
					'1', COALESCE(SUM(CASE WHEN rating = 1 THEN count ELSE 0 END), 0),
					'2', COALESCE(SUM(CASE WHEN rating = 2 THEN count ELSE 0 END), 0),
					'3', COALESCE(SUM(CASE WHEN rating = 3 THEN count ELSE 0 END), 0),
					'4', COALESCE(SUM(CASE WHEN rating = 4 THEN count ELSE 0 END), 0),
					'5', COALESCE(SUM(CASE WHEN rating = 5 THEN count ELSE 0 END), 0),
					'6', COALESCE(SUM(CASE WHEN rating = 6 THEN count ELSE 0 END), 0),
					'7', COALESCE(SUM(CASE WHEN rating = 7 THEN count ELSE 0 END), 0),
					'8', COALESCE(SUM(CASE WHEN rating = 8 THEN count ELSE 0 END), 0),
					'9', COALESCE(SUM(CASE WHEN rating = 9 THEN count ELSE 0 END), 0),
					'10', COALESCE(SUM(CASE WHEN rating = 10 THEN count ELSE 0 END), 0)
				) FROM (?) AS t), 
				'{}'
			) AS ratings_spread`, ratingsSpreadCTE).
		Table("reviews").
		Where("book_id = ?", bookID).
		Scan(ctx, &stats)
	return &stats, err
}
