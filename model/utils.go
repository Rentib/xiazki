package model

import (
	"github.com/uptrace/bun"
)

type BookAuthor struct {
	bun.BaseModel `bun:"table:book_authors"`

	BookID   int64   `bun:"book_id,pk"`
	Book     *Book   `bun:"rel:belongs-to,join:book_id=id"`
	AuthorID int64   `bun:"author_id,pk"`
	Author   *Author `bun:"rel:belongs-to,join:author_id=id"`
}

type BookTag struct {
	bun.BaseModel `bun:"table:book_tags"`

	BookID int64 `bun:"book_id,pk"`
	Book   *Book `bun:"rel:belongs-to,join:book_id=id"`
	TagID  int64 `bun:"tag_id,pk"`
	Tag    *Tag  `bun:"rel:belongs-to,join:tag_id=id"`
}

type BookTranslator struct {
	bun.BaseModel `bun:"table:book_translators"`

	BookID       int64       `bun:"book_id,pk"`
	Book         *Book       `bun:"rel:belongs-to,join:book_id=id"`
	TranslatorID int64       `bun:"translator_id,pk"`
	Translator   *Translator `bun:"rel:belongs-to,join:translator_id=id"`
}

type BookNarrator struct {
	bun.BaseModel `bun:"table:book_narrators"`

	BookID     int64     `bun:"book_id,pk"`
	Book       *Book     `bun:"rel:belongs-to,join:book_id=id"`
	NarratorID int64     `bun:"narrator_id,pk"`
	Narrator   *Narrator `bun:"rel:belongs-to,join:narrator_id=id"`
}
