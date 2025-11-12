package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"xiazki/model"
	"xiazki/view/add_book"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

func (h *Handler) GetAddBook(c echo.Context) error {
	return Render(c, add_book.Show(
		add_book.Data{
			CSRF:   c.Get("csrf").(string),
			Errors: make(map[string]string),
			Values: make(map[string]string),
		},
	))
}

func (h *Handler) PostAddBook(c echo.Context) error {
	var dto model.BookDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	if errors, values := validateBook(dto); len(errors) > 0 {
		return Render(c, add_book.Form(add_book.Data{
			CSRF:   c.Get("csrf").(string),
			Errors: errors,
			Values: values,
		}))
	}

	if err := h.insertBookToDatabase(c, dto); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add book: "+err.Error())
	}

	return HxRedirect(c, "/books")
}

func (h *Handler) insertBookToDatabase(c echo.Context, dto model.BookDto) error {
	tx, err := h.db.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	book := BookDtoToModel(dto)
	if _, err := tx.NewInsert().Model(book).Exec(c.Request().Context()); err != nil {
		return err
	}

	// Process all relationships
	processAuthors(&tx, c, book.ID, dto.Authors)
	processTags(&tx, c, book.ID, dto.Tags)
	processTranslators(&tx, c, book.ID, dto.Translators)
	processNarrators(&tx, c, book.ID, dto.Narrators)

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func BookDtoToModel(dto model.BookDto) *model.Book {
	book := &model.Book{
		Title: dto.Title,
	}

	if dto.Summary != "" {
		book.Summary = dto.Summary
	}
	if dto.ISBN10 != "" {
		book.ISBN10 = dto.ISBN10
	}
	if dto.ISBN13 != "" {
		book.ISBN13 = dto.ISBN13
	}
	if dto.Language != "" {
		book.Language = dto.Language
	}
	if dto.Publisher != "" {
		book.Publisher = dto.Publisher
	}
	if dto.PublishDate != "" {
		if t, err := stringToDate(dto.PublishDate); err == nil {
			book.PublishDate = t
		}
	}
	if dto.PageCount != "" {
		if pc, err := strconv.Atoi(dto.PageCount); err == nil {
			book.PageCount = pc
		}
	}
	if dto.SeriesName != "" {
		book.SeriesName = dto.SeriesName
	}
	if dto.SeriesNumber != "" {
		if sn, err := strconv.Atoi(dto.SeriesNumber); err == nil {
			book.SeriesNumber = sn
		}
	}
	if dto.CoverURL != "" {
		book.CoverURL = dto.CoverURL
	}

	return book
}

// processAuthors processes and links authors to a book
func processAuthors(tx *bun.Tx, c echo.Context, bookID int64, authorsStr string) {
	if authorsStr == "" {
		return
	}

	authorsList := strings.SplitSeq(authorsStr, ",")
	for authorName := range authorsList {
		authorName = strings.TrimSpace(authorName)
		if authorName == "" {
			continue
		}

		var author model.Author
		err := tx.NewSelect().Model(&author).Where("name = ?", authorName).Scan(c.Request().Context())

		if err != nil {
			author = model.Author{Name: authorName}
			_, err = tx.NewInsert().Model(&author).Exec(c.Request().Context())
			if err != nil {
				continue
			}
		}

		bookAuthor := &model.BookAuthor{
			BookID:   bookID,
			AuthorID: author.ID,
		}
		_, err = tx.NewInsert().Model(bookAuthor).Exec(c.Request().Context())
		if err != nil {
			continue
		}
	}
}

func processTags(tx *bun.Tx, c echo.Context, bookID int64, tagsStr string) {
	if tagsStr == "" {
		return
	}

	tagsList := strings.SplitSeq(tagsStr, ",")
	for tagName := range tagsList {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}

		var tag model.Tag
		err := tx.NewSelect().Model(&tag).Where("name = ?", tagName).Scan(c.Request().Context())
		if err != nil {
			tag = model.Tag{Name: tagName}
			_, err = tx.NewInsert().Model(&tag).Exec(c.Request().Context())
			if err != nil {
				continue
			}
		}

		bookTag := &model.BookTag{BookID: bookID, TagID: tag.ID}
		tx.NewInsert().Model(bookTag).Exec(c.Request().Context())
	}
}

