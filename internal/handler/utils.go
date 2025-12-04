package handler

import (
	"context"
	"net/http"

	"xiazki/internal/model"
	"xiazki/internal/services/googlebooks"
	"xiazki/internal/services/openlibrary"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type Fetcher interface {
	GetISBN(string) (*model.Book, error)
}

type Handler struct {
	db      *bun.DB
	fetcher []Fetcher
}

func NewHandler(db *bun.DB, gbAPIKey string) *Handler {
	return &Handler{
		db: db,
		fetcher: []Fetcher{
			googlebooks.NewFetcher(gbAPIKey),
			openlibrary.NewFetcher(),
		},
	}
}

func Render(c echo.Context, component templ.Component) error {
	csrf := c.Get("csrf")
	ctx := context.WithValue(c.Request().Context(), "X-CSRF-Token", csrf)
	return component.Render(ctx, c.Response())
}

func HxRedirect(c echo.Context, path string) error {
	c.Response().Header().Set("HX-Redirect", path)
	return c.NoContent(http.StatusOK)
}
