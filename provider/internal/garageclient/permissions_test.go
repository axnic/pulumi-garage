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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowBucketKey(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/AllowBucketKey", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(bucketInfoWithKeyAndWebsiteFixture))
	})

	bucket, err := c.AllowBucketKey(context.Background(), "bucket-id", "key-id", BucketKeyPerm{Read: true, Write: true})
	require.NoError(t, err)

	assert.Equal(t, "bucket-id", gotBody["bucketId"])
	assert.Equal(t, "key-id", gotBody["accessKeyId"])
	perms, ok := gotBody["permissions"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, perms["read"])
	assert.Equal(t, true, perms["write"])
	assert.Equal(t, false, perms["owner"])
	assert.True(t, bucket.Keys[0].Permissions.Read)
}

func TestDenyBucketKey(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/DenyBucketKey", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(bucketInfoWithKeyAndWebsiteFixture))
	})

	_, err := c.DenyBucketKey(context.Background(), "bucket-id", "key-id", BucketKeyPerm{Owner: true})
	require.NoError(t, err)

	perms, ok := gotBody["permissions"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, perms["owner"])
	assert.Equal(t, false, perms["read"])
}
