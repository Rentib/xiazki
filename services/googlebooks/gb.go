package googlebooks

// https://developers.google.com/books/v1

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

type Volumes struct {
	Items []*Volume `json:"items,omitempty"`
}

type Volume struct {
	VolumeInfo *VolumeVolumeInfo `json:"volumeInfo,omitempty"`
}

type VolumeVolumeInfo struct {
	Authors             []string                               `json:"authors,omitempty"`
	CanonicalVolumeLink string                                 `json:"canonicalVolumeLink,omitempty"`
	Categories          []string                               `json:"categories,omitempty"`
	ContentVersion      string                                 `json:"contentVersion,omitempty"`
	Description         string                                 `json:"description,omitempty"`
	ImageLinks          *VolumeVolumeInfoImageLinks            `json:"imageLinks,omitempty"`
	IndustryIdentifiers []*VolumeVolumeInfoIndustryIdentifiers `json:"industryIdentifiers,omitempty"`
	InfoLink            string                                 `json:"infoLink,omitempty"`
	Language            string                                 `json:"language,omitempty"`
	MainCategory        string                                 `json:"mainCategory,omitempty"`
	PageCount           int64                                  `json:"pageCount,omitempty"`
	PublishedDate       string                                 `json:"publishedDate,omitempty"`
	Publisher           string                                 `json:"publisher,omitempty"`
	SeriesInfo          *Volumeseriesinfo                      `json:"seriesInfo,omitempty"`
	Subtitle            string                                 `json:"subtitle,omitempty"`
	Title               string                                 `json:"title,omitempty"`
}

type VolumeVolumeInfoImageLinks struct {
	ExtraLarge     string `json:"extraLarge,omitempty"`
	Large          string `json:"large,omitempty"`
	Medium         string `json:"medium,omitempty"`
	Small          string `json:"small,omitempty"`
	SmallThumbnail string `json:"smallThumbnail,omitempty"`
	Thumbnail      string `json:"thumbnail,omitempty"`
}

type VolumeVolumeInfoIndustryIdentifiers struct {
	Identifier string `json:"identifier,omitempty"` // Identifier: Industry specific volume identifier.
	Type       string `json:"type,omitempty"`       // Type: Identifier type. Possible values are ISBN_10, ISBN_13, ISSN and OTHER.
}

type Volumeseriesinfo struct {
	ShortSeriesBookTitle string                          `json:"shortSeriesBookTitle,omitempty"`
	VolumeSeries         []*VolumeseriesinfoVolumeSeries `json:"volumeSeries,omitempty"`
}

type VolumeseriesinfoVolumeSeries struct {
	OrderNumber    int64  `json:"orderNumber,omitempty"`
	SeriesBookType string `json:"seriesBookType,omitempty"`
	SeriesID       string `json:"seriesId,omitempty"`
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
	var resp Volumes
	if err := get(f.client, url, &resp); err != nil {
		return nil, fmt.Errorf("failed to get volume data: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no book found for ISBN: %s", isbn)
	}

	volume := resp.Items[0].VolumeInfo
	book := &model.Book{
		Title:   volume.Title,
		Summary: volume.Description,
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
		Language:  volume.Language,
		Publisher: volume.Publisher,
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
		SeriesName: func() string {
			if volume.SeriesInfo != nil {
				return volume.SeriesInfo.ShortSeriesBookTitle
			}
			return ""
		}(),
		SeriesNumber: func() int64 {
			if volume.SeriesInfo != nil && len(volume.SeriesInfo.VolumeSeries) > 0 {
				return volume.SeriesInfo.VolumeSeries[0].OrderNumber
			}
			return 0
		}(),
		CoverURL: func() string {
			if volume.ImageLinks != nil {
				if volume.ImageLinks.ExtraLarge != "" {
					return volume.ImageLinks.ExtraLarge
				}
				if volume.ImageLinks.Large != "" {
					return volume.ImageLinks.Large
				}
				if volume.ImageLinks.Medium != "" {
					return volume.ImageLinks.Medium
				}
				if volume.ImageLinks.Small != "" {
					return volume.ImageLinks.Small
				}
				if volume.ImageLinks.Thumbnail != "" {
					return volume.ImageLinks.Thumbnail
				}
				if volume.ImageLinks.SmallThumbnail != "" {
					return volume.ImageLinks.SmallThumbnail
				}
			}
			return ""
		}(),

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
