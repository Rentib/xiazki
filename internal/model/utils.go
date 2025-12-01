package model

import (
	"github.com/uptrace/bun"
)

type BookAuthor struct {
	bun.BaseModel `bun:"table:book_authors"`

	BookID   int64   `bun:"book_id,pk,unique:book_author"`
	AuthorID int64   `bun:"author_id,pk,unique:book_author"`
	Book     *Book   `bun:"rel:belongs-to,join:book_id=id"`
	Author   *Author `bun:"rel:belongs-to,join:author_id=id"`
}

type BookTag struct {
	bun.BaseModel `bun:"table:book_tags"`

	BookID int64 `bun:"book_id,pk,unique:book_tag"`
	TagID  int64 `bun:"tag_id,pk,unique:book_tag"`
	Book   *Book `bun:"rel:belongs-to,join:book_id=id"`
	Tag    *Tag  `bun:"rel:belongs-to,join:tag_id=id"`
}

type BookTranslator struct {
	bun.BaseModel `bun:"table:book_translators"`

	BookID       int64       `bun:"book_id,pk,unique:book_translator"`
	TranslatorID int64       `bun:"translator_id,pk,unique:book_translator"`
	Book         *Book       `bun:"rel:belongs-to,join:book_id=id"`
	Translator   *Translator `bun:"rel:belongs-to,join:translator_id=id"`
}

type BookNarrator struct {
	bun.BaseModel `bun:"table:book_narrators"`

	BookID     int64     `bun:"book_id,pk,unique:book_narrator"`
	NarratorID int64     `bun:"narrator_id,pk,unique:book_narrator"`
	Book       *Book     `bun:"rel:belongs-to,join:book_id=id"`
	Narrator   *Narrator `bun:"rel:belongs-to,join:narrator_id=id"`
}
