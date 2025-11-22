package handler

import (
	"context"
	"net/http"

	"xiazki/internal/services/googlebooks"
	"xiazki/internal/services/openlibrary"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type Handler struct {
	db *bun.DB
	gb *googlebooks.Fetcher
	ol *openlibrary.Fetcher
}

func NewHandler(db *bun.DB, gbAPIKey string) *Handler {
	return &Handler{
		db: db,
		gb: googlebooks.NewFetcher(gbAPIKey),
		ol: openlibrary.NewFetcher(),
	}
}

func Render(c echo.Context, component templ.Component) error {
	return component.Render(context.Background(), c.Response())
}

func HxRedirect(c echo.Context, path string) error {
	c.Response().Header().Set("HX-Redirect", path)
	return c.NoContent(http.StatusOK)
}
