package aadrm

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/satori/go.uuid"
)

// Client uses DefaultBaseURL if not specified
var DefaultBaseURL *url.URL

func init() {
	var err error
	DefaultBaseURL, err = url.Parse("https://api.aadrm.com")
	if err != nil {
		panic(errors.Wrap(err, "failed to parse DefaultBaseURL"))
	}
}

// Client interacts with aadrm
type Client struct {
	c *http.Client
	// BaseURL is DefaultBaseURL if unspecified
	BaseURL *url.URL
	// Becomes the X-MS-RMS-Platform-Id header if set
	RMSPlatformID string
	// Becomes User-Agent header if set
	UserAgent string
}

// NewRequest creates a request with the expected aadrm headers
func (c *Client) NewRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	u := c.BaseURL.ResolveReference(&url.URL{
		Path: path,
	})
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call http.NewRequestWithContext")
	}
	// Used in logs, I think
	if c.RMSPlatformID != "" {
		req.Header.Set("X-MS-RMS-Platform-Id", c.RMSPlatformID)
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	// Must be unique per request
	requestID, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate UUID")
	}
	req.Header.Set("X-MS-RMS-Request-Id", requestID.String())
	return req, nil
}

// NewClient creates a client, the caller must supply a client with auth (ex: via oauth2.TokenSource)
func NewClient(c *http.Client) *Client {
	return &Client{c: c, BaseURL: DefaultBaseURL}
}
