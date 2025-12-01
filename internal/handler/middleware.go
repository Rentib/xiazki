package handler

import (
	"database/sql"
	"net/http"
	"os"

	"xiazki/internal/model"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	sessionName   = "session"
	authKey       = "authenticated"
	userIDKey     = "user_id"
	sessionMaxAge = 86400 * 7 // 7 days
)

func (h *Handler) createSession(c echo.Context, userID uuid.UUID) error {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return err
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "prod",
		SameSite: http.SameSiteLaxMode,
	}

	sess.Values[authKey] = true
	sess.Values[userIDKey] = userID.String()

	return sess.Save(c.Request(), c.Response())
}

func (h *Handler) clearSession(c echo.Context) error {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return err
	}

	sess.Values = make(map[any]any)
	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response())
}

func (h *Handler) currentUser(c echo.Context) (*model.User, error) {
	userID, err := checkSession(c)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	var user model.User
	if err = h.db.NewSelect().Model(&user).Where("id = ?", userID).Scan(c.Request().Context()); err != nil {
		if err == sql.ErrNoRows {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "User not found")
		}
		return nil, err
	}

	return &user, nil
}

func (h *Handler) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, err := checkSession(c); err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		return next(c)
	}
}

func (h *Handler) RequireAuthHTMX(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, err := checkSession(c); err != nil || c.Request().Header.Get("HX-Request") != "true" {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}

func checkSession(c echo.Context) (uuid.UUID, error) {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return uuid.Nil, err
	}

	if auth, ok := sess.Values[authKey].(bool); !ok || !auth {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	userIDStr, ok := sess.Values[userIDKey].(string)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid session")
	}

	if userID, err := uuid.Parse(userIDStr); err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	} else {
		return userID, nil
	}
}
