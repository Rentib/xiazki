package handler

// NOTE: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html

import (
	"net/http"
	"strings"

	"xiazki/internal/model"
	"xiazki/web/template/auth"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AuthForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func (h *Handler) GetRegister(c echo.Context) error {
	return Render(c, auth.Show(auth.Data{Op: auth.Register}))
}

func (h *Handler) GetLogin(c echo.Context) error {
	return Render(c, auth.Show(auth.Data{Op: auth.Login}))
}

// TODO: password strength meter

func (h *Handler) PostRegister(c echo.Context) error {
	var form AuthForm
	if err := c.Bind(&form); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	data := auth.Data{
		Op:     auth.Register,
		Values: map[string]string{"username": form.Username, "password": form.Password},
		Errors: map[string]string{},
	}

	if form.Username == "" {
		data.Errors["username"] = "Username is required"
	} else if len(form.Username) < 3 {
		data.Errors["username"] = "Username must be at least 3 characters"
	} else if len(form.Username) > 30 {
		data.Errors["username"] = "Username must be at most 30 characters"
	}

	// Check username uniqueness only if basic validation passes
	if len(data.Errors) == 0 {
		var existingUser model.User
		err := h.db.NewSelect().Model(&existingUser).Where("username = ?", form.Username).Scan(c.Request().Context())
		if err == nil {
			data.Errors["username"] = "Username already taken"
		}
	}

	if form.Password == "" {
		data.Errors["password"] = "Password is required"
	} else if len(form.Password) < 8 {
		data.Errors["password"] = "Password must be at least 8 characters"
	} else if len(form.Password) > 128 {
		data.Errors["password"] = "Password must be at most 128 characters"
	}

	if len(data.Errors) > 0 {
		return Render(c, auth.Form(data))
	}

	user := model.User{
		ID:       uuid.New(),
		Username: form.Username,
	}

	userCount, err := h.db.NewSelect().Model((*model.User)(nil)).Count(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check users")
	}

	if userCount == 0 {
		user.Role = model.RoleAdmin
	} else {
		user.Role = model.RoleUser
	}

	if err := user.SetPassword(form.Password); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	if _, err := h.db.NewInsert().Model(&user).Exec(c.Request().Context()); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			data.Errors["username"] = "Username already taken"
			return Render(c, auth.Form(data))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	if err := h.createSession(c, user.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}
	return HxRedirect(c, "/books")
}

func (h *Handler) PostLogin(c echo.Context) error {
	var form AuthForm
	if err := c.Bind(&form); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	data := auth.Data{
		Op:     auth.Login,
		Values: map[string]string{"username": form.Username, "password": form.Password},
		Errors: map[string]string{},
	}

	if form.Username == "" {
		data.Errors["username"] = "Username is required"
		return Render(c, auth.Form(data))
	}
	if form.Password == "" {
		data.Errors["password"] = "Password is required"
		return Render(c, auth.Form(data))
	}

	var user model.User
	err := h.db.NewSelect().Model(&user).Where("username = ?", form.Username).Scan(c.Request().Context())
	if err != nil {
		// Use generic error to avoid revealing whether user exists
		data.Errors["password"] = "Invalid username or password"
		return Render(c, auth.Form(data))
	}

	if !user.CheckPassword(form.Password) {
		// Use generic error to avoid revealing whether user exists
		data.Errors["password"] = "Invalid username or password"
		return Render(c, auth.Form(data))
	}

	if err := h.createSession(c, user.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}
	return HxRedirect(c, "/books")
}

func (h *Handler) PostLogout(c echo.Context) error {
	_ = h.clearSession(c) // FIXME: handle error
	return HxRedirect(c, "/login")
}
