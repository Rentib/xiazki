package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"xiazki/internal/model"

	"github.com/uptrace/bun"
)

func InsertOrUpdateReview(db *bun.DB, ctx context.Context, review *model.Review) error {
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
		if review.Opinion != "" {
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
