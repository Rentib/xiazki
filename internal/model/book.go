package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Book struct {
	bun.BaseModel `bun:"table:books"`

	ID           int64     `bun:"id,pk,autoincrement"`
	Title        string    `bun:"title,notnull"`
	Summary      string    `bun:"summary,type:text,nullzero"`
	ISBN10       string    `bun:"isbn10,nullzero"`
	ISBN13       string    `bun:"isbn13,nullzero"`
	Language     string    `bun:"language,nullzero"`
	Publisher    string    `bun:"publisher,nullzero"`
	PublishDate  time.Time `bun:"publish_date,nullzero"`
	PageCount    int64     `bun:"page_count,nullzero"`
	SeriesName   string    `bun:"series_name,nullzero"`
	SeriesNumber int64     `bun:"series_number,nullzero"`
	CoverURL     string    `bun:"cover_url,nullzero"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Authors     []*Author     `bun:"m2m:book_authors,join:Book=Author"`
	Tags        []*Tag        `bun:"m2m:book_tags,join:Book=Tag"`
	Translators []*Translator `bun:"m2m:book_translators,join:Book=Translator"`
	Narrators   []*Narrator   `bun:"m2m:book_narrators,join:Book=Narrator"`

	Events []*Event `bun:"rel:has-many,join:id=book_id"`
}
