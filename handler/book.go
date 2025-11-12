package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"xiazki/model"
	"xiazki/view/book"
	"xiazki/view/components"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

func (h *Handler) GetBook(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	var b model.Book
	err = h.db.NewSelect().
		Model(&b).
		Where("id = ?", id).
		Relation("Authors").
		Relation("Tags").
		Relation("Translators").
		Relation("Narrators").
		Scan(c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book details")
	}

	return Render(c, book.Show(
		book.Data{
			CSRF: c.Get("csrf").(string),
			Book: b,
		},
	))
}

func (h *Handler) DeleteBook(c echo.Context) error {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	_, err = h.db.NewDelete().
		Model((*model.Book)(nil)).
		Where("id = ?", id).
		Exec(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete book: "+err.Error())
	}

	return HxRedirect(c, "/books")
}

func (h *Handler) PutBook(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	var dto model.BookDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	if errors, values := validateBook(dto); len(errors) > 0 {
		return Render(c, book.ShowEditForm(
			components.BookFormData{
				CSRF:   c.Get("csrf").(string),
				Book:   model.Book{ID: id},
				Errors: errors,
				Values: values,
			},
		))
	}

	if err := h.updateBookInDatabase(c, id, dto); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update book: "+err.Error())
	}

	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetBookEdit(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	var b model.Book
	err = h.db.NewSelect().
		Model(&b).
		Where("id = ?", id).
		Relation("Authors").
		Relation("Tags").
		Relation("Translators").
		Relation("Narrators").
		Scan(c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch book details")
	}

	values := make(map[string]string)

	// populate only existing values
	if b.Title != "" {
		values["title"] = b.Title
	}
	if b.Authors != nil {
		values["authors"] = AuthorsToString(b.Authors)
	}
	if b.Tags != nil {
		values["tags"] = TagsToString(b.Tags)
	}
	if b.Translators != nil {
		values["translators"] = TranslatorsToString(b.Translators)
	}
	if b.Narrators != nil {
		values["narrators"] = NarratorsToString(b.Narrators)
	}
	if b.Summary != "" {
		values["summary"] = b.Summary
	}
	if b.ISBN10 != "" {
		values["isbn10"] = b.ISBN10
	}
	if b.ISBN13 != "" {
		values["isbn13"] = b.ISBN13
	}
	if b.Language != "" {
		values["language"] = b.Language
	}
	if b.Publisher != "" {
		values["publisher"] = b.Publisher
	}
	if b.PublishDate.IsZero() == false {
		values["publish_date"] = b.PublishDate.Format("2006-01-02")
	}
	if b.PageCount != 0 {
		values["page_count"] = strconv.Itoa(b.PageCount)
	}
	if b.SeriesName != "" {
		values["series_name"] = b.SeriesName
		values["series_number"] = strconv.Itoa(b.SeriesNumber)
	}
	if b.CoverURL != "" {
		values["cover_url"] = b.CoverURL
	}

	return Render(c, book.ShowEditForm(
		components.BookFormData{
			CSRF:   c.Get("csrf").(string),
			Book:   b,
			Errors: make(map[string]string),
			Values: values,
		},
	))
}

func AuthorsToString(authors []model.Author) string {
	authorNames := make([]string, len(authors))
	for i, author := range authors {
		authorNames[i] = author.Name
	}
	return strings.Join(authorNames, ", ")
}

func TagsToString(tags []model.Tag) string {
	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}
	return strings.Join(tagNames, ", ")
}

func TranslatorsToString(translators []model.Translator) string {
	translatorNames := make([]string, len(translators))
	for i, translator := range translators {
		translatorNames[i] = translator.Name
	}
	return strings.Join(translatorNames, ", ")
}

func NarratorsToString(narrators []model.Narrator) string {
	narratorNames := make([]string, len(narrators))
	for i, narrator := range narrators {
		narratorNames[i] = narrator.Name
	}
	return strings.Join(narratorNames, ", ")
}

