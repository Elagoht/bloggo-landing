package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Author struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type CategoryInfo struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type TagInfo struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1] // Remove quotes

	// Try multiple formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			ct.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse time: %s", s)
}

type Post struct {
	ID          string       `json:"id"`
	Slug        string       `json:"slug"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Excerpt     string       `json:"excerpt"`
	Description string       `json:"description"`
	Author      Author       `json:"author"`
	CreatedAt   CustomTime   `json:"createdAt"`
	UpdatedAt   CustomTime   `json:"updatedAt"`
	PublishedAt CustomTime   `json:"publishedAt"`
	Category    CategoryInfo `json:"category"`
	Tags        []TagInfo    `json:"tags"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Fetcher struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

func New(baseURL, apiToken string) *Fetcher {
	return &Fetcher{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *Fetcher) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if f.apiToken != "" {
		req.Header.Set("x-trusted-frontend", f.apiToken)
	}

	return f.httpClient.Do(req)
}

func (f *Fetcher) GetAllPosts() ([]Post, error) {
	resp, err := f.doRequest(fmt.Sprintf("%s/posts", f.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Data []Post `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal posts: %w", err)
	}

	return response.Data, nil
}

func (f *Fetcher) GetPostBySlug(slug string) (*Post, error) {
	resp, err := f.doRequest(fmt.Sprintf("%s/posts/%s", f.baseURL, slug))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("post not found: %s", slug)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var post Post
	if err := json.Unmarshal(body, &post); err != nil {
		return nil, fmt.Errorf("failed to unmarshal post: %w", err)
	}

	return &post, nil
}

func (f *Fetcher) GetAllCategories() ([]Category, error) {
	resp, err := f.doRequest(fmt.Sprintf("%s/categories", f.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var categories []Category
	if err := json.Unmarshal(body, &categories); err != nil {
		return nil, fmt.Errorf("failed to unmarshal categories: %w", err)
	}

	return categories, nil
}

func (f *Fetcher) GetAllTags() ([]Tag, error) {
	resp, err := f.doRequest(fmt.Sprintf("%s/tags", f.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var tags []Tag
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return tags, nil
}

func (f *Fetcher) GetDocuments() ([]Post, error) {
	resp, err := f.doRequest(fmt.Sprintf("%s/posts?category=documents", f.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Data []Post `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal documents: %w", err)
	}

	return response.Data, nil
}
