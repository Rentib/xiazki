package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"xiazki/view/add_book"
)

type OpenLibraryService struct {
	client  *http.Client
	baseURL string
}

type author struct {
	Name string `json:"name"`
}

type publisher struct {
	Name string `json:"name"`
}

type language struct {
	Key string `json:"key"`
}

type excerpt struct {
	Text string `json:"text"`
}

type bookData struct {
	Title       string      `json:"title"`
	PublishDate string      `json:"publish_date"`
	NumberPages int         `json:"number_of_pages"`
	Authors     []author    `json:"authors"`
	Publishers  []publisher `json:"publishers"`
	Identifiers struct {
		ISBN10 []string `json:"isbn_10"`
		ISBN13 []string `json:"isbn_13"`
	} `json:"identifiers"`
	Cover struct {
		Large string `json:"large"`
	} `json:"cover"`
	Languages []language `json:"languages"`
	Excerpts  []excerpt  `json:"excerpts"`
	Notes     any        `json:"notes"`
}

func NewOpenLibraryService() *OpenLibraryService {
	return &OpenLibraryService{
		client:  http.DefaultClient,
		baseURL: "https://openlibrary.org/api/books",
	}
}

func (s *OpenLibraryService) FetchBookByISBN(isbn string) (*add_book.BookFormValues, error) {
	fmt.Println("Fetching book data for ISBN:", isbn)

	if isbn == "" {
		return nil, fmt.Errorf("ISBN cannot be empty")
	}

	url := fmt.Sprintf("%s?bibkeys=ISBN:%s&format=json&jscmd=data", s.baseURL, isbn)
	fmt.Println("Requesting URL:", url)
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("Received response with status:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data map[string]bookData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	book, exists := data["ISBN:"+isbn]
	if !exists {
		return nil, fmt.Errorf("book with ISBN %s not found", isbn)
	}

	return &add_book.BookFormValues{
		Title:       book.Title,
		Authors:     s.joinAuthors(book.Authors),
		Publisher:   s.joinPublishers(book.Publishers),
		PublishDate: book.PublishDate,
		PageCount:   fmt.Sprintf("%d", book.NumberPages),
		ISBN10:      s.firstOrEmpty(book.Identifiers.ISBN10),
		ISBN13:      s.firstOrEmpty(book.Identifiers.ISBN13),
		Language:    s.extractLanguage(book.Languages),
		CoverURL:    book.Cover.Large,
		Summary:     s.extractSummary(book.Excerpts, book.Notes),
	}, nil
}

func (s *OpenLibraryService) joinAuthors(authors []author) string {
	if len(authors) == 0 {
		return ""
	}
	names := make([]string, len(authors))
	for i, a := range authors {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}

func (s *OpenLibraryService) joinPublishers(publishers []publisher) string {
	if len(publishers) == 0 {
		return ""
	}
	names := make([]string, len(publishers))
	for i, p := range publishers {
		names[i] = p.Name
	}
	return strings.Join(names, ", ")
}

func (s *OpenLibraryService) firstOrEmpty(items []string) string {
	if len(items) > 0 {
		return items[0]
	}
	return ""
}

func (s *OpenLibraryService) extractLanguage(langs []language) string {
	if len(langs) == 0 {
		return ""
	}
	parts := strings.Split(langs[0].Key, "/")
	return parts[len(parts)-1]
}

func (s *OpenLibraryService) extractSummary(excerpts []excerpt, notes any) string {
	if len(excerpts) > 0 {
		return excerpts[0].Text
	}
	switch v := notes.(type) {
	case string:
		return v
	case map[string]any:
		if value, ok := v["value"].(string); ok {
			return value
		}
	}
	return ""
}
