package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	auth_sessions_key string = "session"
	auth_key          string = "authenticated"
	user_id_key       string = "user_id"
)

func (h *Handler) createSession(c echo.Context, userID uuid.UUID) error {
	sess, _ := session.Get(auth_sessions_key, c) // FIXME: handle error
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	sess.Values = map[any]any{
		auth_key:    true,
		user_id_key: userID.String(),
	}
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}
	return nil
}

func (h *Handler) clearSession(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	for key := range sess.Values {
		delete(sess.Values, key)
	}
	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response())
}

func checkSession(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	if auth, err := sess.Values[auth_key].(bool); !err || !auth {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	return nil
}

func (h *Handler) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := checkSession(c); err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		return next(c)
	}
}

func (h *Handler) RequireAuthHTMX(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := checkSession(c); err != nil || c.Request().Header.Get("HX-Request") != "true" {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}
