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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixtures below are verbatim captures from a real Garage v2.3.0 single-node
// instance (dxflrs/garage:v2.3.0), not hand-guessed from the OpenAPI spec.

const createKeyResponseFixture = `{
	"accessKeyId": "GK26b837222cdd03f4c13a05fb",
	"created": "2026-07-14T10:04:46.018Z",
	"name": "test-key",
	"expiration": null,
	"expired": false,
	"secretAccessKey": "4a3eafc19e6d0bd123dd0ef765048f05d1c1183b0cb4504070efa5d421ce9a87",
	"permissions": {"createBucket": false},
	"buckets": []
}`

const updateKeyResponseFixture = `{
	"accessKeyId": "GK26b837222cdd03f4c13a05fb",
	"created": "2026-07-14T10:04:46.018Z",
	"name": "renamed-key",
	"expiration": null,
	"expired": false,
	"permissions": {"createBucket": false},
	"buckets": []
}`

const listKeysResponseFixture = `[{
	"id": "GK26b837222cdd03f4c13a05fb",
	"name": "test-key",
	"created": "2026-07-14T10:04:46.018Z",
	"expiration": null,
	"expired": false
}]`

func newTestServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(srv.URL, "test-token")
}

func TestCreateKey(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/CreateKey", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(createKeyResponseFixture))
	})

	key, err := c.CreateKey(context.Background(), "test-key")
	require.NoError(t, err)

	assert.Equal(t, "test-key", gotBody["name"])
	assert.Equal(t, "GK26b837222cdd03f4c13a05fb", key.AccessKeyID)
	assert.Equal(t, "4a3eafc19e6d0bd123dd0ef765048f05d1c1183b0cb4504070efa5d421ce9a87", key.SecretAccessKey)
	assert.False(t, key.Expired)
}

func TestGetKeyInfo(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/GetKeyInfo", r.URL.Path)
		assert.Equal(t, "GK26b837222cdd03f4c13a05fb", r.URL.Query().Get("id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(updateKeyResponseFixture))
	})

	key, err := c.GetKeyInfo(context.Background(), "GK26b837222cdd03f4c13a05fb")
	require.NoError(t, err)
	assert.Equal(t, "renamed-key", key.Name)
}

func TestGetKeyInfoNotFound(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "NoSuchAccessKey",
			"message": "Access key not found: nope",
			"region": "garage",
			"path": "/v2/GetKeyInfo"
		}`))
	})

	_, err := c.GetKeyInfo(context.Background(), "nope")
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.NotFound())
}

func TestListKeys(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/ListKeys", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listKeysResponseFixture))
	})

	keys, err := c.ListKeys(context.Background())
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, "GK26b837222cdd03f4c13a05fb", keys[0].ID)
	assert.Equal(t, "test-key", keys[0].Name)
}

func TestUpdateKey(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/UpdateKey", r.URL.Path)
		assert.Equal(t, "GK26b837222cdd03f4c13a05fb", r.URL.Query().Get("id"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(updateKeyResponseFixture))
	})

	key, err := c.UpdateKey(context.Background(), "GK26b837222cdd03f4c13a05fb", "renamed-key")
	require.NoError(t, err)
	assert.Equal(t, "renamed-key", gotBody["name"])
	assert.Equal(t, "renamed-key", key.Name)
}

func TestDeleteKey(t *testing.T) {
	t.Parallel()

	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/DeleteKey", r.URL.Path)
		assert.Equal(t, "GK26b837222cdd03f4c13a05fb", r.URL.Query().Get("id"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("null"))
	})

	err := c.DeleteKey(context.Background(), "GK26b837222cdd03f4c13a05fb")
	require.NoError(t, err)
}
