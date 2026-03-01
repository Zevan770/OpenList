package anyshare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/drivers/base"
	log "github.com/sirupsen/logrus"
)

func (d *AnyShare) apiURL(path string) string {
	addr := strings.TrimRight(d.Address, "/")
	return addr + "/api" + path
}

// request performs an authenticated API request and decodes the JSON response into out.
// If out is nil, the response body is discarded.
func (d *AnyShare) request(ctx context.Context, method, apiPath string, body interface{}, out interface{}) error {
	url := d.apiURL(apiPath)

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := base.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request %s %s: %w", method, apiPath, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Debugf("AnyShare API error: %s %s -> %d: %s", method, apiPath, resp.StatusCode, string(respBody))
		return fmt.Errorf("AnyShare API error: %s %s returned %d: %s", method, apiPath, resp.StatusCode, string(respBody))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response from %s %s: %w", method, apiPath, err)
		}
	}

	return nil
}

// get performs a GET request with optional query parameters appended to the path.
func (d *AnyShare) get(ctx context.Context, apiPath string, out interface{}) error {
	return d.request(ctx, http.MethodGet, apiPath, nil, out)
}

// post performs a POST request with a JSON body.
func (d *AnyShare) post(ctx context.Context, apiPath string, body interface{}, out interface{}) error {
	return d.request(ctx, http.MethodPost, apiPath, body, out)
}
