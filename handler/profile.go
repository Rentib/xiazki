package handler

import (
	"fmt"
	"net/http"

	"xiazki/model"
	"xiazki/view/profile"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GetProfile(c echo.Context) error {
	user, err := getUser(c, h)
	if err != nil {
		return err
	}

	return Render(c, profile.Show(profile.Data{
		CSRF: c.Get("csrf").(string),
		User: user,
	}))
}

func (h *Handler) PostUserChangePassword(c echo.Context) error {
	user, err := getUser(c, h)
	if err != nil {
		return err
	}

	var cpfv profile.ChangePasswordFormValues
	if err := c.Bind(&cpfv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	if errors := cpfv.Validate(user); len(errors) > 0 {
		return Render(c, profile.ChangePasswordForm(profile.Data{
			CSRF:   c.Get("csrf").(string),
			User:   user,
			Values: cpfv,
			Errors: errors,
		}))
	}

	if err := user.SetPassword(cpfv.NewPassword); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to set new password: %v", err))
	}

	if _, err := h.db.NewUpdate().Model(user).WherePK().Exec(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to update password: %v", err))
	}

	// TODO: flash message "Password changed successfully"
	return HxRedirect(c, "/profile")
}

func getUser(c echo.Context, h *Handler) (*model.User, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	userIDStr, ok := sess.Values["user_id"].(string)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var user model.User
	err = h.db.NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Scan(c.Request().Context())
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch profile: %v", err))
	}
	return &user, nil
}
