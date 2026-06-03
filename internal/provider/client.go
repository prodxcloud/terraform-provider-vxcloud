package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client is a thin HTTP client for the vxcloud tenant-node API.
//
// All requests authenticate with the developer API key via the X-API-Key
// header (the same xc_dev_/xc_live_ key the CLI and SDK use). agentcontrol
// requests additionally carry X-Tenant-ID / X-Username so the platform can
// scope the call to the right tenant.
//
// Endpoint is the tenant node base URL (e.g. https://node1.vxcloud.io), NOT
// the marketing API host — container deploys and agentcontrol both live on the
// node.
type Client struct {
	Endpoint string
	Email    string
	APIToken string
	TenantID string
	Username string
	HTTP     *http.Client
}

func NewClient(endpoint, email, apiToken string) *Client {
	return &Client{
		Endpoint: strings.TrimRight(endpoint, "/"),
		Email:    email,
		APIToken: apiToken,
		HTTP:     &http.Client{Timeout: 120 * time.Second},
	}
}

// withTenant returns a shallow copy of the client scoped to a specific tenant,
// letting a single resource override the provider-level X-Tenant-ID. An empty
// tenantID returns the receiver unchanged.
func (c *Client) withTenant(tenantID string) *Client {
	if tenantID == "" || tenantID == c.TenantID {
		return c
	}
	clone := *c
	clone.TenantID = tenantID
	return &clone
}

// apiError is returned for any non-2xx response, carrying the body so callers
// can surface the platform's own error message in a Terraform diagnostic.
type apiError struct {
	Status int
	Body   string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("vxcloud API returned %d: %s", e.Status, e.Body)
}

func (c *Client) url(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.Endpoint + path
}

// auth stamps every request with the credentials the platform expects.
func (c *Client) auth(req *http.Request) {
	if c.APIToken != "" {
		req.Header.Set("X-API-Key", c.APIToken)
	}
	if c.TenantID != "" {
		req.Header.Set("X-Tenant-ID", c.TenantID)
	}
	if c.Username != "" {
		req.Header.Set("X-Username", c.Username)
	} else if c.Email != "" {
		req.Header.Set("X-Username", c.Email)
	}
	req.Header.Set("Accept", "application/json")
}

func (c *Client) do(req *http.Request) (map[string]any, error) {
	c.auth(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &apiError{Status: resp.StatusCode, Body: strings.TrimSpace(string(body))}
	}
	out := map[string]any{}
	if len(bytes.TrimSpace(body)) > 0 {
		// Tolerate non-object JSON (arrays, bare values) by stashing it under
		// "_raw" instead of failing the whole call.
		if err := json.Unmarshal(body, &out); err != nil {
			out = map[string]any{"_raw": string(body)}
		}
	}
	return out, nil
}

// PostMultipart submits a multipart/form-data request — used for container
// deploys, which the node accepts as form fields (image, ports, ssh creds, …).
func (c *Client) PostMultipart(ctx context.Context, path string, fields map[string]string) (map[string]any, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if v == "" {
			continue
		}
		if err := w.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(path), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.do(req)
}

// PostJSON submits a JSON body — used for agentcontrol resources.
func (c *Client) PostJSON(ctx context.Context, path string, payload any) (map[string]any, error) {
	return c.sendJSON(ctx, http.MethodPost, path, payload)
}

// PutJSON updates a resource via JSON body.
func (c *Client) PutJSON(ctx context.Context, path string, payload any) (map[string]any, error) {
	return c.sendJSON(ctx, http.MethodPut, path, payload)
}

func (c *Client) sendJSON(ctx context.Context, method, path string, payload any) (map[string]any, error) {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.url(path), body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req)
}

// Get fetches a resource as JSON.
func (c *Client) Get(ctx context.Context, path string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(path), nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// Delete removes a resource. A 404 is treated as already-gone (nil error) so
// Terraform deletes converge.
func (c *Client) Delete(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url(path), nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	if ae, ok := err.(*apiError); ok && ae.Status == http.StatusNotFound {
		return nil
	}
	return err
}

// firstString walks a response map and returns the first non-empty string
// value among the given keys (supports nested "data"/"result" envelopes).
func firstString(m map[string]any, keys ...string) string {
	if m == nil {
		return ""
	}
	lookup := func(src map[string]any) string {
		for _, k := range keys {
			if v, ok := src[k]; ok {
				if s, ok := v.(string); ok && s != "" {
					return s
				}
				if f, ok := v.(float64); ok {
					return fmt.Sprintf("%v", f)
				}
			}
		}
		return ""
	}
	if s := lookup(m); s != "" {
		return s
	}
	for _, env := range []string{"data", "result", "deployment", "agent"} {
		if nested, ok := m[env].(map[string]any); ok {
			if s := lookup(nested); s != "" {
				return s
			}
		}
	}
	return ""
}
