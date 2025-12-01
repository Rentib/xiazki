package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Quote struct {
	bun.BaseModel `bun:"table:quotes"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Quote     string    `bun:"quote,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	UserID uuid.UUID `bun:"user_id,notnull,unique:user_book"`
	BookID int64     `bun:"book_id,notnull,unique:user_book"`
	User   *User     `bun:"rel:belongs-to,join:user_id=id"`
	Book   *Book     `bun:"rel:belongs-to,join:book_id=id"`
}
