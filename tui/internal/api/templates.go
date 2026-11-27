package api

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/MohamedElashri/snipo/pkg/models"
)

var ErrTemplatesUnavailable = errors.New("this Snipo server does not support templates yet")

type Template = models.Template
type TemplateSummary = models.TemplateSummary
type TemplateField = models.TemplateField

func (c *Client) ListTemplates() ([]TemplateSummary, error) {
	var items []TemplateSummary
	for page := 1; ; page++ {
		var response struct {
			Data       []TemplateSummary `json:"data"`
			Pagination Pagination        `json:"pagination"`
		}
		if err := c.doRequest("GET", "/api/v1/templates?limit=100&page="+strconv.Itoa(page), nil, &response); err != nil {
			return nil, err
		}
		items = append(items, response.Data...)
		if page >= response.Pagination.TotalPages {
			return items, nil
		}
	}
}

func (c *Client) GetTemplate(id string, revision int) (*Template, error) {
	var response struct {
		Data Template `json:"data"`
	}
	path := "/api/v1/templates/" + url.PathEscape(id) + "?revision=" + strconv.Itoa(revision)
	if err := c.doRequest("GET", path, nil, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) RenderTemplate(id string, revision int, values map[string]string) (*SnippetInput, error) {
	var response struct {
		Data SnippetInput `json:"data"`
	}
	input := struct {
		Revision int               `json:"revision"`
		Values   map[string]string `json:"values"`
	}{revision, values}
	if err := c.doRequest("POST", "/api/v1/templates/"+url.PathEscape(id)+"/render", input, &response); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
