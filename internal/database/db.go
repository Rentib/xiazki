package database

import (
	"context"
	"database/sql"
	"log"

	"xiazki/internal/model"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type DB struct {
	*bun.DB
}

func InitDB() (*DB, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, "file:xiazki.db?cache=shared")
	if err != nil {
		return nil, err
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	models := []any{
		(*model.User)(nil),
		(*model.Book)(nil),
		(*model.Author)(nil),
		(*model.Tag)(nil),
		(*model.Translator)(nil),
		(*model.Narrator)(nil),
		(*model.Review)(nil),
		(*model.Event)(nil),
		(*model.Quote)(nil),
		(*model.BookAuthor)(nil),
		(*model.BookTag)(nil),
		(*model.BookTranslator)(nil),
		(*model.BookNarrator)(nil),
	}

	db.RegisterModel(
		(*model.BookAuthor)(nil),
		(*model.BookTag)(nil),
		(*model.BookTranslator)(nil),
		(*model.BookNarrator)(nil),
	)

	ctx := context.Background()
	for _, model := range models {
		_, err = db.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	log.Println("Database initialized successfully")
	return &DB{db}, nil
}
