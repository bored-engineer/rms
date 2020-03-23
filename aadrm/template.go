package aadrm

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type Template struct {
	ID          *string `json:"Id,omitempty"`
	Name        *string `json:"Name,omitempty"`
	Description *string `json:"Description,omitempty"`
}

// ListTemplates calls /my/v2/templates
func (c *Client) ListTemplates(ctx context.Context) ([]Template, *http.Response, error) {
	req, err := c.NewRequest(ctx, "GET", "/my/v2/templates", nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create Request")
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, resp, errors.Wrap(err, "failed to do Request")
	}
	defer resp.Body.Close()
	var templates []Template
	if err := json.NewDecoder(resp.Body).Decode(&templates); err != nil {
		return nil, resp, errors.Wrap(err, "failed decode JSON")
	}
	return templates, resp, nil
}
