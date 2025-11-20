package handler

import (
	"net/http"

	"xiazki/utils"
	"xiazki/view/add_book"
	"xiazki/view/autofill"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetAddBookAutofill(c echo.Context) error {
	// TODO: autofill from title/author

	// TODO: better error handling
	isbn, err := utils.StringToISBN(c.QueryParam("isbn"))
	afv := autofill.AutofillFormValues{
		ISBN: isbn,
	}
	matches := []add_book.BookFormValues{}

	if err != nil {
		errors := map[string]string{}
		if isbn != "" {
			errors = map[string]string{"isbn": "Invalid ISBN format"}
		}
		return Render(c, autofill.AutofillModal(autofill.Data{
			CSRF:   c.Get("csrf").(string),
			Values: afv,
			Errors: errors,
		}))
	}

	// TODO: more sources
	if book, err := h.gb.GetISBN(afv.ISBN); err == nil && book != nil {
		matches = append(matches, add_book.BookToBookFormValues(*book))
	}
	if book, err := h.ol.GetISBN(afv.ISBN); err == nil && book != nil {
		matches = append(matches, add_book.BookToBookFormValues(*book))
	}

	if len(matches) == 0 {
		// TODO: show info that no matches were found
		return Render(c, autofill.AutofillModal(autofill.Data{
			CSRF:   c.Get("csrf").(string),
			Values: afv,
		}))
	}

	return Render(c, autofill.MatchListModal(autofill.Data{
		CSRF:    c.Get("csrf").(string),
		Matches: matches,
	}))
}

func (h *Handler) PostAddBookAutofillSelect(c echo.Context) error {
	var bfv add_book.BookFormValues
	if err := c.Bind(&bfv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	return Render(c, add_book.Show(add_book.Data{
		CSRF:   c.Get("csrf").(string),
		Op:     add_book.Add,
		Values: bfv,
	}))
}
