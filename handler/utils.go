package handler

import (
	"context"
	"net/http"

	"xiazki/services"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type Handler struct {
	db *bun.DB
	ol *services.OpenLibraryService
}

func NewHandler(db *bun.DB) *Handler {
	return &Handler{
		db: db,
		ol: services.NewOpenLibraryService(),
	}
}

func Render(c echo.Context, component templ.Component) error {
	return component.Render(context.Background(), c.Response())
}

func HxRedirect(c echo.Context, path string) error {
	c.Response().Header().Set("HX-Redirect", path)
	return c.NoContent(http.StatusOK)
}