func (h *Handler) updateBookInDatabase(c echo.Context, id int64, dto model.BookDto) error {
	ctx := c.Request().Context()

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to start transaction",
		})
	}
	defer tx.Rollback()

	book := BookDtoToModel(dto)
	book.ID = id
	book.UpdatedAt = time.Now()

	_, err = tx.NewUpdate().
		Model(book).
		ExcludeColumn("created_at").
		WherePK().
		Exec(ctx)
	if err != nil {
		log.Printf("updateBookInDatabase failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update book",
		})
	}

	// Update relationships if they are provided in the DTO
	if err := h.updateBookRelationships(ctx, tx, id, dto); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update book relationships: " + err.Error(),
		})
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to commit transaction",
		})
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) updateBookRelationships(ctx context.Context, tx bun.Tx, bookID int64, dto model.BookDto) error {
	// Update authors if provided
	if dto.Authors != "" {
		if err := h.updateBookAuthors(ctx, tx, bookID, dto.Authors); err != nil {
			return err
		}
	}

	// Update tags if provided
	if dto.Tags != "" {
		if err := h.updateBookTags(ctx, tx, bookID, dto.Tags); err != nil {
			return err
		}
	}

	// Update translators if provided
	if dto.Translators != "" {
		if err := h.updateBookTranslators(ctx, tx, bookID, dto.Translators); err != nil {
			return err
		}
	}

	// Update narrators if provided
	if dto.Narrators != "" {
		if err := h.updateBookNarrators(ctx, tx, bookID, dto.Narrators); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) updateBookAuthors(ctx context.Context, tx bun.Tx, bookID int64, authorsStr string) error {
	// First, remove existing authors
	_, err := tx.NewDelete().
		Model((*model.BookAuthor)(nil)).
		Where("book_id = ?", bookID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Parse and create new authors
	authorNames := parseCommaSeparated(authorsStr)
	for _, authorName := range authorNames {
		authorName = strings.TrimSpace(authorName)
		if authorName == "" {
			continue
		}

		author := &model.Author{Name: authorName}

		// Try to find existing author first
		err := tx.NewSelect().
			Model(author).
			Where("name = ?", authorName).
			Scan(ctx)

		if err != nil {
			// Author doesn't exist, create new one
			_, err = tx.NewInsert().
				Model(author).
				Exec(ctx)
			if err != nil {
				return err
			}
		}

		// Create relationship
		bookAuthor := &model.BookAuthor{
			BookID:   bookID,
			AuthorID: author.ID,
		}
		_, err = tx.NewInsert().
			Model(bookAuthor).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) updateBookTags(ctx context.Context, tx bun.Tx, bookID int64, tagsStr string) error {
	// Remove existing tags
	_, err := tx.NewDelete().
		Model((*model.BookTag)(nil)).
		Where("book_id = ?", bookID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Parse and create new tags
	tagNames := parseCommaSeparated(tagsStr)
	for _, tagName := range tagNames {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}

		tag := &model.Tag{Name: tagName}

		// Try to find existing tag first
		err := tx.NewSelect().
			Model(tag).
			Where("name = ?", tagName).
			Scan(ctx)

		if err != nil {
			// Tag doesn't exist, create new one
			_, err = tx.NewInsert().
				Model(tag).
				Exec(ctx)
			if err != nil {
				return err
			}
		}

		// Create relationship
		bookTag := &model.BookTag{
			BookID: bookID,
			TagID:  tag.ID,
		}
		_, err = tx.NewInsert().
			Model(bookTag).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) updateBookTranslators(ctx context.Context, tx bun.Tx, bookID int64, translatorsStr string) error {
	// Remove existing translators
	_, err := tx.NewDelete().
		Model((*model.BookTranslator)(nil)).
		Where("book_id = ?", bookID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Parse and create new translators
	translatorNames := parseCommaSeparated(translatorsStr)
	for _, translatorName := range translatorNames {
		translatorName = strings.TrimSpace(translatorName)
		if translatorName == "" {
			continue
		}

		translator := &model.Translator{Name: translatorName}

		// Try to find existing translator first
		err := tx.NewSelect().
			Model(translator).
			Where("name = ?", translatorName).
			Scan(ctx)

		if err != nil {
			// Translator doesn't exist, create new one
			_, err = tx.NewInsert().
				Model(translator).
				Exec(ctx)
			if err != nil {
				return err
			}
		}

		// Create relationship
		bookTranslator := &model.BookTranslator{
			BookID:       bookID,
			TranslatorID: translator.ID,
		}
		_, err = tx.NewInsert().
			Model(bookTranslator).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) updateBookNarrators(ctx context.Context, tx bun.Tx, bookID int64, narratorsStr string) error {
	// Remove existing narrators
	_, err := tx.NewDelete().
		Model((*model.BookNarrator)(nil)).
		Where("book_id = ?", bookID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Parse and create new narrators
	narratorNames := parseCommaSeparated(narratorsStr)
	for _, narratorName := range narratorNames {
		narratorName = strings.TrimSpace(narratorName)
		if narratorName == "" {
			continue
		}

		narrator := &model.Narrator{Name: narratorName}

		// Try to find existing narrator first
		err := tx.NewSelect().
			Model(narrator).
			Where("name = ?", narratorName).
			Scan(ctx)

		if err != nil {
			// Narrator doesn't exist, create new one
			_, err = tx.NewInsert().
				Model(narrator).
				Exec(ctx)
			if err != nil {
				return err
			}
		}

		// Create relationship
		bookNarrator := &model.BookNarrator{
			BookID:     bookID,
			NarratorID: narrator.ID,
		}
		_, err = tx.NewInsert().
			Model(bookNarrator).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Helper function to parse comma-separated strings
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}
