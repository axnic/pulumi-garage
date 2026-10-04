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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLifecycleConfigRoundTrip(t *testing.T) {
	t.Parallel()

	expirationDays := 30
	abortDays := 7
	rules := []LifecycleRule{
		{ID: "expire-logs", Prefix: "logs/", Enabled: true, ExpirationDays: &expirationDays},
		{ID: "abort-uploads", Enabled: true, AbortIncompleteMultipartUploadDays: &abortDays},
		{ID: "disabled-rule", Enabled: false, ExpirationDays: &expirationDays},
	}

	cfg := lifecycleConfigFrom(rules)
	require.False(t, cfg.Empty())

	got := lifecycleRulesFrom(cfg)
	require.Len(t, got, 3)

	assert.Equal(t, "expire-logs", got[0].ID)
	assert.Equal(t, "logs/", got[0].Prefix)
	assert.True(t, got[0].Enabled)
	require.NotNil(t, got[0].ExpirationDays)
	assert.Equal(t, 30, *got[0].ExpirationDays)
	assert.Nil(t, got[0].AbortIncompleteMultipartUploadDays)

	assert.Equal(t, "abort-uploads", got[1].ID)
	require.NotNil(t, got[1].AbortIncompleteMultipartUploadDays)
	assert.Equal(t, 7, *got[1].AbortIncompleteMultipartUploadDays)

	assert.False(t, got[2].Enabled)
}

func TestLifecycleConfigFromEmptyIsEmpty(t *testing.T) {
	t.Parallel()

	cfg := lifecycleConfigFrom(nil)
	assert.True(t, cfg.Empty())
	assert.Nil(t, lifecycleRulesFrom(cfg))
}

func TestSetBucketLifecycleWithoutS3Errors(t *testing.T) {
	t.Parallel()

	c := New("http://localhost:3903", "token")
	err := c.SetBucketLifecycle(context.Background(), "my-bucket", []LifecycleRule{{ID: "r1", Enabled: true}})
	assert.ErrorIs(t, err, ErrS3NotConfigured)
}

func TestGetBucketLifecycleWithoutS3Errors(t *testing.T) {
	t.Parallel()

	c := New("http://localhost:3903", "token")
	_, err := c.GetBucketLifecycle(context.Background(), "my-bucket")
	assert.ErrorIs(t, err, ErrS3NotConfigured)
}

func TestNewS3ClientDefaultsRegion(t *testing.T) {
	t.Parallel()

	client, err := NewS3Client("http://localhost:3900", "", "access-key", "secret-key")
	require.NoError(t, err)
	assert.NotNil(t, client)
}
