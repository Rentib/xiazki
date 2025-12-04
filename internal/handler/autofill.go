package handler

import (
	"context"
	"io"
	"net/http"
	"sync"

	"xiazki/internal/model"
	"xiazki/internal/utils"
	"xiazki/web/template/add_book"
	"xiazki/web/template/autofill"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetAddBookAutofill(c echo.Context) error {
	// TODO: autofill from title/author

	// TODO: better error handling
	isbn, err := utils.StringToISBN(c.QueryParam("isbn"))
	afv := autofill.AutofillFormValues{
		ISBN: isbn,
	}

	if err != nil {
		errors := map[string]string{}
		if isbn != "" {
			errors = map[string]string{"isbn": "Invalid ISBN format"}
		}
		return Render(c, autofill.AutofillModal(autofill.Data{
			Values: afv,
			Errors: errors,
		}))
	}

	return Render(c, autofill.MatchListModal(autofill.Data{Values: afv}))
}

func (h *Handler) GetAddBookAutofillSSE(c echo.Context) error {
	isbn, err := utils.StringToISBN(c.QueryParam("isbn"))
	if err != nil {
		return err
	}

	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	wg := sync.WaitGroup{}
	matches := make(chan *model.Book, 2)
	done := make(chan struct{})

	for _, fetcher := range h.fetcher {
		wg.Go(func() {
			if book, err := fetcher.GetISBN(isbn); err == nil && book != nil {
				select {
				case matches <- book:
				case <-done:
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(matches)
	}()

	for {
		select {
		case <-c.Request().Context().Done():
			close(done)
			return nil
		case match, ok := <-matches:
			if !ok {
				if _, err := io.WriteString(w, "event: close\ndata:\n\n"); err != nil {
					close(done)
					return err
				}
				w.Flush()
				return nil
			}

			component := autofill.MatchItem(match)
			if _, err := io.WriteString(w, "data: "); err != nil {
				close(done)
				return err
			}
			if err = component.Render(context.Background(), w); err != nil {
				close(done)
				return err
			}
			if _, err := io.WriteString(w, "\n\n"); err != nil {
				close(done)
				return err
			}
			w.Flush()
		}
	}
}

func (h *Handler) PostAddBookAutofillSelect(c echo.Context) error {
	var bfv add_book.BookFormValues
	if err := c.Bind(&bfv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	return Render(c, add_book.Show(add_book.Data{
		Op:     add_book.Add,
		Values: bfv,
	}))
}
