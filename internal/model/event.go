package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type EventType string

const (
	EventFinished EventType = "finished"
	EventReading  EventType = "reading"
	EventDropped  EventType = "dropped"
)

type Event struct {
	bun.BaseModel `bun:"table:events"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Type      EventType `bun:"type,notnull"`
	Date      time.Time `bun:"date,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	UserID uuid.UUID `bun:"user_id,notnull"`
	BookID int64     `bun:"book_id,notnull"`
	User   *User     `bun:"rel:belongs-to,join:user_id=id"`
	Book   *Book     `bun:"rel:belongs-to,join:book_id=id"`
}
