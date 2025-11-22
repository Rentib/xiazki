package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Book struct {
	bun.BaseModel `bun:"table:books"`

	ID           int64     `bun:"id,pk,autoincrement"`
	Title        string    `bun:"title,notnull"`
	Summary      string    `bun:"summary"`
	ISBN10       string    `bun:"isbn10"`
	ISBN13       string    `bun:"isbn13"`
	Language     string    `bun:"language"`
	Publisher    string    `bun:"publisher"`
	PublishDate  time.Time `bun:"publish_date"`
	PageCount    int64     `bun:"page_count"`
	SeriesName   string    `bun:"series_name"`
	SeriesNumber int64     `bun:"series_number"`
	CoverURL     string    `bun:"cover_url"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Authors     []*Author     `bun:"m2m:book_authors,join:Book=Author"`
	Tags        []*Tag        `bun:"m2m:book_tags,join:Book=Tag"`
	Translators []*Translator `bun:"m2m:book_translators,join:Book=Translator"`
	Narrators   []*Narrator   `bun:"m2m:book_narrators,join:Book=Narrator"`

	Events []*Event `bun:"rel:has-many,join:id=book_id"`
}
