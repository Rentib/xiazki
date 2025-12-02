package db

import (
	"context"
	"database/sql"
	"errors"

	"xiazki/internal/model"

	"github.com/uptrace/bun"
)

func InsertEvent(db *bun.DB, ctx context.Context, book *model.Book, user *model.User, event *model.Event) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		event.UserID = user.ID
		event.BookID = book.ID

		switch event.Type {
		case model.EventFinished, model.EventDropped:
			// Reject inserting a "finished" or "dropped" event if there is a later "reading" event
			err := tx.NewSelect().
				Model((*model.Event)(nil)).
				Where("user_id = ? AND book_id = ? AND type = ? AND date > ?", event.UserID, event.BookID, model.EventReading, event.Date).
				Limit(1).
				Scan(ctx)

			if err == nil {
				return errors.New("cannot insert finished/dropped event: reading event exists with later date")
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}

			// Delete existing "finished" or "dropped" events for this book and user
			_, err = tx.NewDelete().
				Model((*model.Event)(nil)).
				Where("user_id = ? AND book_id = ? AND type IN (?,?)", event.UserID, event.BookID, model.EventFinished, model.EventDropped).
				Exec(ctx)
			if err != nil {
				return err
			}
		case model.EventReading:
			// Reject inserting a "reading" event if there is an earlier "finished" or "dropped" event
			err := tx.NewSelect().
				Model((*model.Event)(nil)).
				Where("user_id = ? AND book_id = ? AND type IN (?,?) AND date < ?", event.UserID, event.BookID, model.EventFinished, model.EventDropped, event.Date).
				Limit(1).
				Scan(ctx)

			if err == nil {
				return errors.New("cannot insert reading event: finished/dropped event exists with earlier date")
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}

			// Delete existing "reading" events for this book and user
			_, err = tx.NewDelete().
				Model((*model.Event)(nil)).
				Where("user_id = ? AND book_id = ? AND type = ?", event.UserID, event.BookID, model.EventReading).
				Exec(ctx)
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown event type")
		}

		// Insert the new event
		_, err := tx.NewInsert().Model(event).Exec(ctx)
		return err
	})
}
