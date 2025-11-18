package handler

// NOTE: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html

import (
	"net/http"

	"xiazki/model"
	"xiazki/view/auth"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AuthForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
	CSRF     string `form:"csrf"`
}

func (h *Handler) GetRegister(c echo.Context) error {
	return Render(c, auth.Show(auth.Data{Op: auth.Register, CSRF: c.Get("csrf").(string)}))
}

func (h *Handler) GetLogin(c echo.Context) error {
	return Render(c, auth.Show(auth.Data{Op: auth.Login, CSRF: c.Get("csrf").(string)}))
}

// TODO: password strength meter

func (h *Handler) PostRegister(c echo.Context) error {
	var form AuthForm
	if err := c.Bind(&form); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	// FIXME: is returnign password back safe???
	data := auth.Data{
		Op:     auth.Register,
		CSRF:   c.Get("csrf").(string),
		Values: map[string]string{"username": form.Username, "password": form.Password},
		Errors: map[string]string{},
	}

	var user model.User
	if form.Username == "" {
		data.Errors["username"] = "Username is required"
		return Render(c, auth.Form(data))
	} else if err := h.db.NewSelect().Model(&user).Where("username = ?", form.Username).Scan(c.Request().Context()); err == nil {
		data.Errors["username"] = "Username already taken"
		return Render(c, auth.Form(data))
	}

	if form.Password == "" {
		data.Errors["password"] = "Password is required"
		return Render(c, auth.Form(data))
	} else if len(form.Password) < 8 {
		data.Errors["password"] = "Password must be at least 8 characters"
		return Render(c, auth.Form(data))
	} else if len(form.Password) > 128 {
		data.Errors["password"] = "Password must be at most 128 characters"
		return Render(c, auth.Form(data))
	}

	user = model.User{
		ID:       uuid.New(),
		Username: form.Username,
	}
	if userCount, err := h.db.NewSelect().Model((*model.User)(nil)).Count(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check users")
	} else if userCount == 0 {
		user.Role = model.RoleAdmin
	} else if userCount != 0 {
		user.Role = model.RoleUser
	}

	if err := user.SetPassword(form.Password); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	if _, err := h.db.NewInsert().Model(&user).Exec(c.Request().Context()); err != nil {
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

	// FIXME: is returnign password back safe???
	data := auth.Data{
		Op:     auth.Login,
		CSRF:   c.Get("csrf").(string),
		Values: map[string]string{"username": form.Username, "password": form.Password},
		Errors: map[string]string{},
	}

	var user model.User
	err := h.db.NewSelect().Model(&user).Where("username = ?", form.Username).Scan(c.Request().Context())
	if err != nil {
		data.Errors["username"] = "No such user"
		return Render(c, auth.Form(data))
	}

	if !user.CheckPassword(form.Password) {
		data.Errors["password"] = "Incorrect password"
		return Render(c, auth.Form(data))
	}

	if err := h.createSession(c, user.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}
	return HxRedirect(c, "/books")
}

func (h *Handler) PostLogout(c echo.Context) error {
	h.clearSession(c)
	return HxRedirect(c, "/login")
}
