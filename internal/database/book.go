package database

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"xiazki/internal/model"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

func (db *DB) InsertBook(c echo.Context, book *model.Book) error {
	ctx := c.Request().Context()

	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(book).Exec(ctx); err != nil {
			return err
		} else if err := insertBookRelation(ctx, tx, book.ID, book.Authors, newBookAuthor); err != nil {
			return err
		} else if err := insertBookRelation(ctx, tx, book.ID, book.Tags, newBookTag); err != nil {
			return err
		} else if err := insertBookRelation(ctx, tx, book.ID, book.Translators, newBookTranslator); err != nil {
			return err
		} else if err := insertBookRelation(ctx, tx, book.ID, book.Narrators, newBookNarrator); err != nil {
			return err
		}
		return nil
	})
}

func insertBookRelation[T any](ctx context.Context, tx bun.Tx, bookID int64, items []*T, newLink func(bookID, id int64) any) error {
	for _, item := range items {
		name := reflect.ValueOf(item).Elem().FieldByName("Name").String()
		if err := tx.NewSelect().Model(item).Where("name = ?", name).Scan(ctx); err != nil {
			if _, err := tx.NewInsert().Model(item).Exec(ctx); err != nil {
				return fmt.Errorf("insert base '%s': %w", name, err)
			}
		}
		id := reflect.ValueOf(item).Elem().FieldByName("ID").Int()
		link := newLink(bookID, id)
		if _, err := tx.NewInsert().Model(link).Exec(ctx); err != nil {
			return fmt.Errorf("insert link: %w", err)
		}
	}

	return nil
}

func (db *DB) UpdateBook(c echo.Context, id int64, book *model.Book) error {
	ctx := c.Request().Context()

	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		book.ID = id
		book.UpdatedAt = time.Now()

		_, err := tx.NewUpdate().
			Model(book).
			ExcludeColumn("created_at").
			WherePK().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("update book: %w", err)
		}

		if err := updateBookRelation(ctx, tx, book.ID, book.Authors, newBookAuthor, (*model.BookAuthor)(nil)); err != nil {
			return fmt.Errorf("update relationships: %w", err)
		} else if err := updateBookRelation(ctx, tx, book.ID, book.Tags, newBookTag, (*model.BookTag)(nil)); err != nil {
			return fmt.Errorf("update relationships: %w", err)
		} else if err := updateBookRelation(ctx, tx, book.ID, book.Translators, newBookTranslator, (*model.BookTranslator)(nil)); err != nil {
			return fmt.Errorf("update relationships: %w", err)
		} else if err := updateBookRelation(ctx, tx, book.ID, book.Narrators, newBookNarrator, (*model.BookNarrator)(nil)); err != nil {
			return fmt.Errorf("update relationships: %w", err)
		}

		return nil
	})
}

func updateBookRelation[T any, L any](ctx context.Context, tx bun.Tx, bookID int64, items []*T, newLink func(bookID, id int64) any, linkTable L) error {
	_, err := tx.NewDelete().
		Model(linkTable).
		Where("book_id = ?", bookID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete relations: %w", err)
	}

	return insertBookRelation(ctx, tx, bookID, items, newLink)
}

func newBookAuthor(bookID, id int64) any {
	return &model.BookAuthor{BookID: bookID, AuthorID: id}
}

func newBookTag(bookID, id int64) any {
	return &model.BookTag{BookID: bookID, TagID: id}
}

func newBookTranslator(bookID, id int64) any {
	return &model.BookTranslator{BookID: bookID, TranslatorID: id}
}

func newBookNarrator(bookID, id int64) any {
	return &model.BookNarrator{BookID: bookID, NarratorID: id}
}
