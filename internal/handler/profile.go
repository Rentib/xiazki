package handler

import (
	"fmt"
	"net/http"

	"xiazki/web/template/profile"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetProfile(c echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}

	return Render(c, profile.Show(profile.Data{
		CSRF: c.Get("csrf").(string),
		User: user,
	}))
}

func (h *Handler) PostUserChangePassword(c echo.Context) error {
	user, err := h.currentUser(c)
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
