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
	PageCount    int       `bun:"page_count"`
	SeriesName   string    `bun:"series_name"`
	SeriesNumber int       `bun:"series_number"`
	CoverURL     string    `bun:"cover_url"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Authors     []Author     `bun:"m2m:book_authors,join:Book=Author"`
	Tags        []Tag        `bun:"m2m:book_tags,join:Book=Tag"`
	Translators []Translator `bun:"m2m:book_translators,join:Book=Translator"`
	Narrators   []Narrator   `bun:"m2m:book_narrators,join:Book=Narrator"`
}

type BookDto struct {
	Authors      string `json:"authors" form:"authors"`
	CoverURL     string `json:"cover_url" form:"cover_url"`
	ISBN10       string `json:"isbn10" form:"isbn10"`
	ISBN13       string `json:"isbn13" form:"isbn13"`
	Language     string `json:"language" form:"language"`
	Narrators    string `json:"narrators" form:"narrators"`
	PageCount    string `json:"page_count" form:"page_count"`
	PublishDate  string `json:"publish_date" form:"publish_date"`
	Publisher    string `json:"publisher" form:"publisher"`
	SeriesName   string `json:"series_name" form:"series_name"`
	SeriesNumber string `json:"series_number" form:"series_number"`
	Summary      string `json:"summary" form:"summary"`
	Tags         string `json:"tags" form:"tags"`
	Title        string `json:"title" form:"title"`
	Translators  string `json:"translators" form:"translators"`
}
