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

// Fixtures below are verbatim captures from a real Garage v2.3.0 single-node
// instance (dxflrs/garage:v2.3.0), not hand-guessed from the OpenAPI spec.

const createBucketResponseFixture = `{
	"id": "a46a69e81cb40c767c1bee786c9517b5576f5dc4a908e5aa5bde04f9b76344c1",
	"created": "2026-07-14T10:04:53.512Z",
	"globalAliases": ["my-test-bucket"],
	"websiteAccess": false,
	"keys": [],
	"objects": 0,
	"bytes": 0,
	"unfinishedUploads": 0,
	"unfinishedMultipartUploads": 0,
	"unfinishedMultipartUploadParts": 0,
	"unfinishedMultipartUploadBytes": 0,
	"quotas": {"maxSize": null, "maxObjects": null}
}`

const bucketInfoWithKeyAndWebsiteFixture = `{
	"id": "a46a69e81cb40c767c1bee786c9517b5576f5dc4a908e5aa5bde04f9b76344c1",
	"created": "2026-07-14T10:04:53.512Z",
	"globalAliases": ["my-test-bucket"],
	"websiteAccess": true,
	"websiteConfig": {"indexDocument": "index.html", "errorDocument": "error.html", "routingRules": []},
	"keys": [{
		"accessKeyId": "GK26b837222cdd03f4c13a05fb",
		"name": "renamed-key",
		"permissions": {"read": true, "write": true, "owner": false},
		"bucketLocalAliases": []
	}],
	"objects": 0,
	"bytes": 0,
	"unfinishedUploads": 0,
	"unfinishedMultipartUploads": 0,
	"unfinishedMultipartUploadParts": 0,
	"unfinishedMultipartUploadBytes": 0,
	"quotas": {"maxSize": 1000000, "maxObjects": 100}
}`

const listBucketsResponseFixture = `[{
	"id": "a46a69e81cb40c767c1bee786c9517b5576f5dc4a908e5aa5bde04f9b76344c1",
	"created": "2026-07-14T10:04:53.512Z",
	"globalAliases": ["my-test-bucket"],
	"localAliases": []
}]`

const notFoundBucketFixture = `{
	"code": "NoSuchBucket",
	"message": "Bucket not found: xyz",
	"region": "garage",
	"path": "/v2/GetBucketInfo"
}`

func TestCreateBucket(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/CreateBucket", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(createBucketResponseFixture))
	})

	bucket, err := c.CreateBucket(context.Background(), "my-test-bucket")
	require.NoError(t, err)

	assert.Equal(t, "my-test-bucket", gotBody["globalAlias"])
	assert.Equal(t, "a46a69e81cb40c767c1bee786c9517b5576f5dc4a908e5aa5bde04f9b76344c1", bucket.ID)
	assert.Equal(t, []string{"my-test-bucket"}, bucket.GlobalAliases)
}

func TestCreateBucketWithoutAlias(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(createBucketResponseFixture))
	})

	_, err := c.CreateBucket(context.Background(), "")
	require.NoError(t, err)
	_, hasAlias := gotBody["globalAlias"]
	assert.False(t, hasAlias, "globalAlias should be omitted from the request body when empty")
}

func TestGetBucketInfo(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/GetBucketInfo", r.URL.Path)
		assert.Equal(t, "bucket-id", r.URL.Query().Get("id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(bucketInfoWithKeyAndWebsiteFixture))
	})

	bucket, err := c.GetBucketInfo(context.Background(), "bucket-id")
	require.NoError(t, err)

	assert.True(t, bucket.WebsiteAccess)
	require.NotNil(t, bucket.WebsiteConfig)
	assert.Equal(t, "index.html", bucket.WebsiteConfig.IndexDocument)
	require.Len(t, bucket.Keys, 1)
	assert.Equal(t, "GK26b837222cdd03f4c13a05fb", bucket.Keys[0].AccessKeyID)
	assert.True(t, bucket.Keys[0].Permissions.Read)
	require.NotNil(t, bucket.Quotas.MaxSize)
	assert.Equal(t, int64(1000000), *bucket.Quotas.MaxSize)
}

func TestGetBucketInfoNotFound(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(notFoundBucketFixture))
	})

	_, err := c.GetBucketInfo(context.Background(), "xyz")
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.NotFound())
}

func TestListBuckets(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/ListBuckets", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listBucketsResponseFixture))
	})

	buckets, err := c.ListBuckets(context.Background())
	require.NoError(t, err)
	require.Len(t, buckets, 1)
	assert.Equal(t, []string{"my-test-bucket"}, buckets[0].GlobalAliases)
}

func TestUpdateBucket(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/UpdateBucket", r.URL.Path)
		assert.Equal(t, "bucket-id", r.URL.Query().Get("id"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(bucketInfoWithKeyAndWebsiteFixture))
	})

	maxSize := int64(1000000)
	maxObjects := int64(100)
	bucket, err := c.UpdateBucket(context.Background(), "bucket-id", &UpdateBucketInput{
		Website: &WebsiteAccessUpdate{
			Enabled:       true,
			IndexDocument: "index.html",
			ErrorDocument: "error.html",
		},
		Quotas: &BucketQuotas{MaxSize: &maxSize, MaxObjects: &maxObjects},
	})
	require.NoError(t, err)

	websiteAccess, ok := gotBody["websiteAccess"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, websiteAccess["enabled"])
	assert.Equal(t, int64(1000000), *bucket.Quotas.MaxSize)
}

func TestDeleteBucket(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/DeleteBucket", r.URL.Path)
		assert.Equal(t, "bucket-id", r.URL.Query().Get("id"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("null"))
	})

	err := c.DeleteBucket(context.Background(), "bucket-id")
	require.NoError(t, err)
}

func TestAddBucketAlias(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/AddBucketAlias", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(createBucketResponseFixture))
	})

	_, err := c.AddGlobalBucketAlias(context.Background(), "bucket-id", "second-alias")
	require.NoError(t, err)
	assert.Equal(t, "bucket-id", gotBody["bucketId"])
	assert.Equal(t, "second-alias", gotBody["globalAlias"])
}

func TestRemoveBucketAlias(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/RemoveBucketAlias", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(createBucketResponseFixture))
	})

	_, err := c.RemoveGlobalBucketAlias(context.Background(), "bucket-id", "second-alias")
	require.NoError(t, err)
	assert.Equal(t, "second-alias", gotBody["globalAlias"])
}
