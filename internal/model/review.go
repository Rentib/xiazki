package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Review struct {
	bun.BaseModel `bun:"table:reviews"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Rating    int       `bun:"rating,notnull"` // 1-10
	Opinion   string    `bun:"opinion"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	UserID uuid.UUID `bun:"user_id,notnull"`
	User   *User     `bun:"rel:belongs-to,join:user_id=id"`
	BookID int64     `bun:"book_id,notnull"`
	Book   *Book     `bun:"rel:belongs-to,join:book_id=id"`
}

type ReviewStats struct {
	UserRating    int64   `bun:"user_rating"`
	AverageRating float64 `bun:"average_rating"`
	RatingsCount  int64   `bun:"ratings_count"`
	OpinionsCount int64   `bun:"opinions_count"`
}
