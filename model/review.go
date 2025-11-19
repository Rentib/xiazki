package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Review struct {
	bun.BaseModel `bun:"table:reviews"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Rating    int       `bun:"rating,notnull"` // 1-10
	Opinion   string    `bun:"opinion"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	UserID int64 `bun:"user_id,notnull"`
	User   *User `bun:"rel:belongs-to,join:user_id=id"`
	BookID int64 `bun:"book_id,notnull"`
	Book   *Book `bun:"rel:belongs-to,join:book_id=id"`
}
