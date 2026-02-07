package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// Book represents a book from Open Library.
type Book struct {
	Key              string
	Title            string
	AuthorName       string
	FirstPublishYear int
	CoverID          int
	PageCount        int
	EditionCount     int
	Subjects         []string
}

// CoverURL returns the Open Library cover image URL for the given size (S, M, L).
// Returns empty string if the book has no cover.
func (b Book) CoverURL(size string) string {
	if b.CoverID <= 0 {
		return ""
	}

	return fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-%s.jpg", b.CoverID, size)
}

// OpenLibraryURL returns the Open Library page URL for this book.
func (b Book) OpenLibraryURL() string {
	return "https://openlibrary.org" + b.Key
}

type searchResponse struct {
	Docs []struct {
		Key              string   `json:"key"`
		Title            string   `json:"title"`
		AuthorName       []string `json:"author_name"`
		FirstPublishYear int      `json:"first_publish_year"`
		CoverID          int      `json:"cover_i"`
		PageCount        int      `json:"number_of_pages_median"`
		EditionCount     int      `json:"edition_count"`
		Subjects         []string `json:"subject"`
	} `json:"docs"`
}

// BooksClient is an HTTP client for the Open Library Search API.
type BooksClient struct {
	Doer *http.Client
}

const searchFields = "key,title,author_name,first_publish_year,cover_i,number_of_pages_median,edition_count,subject"

func (c *BooksClient) Search(ctx context.Context, query string, offset, limit int) ([]Book, error) {
	u, err := url.Parse("https://openlibrary.org/search.json")
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	q.Set("q", query)
	q.Set("fields", searchFields)

	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}

	u.RawQuery = q.Encode()

	return c.doSearch(ctx, u.String())
}

func (c *BooksClient) SearchByAuthor(ctx context.Context, query, author string, offset, limit int) ([]Book, error) {
	u, err := url.Parse("https://openlibrary.org/search.json")
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	q.Set("author", author)
	q.Set("fields", searchFields)

	if query != "" {
		q.Set("q", query)
	}

	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}

	u.RawQuery = q.Encode()

	return c.doSearch(ctx, u.String())
}

func (c *BooksClient) doSearch(ctx context.Context, rawURL string) ([]Book, error) {
	log.Printf("GET %s", rawURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	res, err := c.Doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	var resp searchResponse

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	books := make([]Book, len(resp.Docs))
	for i, doc := range resp.Docs {
		authorName := "Unknown"
		if len(doc.AuthorName) > 0 {
			authorName = doc.AuthorName[0]
		}

		subjects := doc.Subjects
		if len(subjects) > 3 {
			subjects = subjects[:3]
		}

		books[i] = Book{
			Key:              doc.Key,
			Title:            doc.Title,
			AuthorName:       authorName,
			FirstPublishYear: doc.FirstPublishYear,
			CoverID:          doc.CoverID,
			PageCount:        doc.PageCount,
			EditionCount:     doc.EditionCount,
			Subjects:         subjects,
		}
	}

	return books, nil
}
