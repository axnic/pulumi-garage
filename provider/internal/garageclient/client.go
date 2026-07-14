// Copyright 2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package garageclient is a minimal client for the Garage Admin API v2
// (https://garagehq.deuxfleurs.fr/documentation/reference-manual/admin-api/),
// covering only the bucket, key, and bucket-key-permission operations
// needed by the Pulumi provider.
package garageclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Client talks to a single Garage cluster's Admin API.
type Client struct {
	endpoint string
	token    string
	http     *http.Client
}

// New creates a Client for the Admin API rooted at endpoint (e.g.
// "http://localhost:3903"), authenticating with the given bearer token.
func New(endpoint, token string) *Client {
	return &Client{
		endpoint: strings.TrimRight(endpoint, "/"),
		token:    token,
		http:     http.DefaultClient,
	}
}

// APIError is returned when the Admin API responds with a non-2xx status.
// Its shape mirrors Garage's error envelope: {"code","message","region","path"}.
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Region     string `json:"region"`
	Path       string `json:"path"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("garage admin api: %d %s: %s", e.StatusCode, e.Code, e.Message)
}

// NotFound reports whether the error corresponds to a 404 response, i.e. the
// bucket, key, or other referenced object does not exist.
func (e *APIError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// do issues an HTTP request against the Admin API and decodes the JSON
// response into out (which may be nil if the caller doesn't need the body).
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	var reqBody bytes.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		reqBody = *bytes.NewReader(encoded)
	}

	u := c.endpoint + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, &reqBody)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("calling garage admin api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		_ = json.NewDecoder(resp.Body).Decode(apiErr)
		return apiErr
	}

	if out == nil {
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}
