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

package garageclient

import (
	"context"
	"net/http"
	"net/url"
)

// CreateKey creates a new S3 access key. name may be empty, in which case
// Garage assigns a default name.
func (c *Client) CreateKey(ctx context.Context, name string) (*KeyInfo, error) {
	var out KeyInfo
	body := map[string]string{}
	if name != "" {
		body["name"] = name
	}
	if err := c.do(ctx, http.MethodPost, "/v2/CreateKey", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetKeyInfo fetches a key by its access key ID. Returns an *APIError with
// NotFound() true if the key doesn't exist.
func (c *Client) GetKeyInfo(ctx context.Context, accessKeyID string) (*KeyInfo, error) {
	var out KeyInfo
	q := url.Values{"id": {accessKeyID}}
	if err := c.do(ctx, http.MethodGet, "/v2/GetKeyInfo", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListKeys lists all access keys known to the cluster.
func (c *Client) ListKeys(ctx context.Context) ([]KeyListEntry, error) {
	var out []KeyListEntry
	if err := c.do(ctx, http.MethodGet, "/v2/ListKeys", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateKey renames an existing access key.
func (c *Client) UpdateKey(ctx context.Context, accessKeyID, name string) (*KeyInfo, error) {
	var out KeyInfo
	q := url.Values{"id": {accessKeyID}}
	body := map[string]string{"name": name}
	if err := c.do(ctx, http.MethodPost, "/v2/UpdateKey", q, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteKey deletes an access key. Deleting an already-deleted key returns
// an *APIError with NotFound() true.
func (c *Client) DeleteKey(ctx context.Context, accessKeyID string) error {
	q := url.Values{"id": {accessKeyID}}
	return c.do(ctx, http.MethodPost, "/v2/DeleteKey", q, nil, nil)
}
