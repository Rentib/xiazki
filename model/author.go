package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Author struct {
	bun.BaseModel `bun:"table:authors"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Books []*Book `bun:"m2m:book_authors,join:Author=Book"`
}
