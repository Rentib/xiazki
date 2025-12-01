package openlibrary

// https://github.com/internetarchive/openlibrary-client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"xiazki/internal/model"
	"xiazki/internal/utils"
)

type KeyStruct struct {
	Key string `json:"key"`
}

type Work struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Covers      []int  `json:"covers"`

	Authors []struct {
		Author KeyStruct `json:"author"`
		Type   KeyStruct `json:"type"`
	} `json:"authors"`

	SubjectPeople []string `json:"subject_people"`
	SubjectPlaces []string `json:"subject_places"`
	SubjectTimes  []string `json:"subject_times"`
	Subjects      []string `json:"subjects"`

	Key            string    `json:"key"`
	Type           KeyStruct `json:"type"`
	Location       string    `json:"location"`
	Revision       int       `json:"revision"`
	LatestRevision int       `json:"latest_revision"`
	Created        struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"created"`
	LastModified struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"last_modified"`
}

type Edition struct {
	Title         string      `json:"title"`
	ISBN10        []string    `json:"isbn_10"`
	ISBN13        []string    `json:"isbn_13"`
	Languages     []KeyStruct `json:"languages"`
	Publishers    []string    `json:"publishers"`
	PublishDate   string      `json:"publish_date"`
	NumberOfPages int64       `json:"number_of_pages"`
	Covers        []int       `json:"covers"`

	Works       []KeyStruct         `json:"works"`
	Authors     []KeyStruct         `json:"authors"`
	Identifiers map[string][]string `json:"identifiers"`

	Key             string              `json:"key"`
	Type            KeyStruct           `json:"type"`
	LocalID         []string            `json:"local_id"`
	OCAID           string              `json:"ocaid"`
	Classifications map[string][]string `json:"classifications"`
	Contributions   []string            `json:"contributions"`
	SourceRecords   []string            `json:"source_records"`
	Revision        int                 `json:"revision"`
	LatestRevision  int                 `json:"latest_revision"`
	Created         struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"created"`
	LastModified struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"last_modified"`
}

type Author struct {
	Name         string `json:"name"`
	Bio          string `json:"bio"`
	PersonalName string `json:"personal_name"`

	Type           KeyStruct           `json:"type"`
	SourceRecords  []string            `json:"source_records"`
	AlternateNames []string            `json:"alternate_names"`
	RemoteIDs      map[string][]string `json:"remote_ids"`
	DeathDate      string              `json:"death_date"`
	BirthDate      string              `json:"birth_date"`
	Photos         []int               `json:"photos"`
	Links          []struct {
		Title string    `json:"title"`
		URL   string    `json:"url"`
		Type  KeyStruct `json:"type"`
	} `json:"links"`
	Key            string `json:"key"`
	LatestRevision int    `json:"latest_revision"`
	Revision       int    `json:"revision"`
	Created        struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"created"`
	LastModified struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"last_modified"`
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

type Fetcher struct {
	client  *http.Client
	baseURL string
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: "https://openlibrary.org",
	}
}

func (f *Fetcher) GetISBN(isbn string) (*model.Book, error) {
	if !utils.IsValidISBN(isbn) {
		return nil, fmt.Errorf("not a valid ISBN")
	}

	url := fmt.Sprintf("%s/isbn/%s.json", f.baseURL, isbn)
	var edition Edition
	if err := get(f.client, url, &edition); err != nil {
		return nil, fmt.Errorf("failed to get edition data: %w", err)
	}

	var work Work
	if len(edition.Works) > 0 {
		workURL := fmt.Sprintf("%s%s.json", f.baseURL, edition.Works[0].Key)
		if err := get(f.client, workURL, &work); err != nil {
			return nil, fmt.Errorf("failed to get work data: %w", err)
		}
	}

	var authors []Author
	for _, authorRef := range edition.Authors {
		authorURL := fmt.Sprintf("%s%s.json", f.baseURL, authorRef.Key)
		var author Author
		if err := get(f.client, authorURL, &author); err != nil {
			return nil, fmt.Errorf("failed to get author data: %w", err)
		}
		authors = append(authors, author)
	}

	book := &model.Book{
		Title:   edition.Title,
		Summary: work.Description,
		ISBN10: func() string {
			if len(edition.ISBN10) > 0 {
				return edition.ISBN10[0]
			}
			return ""
		}(),
		ISBN13: func() string {
			if len(edition.ISBN13) > 0 {
				return edition.ISBN13[0]
			}
			return ""
		}(),
		Language: func() string {
			if len(edition.Languages) > 0 {
				return edition.Languages[0].Key
			}
			return ""
		}(),
		Publisher: func() string {
			if len(edition.Publishers) > 0 {
				return edition.Publishers[0]
			}
			return ""
		}(),
		PublishDate: func() time.Time {
			formats := []string{
				"2006",
				"Jan 2006",
				"Jan 2, 2006",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, edition.PublishDate); err == nil {
					return t
				}
			}
			return time.Time{}
		}(),
		PageCount: edition.NumberOfPages,
		// SeriesName
		// SeriesNumber
		CoverURL: func() string {
			if len(edition.Covers) > 0 {
				return fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", edition.Covers[0])
			}
			if len(work.Covers) > 0 {
				return fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", work.Covers[0])
			}
			return ""
		}(),

		Authors: func() []*model.Author {
			var result []*model.Author
			for _, a := range authors {
				result = append(result, &model.Author{
					Name: a.Name,
				})
			}
			return result
		}(),
		Tags: func() []*model.Tag {
			var result []*model.Tag
			for _, subject := range work.Subjects {
				result = append(result, &model.Tag{
					Name: subject,
				})
			}
			return result
		}(),
		Translators: []*model.Translator{},
		Narrators:   []*model.Narrator{},
	}

	return book, nil
}
