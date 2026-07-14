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
)

// AllowBucketKey grants accessKeyID exactly the permissions in perm that are
// true on bucketID; permissions left false in perm are untouched (see
// BucketKeyPerm's doc comment).
func (c *Client) AllowBucketKey(
	ctx context.Context, bucketID, accessKeyID string, perm BucketKeyPerm,
) (*BucketInfo, error) {
	var out BucketInfo
	body := map[string]any{
		"bucketId":    bucketID,
		"accessKeyId": accessKeyID,
		"permissions": perm,
	}
	if err := c.do(ctx, http.MethodPost, "/v2/AllowBucketKey", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DenyBucketKey revokes accessKeyID exactly the permissions in perm that are
// true on bucketID; permissions left false in perm are untouched (see
// BucketKeyPerm's doc comment).
func (c *Client) DenyBucketKey(
	ctx context.Context, bucketID, accessKeyID string, perm BucketKeyPerm,
) (*BucketInfo, error) {
	var out BucketInfo
	body := map[string]any{
		"bucketId":    bucketID,
		"accessKeyId": accessKeyID,
		"permissions": perm,
	}
	if err := c.do(ctx, http.MethodPost, "/v2/DenyBucketKey", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
