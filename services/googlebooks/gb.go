package googlebooks

// https://developers.google.com/books

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"xiazki/model"
	"xiazki/utils"
)

type Fetcher struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func NewFetcher(apiKey string) *Fetcher {
	return &Fetcher{
		client:  http.DefaultClient,
		baseURL: "https://www.googleapis.com/books/v1/volumes",
		apiKey:  apiKey,
	}
}

type Response struct {
	Items []struct {
		VolumeInfo Volume `json:"volumeInfo"`
	} `json:"items"`
}

type Volume struct {
	Title               string     `json:"title"`
	Authors             []string   `json:"authors"`
	PublishedDate       string     `json:"publishedDate"`
	IndustryIdentifiers []struct { // ISBN_10, ISBN_13
		Type       string `json:"type"`
		Identifier string `json:"identifier"`
	} `json:"industryIdentifiers"`
	ReadingModels struct {
		Text  bool `json:"text"`
		Image bool `json:"Image"`
	} `json:"readingModes"`
	PageCount  int      `json:"pageCount"`
	PrintType  string   `json:"printType"` // BOOK
	Categories []string `json:"categories"`
	Language   string   `json:"language"`
}

func get(client *http.Client, url string, target any) error {
	response, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch data: status code %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

func (f *Fetcher) GetISBN(isbn string) (*model.Book, error) {
	if !utils.IsValidISBN(isbn) {
		return nil, fmt.Errorf("invalid ISBN: %s", isbn)
	}

	url := fmt.Sprintf("%s?q=isbn:%s&key=%s", f.baseURL, isbn, f.apiKey)
	var resp Response
	if err := get(f.client, url, &resp); err != nil {
		return nil, fmt.Errorf("failed to get volume data: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no book found for ISBN: %s", isbn)
	}

	volume := resp.Items[0].VolumeInfo
	book := &model.Book{
		Title: volume.Title,
		// Summary:
		ISBN10: func() string {
			for _, id := range volume.IndustryIdentifiers {
				if id.Type == "ISBN_10" {
					return id.Identifier
				}
			}
			return ""
		}(),
		ISBN13: func() string {
			for _, id := range volume.IndustryIdentifiers {
				if id.Type == "ISBN_13" {
					return id.Identifier
				}
			}
			return ""
		}(),
		Language: volume.Language,
		// Publisher:
		PublishDate: func() time.Time {
			formats := []string{
				"2006",
				"Jan 2006",
				"Jan 2, 2006",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, volume.PublishedDate); err == nil {
					return t
				}
			}
			return time.Time{}
		}(),
		PageCount: volume.PageCount,
		// SeriesName
		// SeriesNumber
		// CoverURL

		Authors: func() []*model.Author {
			var authors []*model.Author
			for _, name := range volume.Authors {
				authors = append(authors, &model.Author{Name: name})
			}
			return authors
		}(),
		Tags: func() []*model.Tag {
			var tags []*model.Tag
			for _, category := range volume.Categories {
				tags = append(tags, &model.Tag{Name: category})
			}
			return tags
		}(),
		// Translators:
		// Narrators:
	}

	return book, nil
}
