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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoSetsAuthorizationHeaderAndDecodesResponse(t *testing.T) {
	t.Parallel()

	var gotAuth, gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"hello": "world"})
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")

	var out map[string]string
	err := c.do(context.Background(), http.MethodGet, "/v2/Whatever", nil, nil, &out)
	require.NoError(t, err)

	assert.Equal(t, "Bearer test-token", gotAuth)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/v2/Whatever", gotPath)
	assert.Equal(t, map[string]string{"hello": "world"}, out)
}

func TestDoSendsQueryParams(t *testing.T) {
	t.Parallel()

	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("null"))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")

	err := c.do(context.Background(), http.MethodGet, "/v2/GetBucketInfo", url.Values{"id": {"abc123"}}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "abc123", gotQuery.Get("id"))
}

func TestDoSendsJSONBody(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("null"))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")

	err := c.do(context.Background(), http.MethodPost, "/v2/CreateKey", nil, map[string]string{"name": "my-key"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "my-key", gotBody["name"])
}

func TestDoReturnsAPIErrorOnNon2xx(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"code":    "NoSuchBucket",
			"message": "Bucket not found: abc123",
			"region":  "garage",
			"path":    "/v2/GetBucketInfo",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")

	err := c.do(context.Background(), http.MethodGet, "/v2/GetBucketInfo", nil, nil, nil)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "NoSuchBucket", apiErr.Code)
	assert.Equal(t, "Bucket not found: abc123", apiErr.Message)
	assert.True(t, apiErr.NotFound())
}

func TestAPIErrorNotFoundOnlyTrueFor404(t *testing.T) {
	t.Parallel()

	err := &APIError{StatusCode: http.StatusBadRequest, Code: "InvalidRequest"}
	assert.False(t, err.NotFound())
}
