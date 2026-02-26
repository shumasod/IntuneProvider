// Package client provides a Microsoft Graph API client for Intune operations.
// It handles Azure AD authentication via client credentials flow and wraps
// the common CRUD operations used across all Intune resources.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const (
	graphBaseURL   = "https://graph.microsoft.com/v1.0"
	graphScope     = "https://graph.microsoft.com/.default"
	defaultTimeout = 60 * time.Second
)

// Client is a Microsoft Graph API client authenticated with Azure AD client credentials.
type Client struct {
	httpClient *http.Client
	credential *azidentity.ClientSecretCredential
	baseURL    string
}

// GraphError represents an error response from the Microsoft Graph API.
type GraphError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a new Graph API client using client credentials (service principal) authentication.
func NewClient(tenantID, clientID, clientSecret string) (*Client, error) {
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		credential: cred,
		baseURL:    graphBaseURL,
	}, nil
}

// getAccessToken retrieves a valid access token for the Microsoft Graph API.
func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	token, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{graphScope},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	return token.Token, nil
}

// doRequest executes an HTTP request against the Graph API with authentication.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, 0, err
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var graphErr GraphError
		if jsonErr := json.Unmarshal(respBody, &graphErr); jsonErr == nil && graphErr.Error.Code != "" {
			return nil, resp.StatusCode, fmt.Errorf("Graph API error %d: [%s] %s", resp.StatusCode, graphErr.Error.Code, graphErr.Error.Message)
		}
		return nil, resp.StatusCode, fmt.Errorf("Graph API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}

// Get performs a GET request and unmarshals the JSON response into result.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	body, _, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

// Post performs a POST request and unmarshals the JSON response into result.
func (c *Client) Post(ctx context.Context, path string, payload, result interface{}) error {
	body, _, err := c.doRequest(ctx, http.MethodPost, path, payload)
	if err != nil {
		return err
	}
	if result != nil {
		return json.Unmarshal(body, result)
	}
	return nil
}

// Patch performs a PATCH request (partial update).
func (c *Client) Patch(ctx context.Context, path string, payload interface{}) error {
	_, _, err := c.doRequest(ctx, http.MethodPatch, path, payload)
	return err
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	_, status, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil && status != http.StatusNoContent {
		return err
	}
	return nil
}

// IsNotFound returns true if the error corresponds to a 404 Not Found response.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return containsStatusCode(err.Error(), "404")
}

func containsStatusCode(msg, code string) bool {
	return len(msg) > 0 && (contains(msg, "status "+code) || contains(msg, "error "+code))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
