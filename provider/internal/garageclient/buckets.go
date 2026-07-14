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

// CreateBucket creates a new bucket, optionally registering globalAlias as
// its global alias in the same call. Pass an empty globalAlias to create an
// unaliased bucket (addressable only by ID).
func (c *Client) CreateBucket(ctx context.Context, globalAlias string) (*BucketInfo, error) {
	var out BucketInfo
	body := map[string]string{}
	if globalAlias != "" {
		body["globalAlias"] = globalAlias
	}
	if err := c.do(ctx, http.MethodPost, "/v2/CreateBucket", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBucketInfo fetches a bucket by its ID. Returns an *APIError with
// NotFound() true if the bucket doesn't exist.
func (c *Client) GetBucketInfo(ctx context.Context, id string) (*BucketInfo, error) {
	var out BucketInfo
	q := url.Values{"id": {id}}
	if err := c.do(ctx, http.MethodGet, "/v2/GetBucketInfo", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListBuckets lists every bucket in the cluster.
func (c *Client) ListBuckets(ctx context.Context) ([]BucketListEntry, error) {
	var out []BucketListEntry
	if err := c.do(ctx, http.MethodGet, "/v2/ListBuckets", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateBucketInput carries the fields UpdateBucket can change. A nil field
// leaves the corresponding bucket property untouched.
type UpdateBucketInput struct {
	Website *WebsiteAccessUpdate
	Quotas  *BucketQuotas
}

// UpdateBucket changes a bucket's website-hosting configuration and/or
// quotas.
func (c *Client) UpdateBucket(ctx context.Context, id string, in *UpdateBucketInput) (*BucketInfo, error) {
	var out BucketInfo
	q := url.Values{"id": {id}}
	body := map[string]any{}
	if in.Website != nil {
		body["websiteAccess"] = in.Website
	}
	if in.Quotas != nil {
		body["quotas"] = in.Quotas
	}
	if err := c.do(ctx, http.MethodPost, "/v2/UpdateBucket", q, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBucket deletes a bucket. Garage refuses to delete a non-empty
// bucket; the resulting *APIError carries that detail in its Message.
func (c *Client) DeleteBucket(ctx context.Context, id string) error {
	q := url.Values{"id": {id}}
	return c.do(ctx, http.MethodPost, "/v2/DeleteBucket", q, nil, nil)
}

// AddGlobalBucketAlias registers alias as an additional global alias for
// the bucket.
func (c *Client) AddGlobalBucketAlias(ctx context.Context, bucketID, alias string) (*BucketInfo, error) {
	var out BucketInfo
	body := map[string]string{"bucketId": bucketID, "globalAlias": alias}
	if err := c.do(ctx, http.MethodPost, "/v2/AddBucketAlias", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RemoveGlobalBucketAlias removes a global alias from the bucket.
func (c *Client) RemoveGlobalBucketAlias(ctx context.Context, bucketID, alias string) (*BucketInfo, error) {
	var out BucketInfo
	body := map[string]string{"bucketId": bucketID, "globalAlias": alias}
	if err := c.do(ctx, http.MethodPost, "/v2/RemoveBucketAlias", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
