package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Tag struct {
	bun.BaseModel `bun:"table:tags"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	Books []*Book `bun:"m2m:book_tags,join:Tag=Book"`
}