// processTranslators processes and links translators to a book
func processTranslators(tx *bun.Tx, c echo.Context, bookID int64, translatorsStr string) {
	if translatorsStr == "" {
		return
	}

	translatorsList := strings.SplitSeq(translatorsStr, ",")
	for translatorName := range translatorsList {
		translatorName = strings.TrimSpace(translatorName)
		if translatorName == "" {
			continue
		}

		var translator model.Translator
		err := tx.NewSelect().Model(&translator).Where("name = ?", translatorName).Scan(c.Request().Context())
		if err != nil {
			translator = model.Translator{Name: translatorName}
			_, err = tx.NewInsert().Model(&translator).Exec(c.Request().Context())
			if err != nil {
				continue
			}
		}

		bookTranslator := &model.BookTranslator{BookID: bookID, TranslatorID: translator.ID}
		tx.NewInsert().Model(bookTranslator).Exec(c.Request().Context())
	}
}

// processNarrators processes and links narrators to a book
func processNarrators(tx *bun.Tx, c echo.Context, bookID int64, narratorsStr string) {
	if narratorsStr == "" {
		return
	}

	narratorsList := strings.SplitSeq(narratorsStr, ",")
	for narratorName := range narratorsList {
		narratorName = strings.TrimSpace(narratorName)
		if narratorName == "" {
			continue
		}

		var narrator model.Narrator
		err := tx.NewSelect().Model(&narrator).Where("name = ?", narratorName).Scan(c.Request().Context())
		if err != nil {
			narrator = model.Narrator{Name: narratorName}
			_, err = tx.NewInsert().Model(&narrator).Exec(c.Request().Context())
			if err != nil {
				continue
			}
		}

		bookNarrator := &model.BookNarrator{BookID: bookID, NarratorID: narrator.ID}
		tx.NewInsert().Model(bookNarrator).Exec(c.Request().Context())
	}
}

func validateBook(dto model.BookDto) (map[string]string, map[string]string) {
	// Validate form
	errors := make(map[string]string)
	values := make(map[string]string)

	// Store values for repopulation
	values["title"] = dto.Title
	values["authors"] = dto.Authors
	values["tags"] = dto.Tags
	values["translators"] = dto.Translators
	values["narrators"] = dto.Narrators
	values["summary"] = dto.Summary
	values["isbn10"] = dto.ISBN10
	values["isbn13"] = dto.ISBN13
	values["language"] = dto.Language
	values["publisher"] = dto.Publisher
	values["publish_date"] = dto.PublishDate
	values["page_count"] = dto.PageCount
	values["series_name"] = dto.SeriesName
	values["series_number"] = dto.SeriesNumber
	values["cover_url"] = dto.CoverURL

	// Validation rules
	if strings.TrimSpace(dto.Title) == "" {
		errors["title"] = "Title is required"
	}

	if strings.TrimSpace(dto.Authors) == "" {
		errors["authors"] = "At least one author is required"
	} else {
		authors := strings.Split(dto.Authors, ",")
		validAuthors := 0
		for _, author := range authors {
			if strings.TrimSpace(author) != "" {
				validAuthors++
			}
		}
		if validAuthors == 0 {
			errors["authors"] = "At least one valid author is required"
		}
	}

	if dto.ISBN10 != "" && !isValidISBN10(dto.ISBN10) {
		errors["isbn10"] = "Invalid ISBN-10 format"
	}

	if dto.ISBN13 != "" && !isValidISBN13(dto.ISBN13) {
		errors["isbn13"] = "Invalid ISBN-13 format"
	}

	if dto.PageCount != "" {
		cnt, err := strconv.Atoi(dto.PageCount)
		if err != nil {
			errors["page_count"] = "Page count must be a number"
		}
		if cnt <= 0 {
			errors["page_count"] = "Page count cannot be negative"
		}
	}

	if dto.SeriesNumber != "" {
		if _, err := strconv.Atoi(dto.SeriesNumber); err != nil {
			errors["series_number"] = "Series number must be a number"
		}
	}

	if dto.CoverURL != "" && !isValidURL(dto.CoverURL) {
		errors["cover_url"] = "Invalid URL format"
	}
	return errors, values
}

// Helper validation functions (keep these)
func isValidISBN10(isbn string) bool {
	// Remove hyphens and spaces
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")

	if len(isbn) != 10 {
		return false
	}

	sum := 0
	for i := range 9 {
		digit := int(isbn[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		sum += digit * (10 - i)
	}

	lastChar := isbn[9]
	if lastChar == 'X' || lastChar == 'x' {
		sum += 10
	} else {
		digit := int(lastChar - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		sum += digit
	}

	return sum%11 == 0
}

func isValidISBN13(isbn string) bool {
	// Remove hyphens and spaces
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")

	if len(isbn) != 13 {
		return false
	}

	sum := 0
	for i := range 13 {
		digit := int(isbn[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	return sum%10 == 0
}

func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func stringToDate(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02",
		"2006/01/02",
		"02-01-2006",
		"02/01/2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}
