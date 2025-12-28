package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("API error: %s", errResp.Error.Message)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

func (c *Client) Health() (*HealthResponse, error) {
	var response struct {
		Data HealthResponse `json:"data"`
		Meta Meta           `json:"meta"`
	}
	if err := c.doRequest("GET", "/health", nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) ListSnippets(page, limit int, query string, tagIDs, folderIDs []int, language string, favorite, archived *bool) ([]Snippet, *Pagination, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if query != "" {
		params.Set("q", query)
	}
	if language != "" {
		params.Set("language", language)
	}
	if favorite != nil {
		params.Set("favorite", strconv.FormatBool(*favorite))
	}
	if archived != nil {
		params.Set("is_archived", strconv.FormatBool(*archived))
	}
	for _, id := range tagIDs {
		params.Add("tag_ids", strconv.Itoa(id))
	}
	for _, id := range folderIDs {
		params.Add("folder_ids", strconv.Itoa(id))
	}

	path := "/api/v1/snippets"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var response ListResponse
	if err := c.doRequest("GET", path, nil, &response); err != nil {
		return nil, nil, err
	}

	snippetsData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, nil, err
	}

	var snippets []Snippet
	if err := json.Unmarshal(snippetsData, &snippets); err != nil {
		return nil, nil, err
	}

	return snippets, &response.Pagination, nil
}

func (c *Client) GetSnippet(id string) (*Snippet, error) {
	var response APIResponse
	if err := c.doRequest("GET", fmt.Sprintf("/api/v1/snippets/%s", id), nil, &response); err != nil {
		return nil, err
	}

	snippetData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var snippet Snippet
	if err := json.Unmarshal(snippetData, &snippet); err != nil {
		return nil, err
	}

	return &snippet, nil
}

func (c *Client) CreateSnippet(input SnippetInput) (*Snippet, error) {
	var response APIResponse
	if err := c.doRequest("POST", "/api/v1/snippets", input, &response); err != nil {
		return nil, err
	}

	snippetData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var snippet Snippet
	if err := json.Unmarshal(snippetData, &snippet); err != nil {
		return nil, err
	}

	return &snippet, nil
}

func (c *Client) UpdateSnippet(id string, input SnippetInput) (*Snippet, error) {
	var response APIResponse
	if err := c.doRequest("PUT", fmt.Sprintf("/api/v1/snippets/%s", id), input, &response); err != nil {
		return nil, err
	}

	snippetData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var snippet Snippet
	if err := json.Unmarshal(snippetData, &snippet); err != nil {
		return nil, err
	}

	return &snippet, nil
}

func (c *Client) DeleteSnippet(id string) error {
	return c.doRequest("DELETE", fmt.Sprintf("/api/v1/snippets/%s", id), nil, nil)
}

func (c *Client) ToggleFavorite(id string) (*Snippet, error) {
	var response APIResponse
	if err := c.doRequest("POST", fmt.Sprintf("/api/v1/snippets/%s/favorite", id), nil, &response); err != nil {
		return nil, err
	}

	snippetData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var snippet Snippet
	if err := json.Unmarshal(snippetData, &snippet); err != nil {
		return nil, err
	}

	return &snippet, nil
}

func (c *Client) ListTags() ([]Tag, error) {
	var response ListResponse
	if err := c.doRequest("GET", "/api/v1/tags", nil, &response); err != nil {
		return nil, err
	}

	tagsData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var tags []Tag
	if err := json.Unmarshal(tagsData, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}

func (c *Client) CreateTag(input TagInput) (*Tag, error) {
	var response APIResponse
	if err := c.doRequest("POST", "/api/v1/tags", input, &response); err != nil {
		return nil, err
	}

	tagData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var tag Tag
	if err := json.Unmarshal(tagData, &tag); err != nil {
		return nil, err
	}

	return &tag, nil
}

func (c *Client) ListFolders() ([]Folder, error) {
	var response ListResponse
	if err := c.doRequest("GET", "/api/v1/folders", nil, &response); err != nil {
		return nil, err
	}

	foldersData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var folders []Folder
	if err := json.Unmarshal(foldersData, &folders); err != nil {
		return nil, err
	}

	return folders, nil
}

func (c *Client) CreateFolder(input FolderInput) (*Folder, error) {
	var response APIResponse
	if err := c.doRequest("POST", "/api/v1/folders", input, &response); err != nil {
		return nil, err
	}

	folderData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var folder Folder
	if err := json.Unmarshal(folderData, &folder); err != nil {
		return nil, err
	}

	return &folder, nil
}
