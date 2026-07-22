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

package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/axnic/pulumi-garage/provider/internal/garageclient"
)

type fakeBucketAPI struct {
	createBucket func(ctx context.Context, globalAlias string) (*garageclient.BucketInfo, error)
	getBucket    func(ctx context.Context, id string) (*garageclient.BucketInfo, error)
	updateBucket func(ctx context.Context, id string, in *garageclient.UpdateBucketInput) (*garageclient.BucketInfo, error)
	deleteBucket func(ctx context.Context, id string) error
	addAlias     func(ctx context.Context, bucketID, alias string) (*garageclient.BucketInfo, error)
	removeAlias  func(ctx context.Context, bucketID, alias string) (*garageclient.BucketInfo, error)
	setLifecycle func(ctx context.Context, bucketName string, rules []garageclient.LifecycleRule) error
	getLifecycle func(ctx context.Context, bucketName string) ([]garageclient.LifecycleRule, error)
}

func (f *fakeBucketAPI) CreateBucket(ctx context.Context, globalAlias string) (*garageclient.BucketInfo, error) {
	return f.createBucket(ctx, globalAlias)
}

func (f *fakeBucketAPI) GetBucketInfo(ctx context.Context, id string) (*garageclient.BucketInfo, error) {
	return f.getBucket(ctx, id)
}

func (f *fakeBucketAPI) UpdateBucket(
	ctx context.Context, id string, in *garageclient.UpdateBucketInput,
) (*garageclient.BucketInfo, error) {
	return f.updateBucket(ctx, id, in)
}

func (f *fakeBucketAPI) DeleteBucket(ctx context.Context, id string) error {
	return f.deleteBucket(ctx, id)
}

func (f *fakeBucketAPI) AddGlobalBucketAlias(
	ctx context.Context, bucketID, alias string,
) (*garageclient.BucketInfo, error) {
	return f.addAlias(ctx, bucketID, alias)
}

func (f *fakeBucketAPI) RemoveGlobalBucketAlias(
	ctx context.Context, bucketID, alias string,
) (*garageclient.BucketInfo, error) {
	return f.removeAlias(ctx, bucketID, alias)
}

func (f *fakeBucketAPI) SetBucketLifecycle(
	ctx context.Context, bucketName string, rules []garageclient.LifecycleRule,
) error {
	if f.setLifecycle == nil {
		return garageclient.ErrS3NotConfigured
	}
	return f.setLifecycle(ctx, bucketName, rules)
}

func (f *fakeBucketAPI) GetBucketLifecycle(
	ctx context.Context, bucketName string,
) ([]garageclient.LifecycleRule, error) {
	if f.getLifecycle == nil {
		return nil, garageclient.ErrS3NotConfigured
	}
	return f.getLifecycle(ctx, bucketName)
}

const testBucketAlias = "my-bucket"

func sampleBucketInfo() *garageclient.BucketInfo {
	return &garageclient.BucketInfo{
		ID:            "bucket-id-123",
		Created:       "2026-07-14T10:04:53.512Z",
		GlobalAliases: []string{testBucketAlias},
		WebsiteAccess: false,
		Quotas:        garageclient.BucketQuotas{},
	}
}

func TestBucketCreate(t *testing.T) {
	t.Parallel()

	var gotAlias string
	api := &fakeBucketAPI{
		createBucket: func(_ context.Context, alias string) (*garageclient.BucketInfo, error) {
			gotAlias = alias
			return sampleBucketInfo(), nil
		},
	}

	alias := testBucketAlias
	state, id, err := createBucket(context.Background(), api, BucketArgs{GlobalAlias: &alias})
	require.NoError(t, err)

	assert.Equal(t, testBucketAlias, gotAlias)
	assert.Equal(t, "bucket-id-123", id)
	assert.Equal(t, "2026-07-14T10:04:53.512Z", state.CreatedAt)
}

func TestBucketCreateAppliesWebsiteAndQuotas(t *testing.T) {
	t.Parallel()

	var updateCalled bool
	var gotInput *garageclient.UpdateBucketInput
	api := &fakeBucketAPI{
		createBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
		updateBucket: func(_ context.Context, id string, in *garageclient.UpdateBucketInput) (
			*garageclient.BucketInfo, error,
		) {
			updateCalled = true
			gotInput = in
			assert.Equal(t, "bucket-id-123", id)
			info := sampleBucketInfo()
			info.WebsiteAccess = true
			return info, nil
		},
	}

	maxSize := 1000
	_, _, err := createBucket(context.Background(), api, BucketArgs{
		Website: &WebsiteArgs{IndexDocument: "index.html"},
		Quotas:  &QuotasArgs{MaxSize: &maxSize},
	})
	require.NoError(t, err)

	require.True(t, updateCalled)
	require.NotNil(t, gotInput.Website)
	assert.True(t, gotInput.Website.Enabled)
	assert.Equal(t, "index.html", gotInput.Website.IndexDocument)
	require.NotNil(t, gotInput.Quotas.MaxSize)
	assert.EqualValues(t, 1000, *gotInput.Quotas.MaxSize)
}

func TestBucketReadNotFoundSignalsDeletion(t *testing.T) {
	t.Parallel()

	api := &fakeBucketAPI{
		getBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return nil, &garageclient.APIError{StatusCode: http.StatusNotFound, Code: "NoSuchBucket"}
		},
	}

	id, _, err := readBucket(context.Background(), api, "gone-bucket-id")
	require.NoError(t, err)
	assert.Empty(t, id, "empty ID signals to the engine that the resource is gone")
}

func TestBucketReadReturnsCurrentState(t *testing.T) {
	t.Parallel()

	api := &fakeBucketAPI{
		getBucket: func(_ context.Context, id string) (*garageclient.BucketInfo, error) {
			assert.Equal(t, "bucket-id-123", id)
			return sampleBucketInfo(), nil
		},
	}

	id, state, err := readBucket(context.Background(), api, "bucket-id-123")
	require.NoError(t, err)
	assert.Equal(t, "bucket-id-123", id)
	require.NotNil(t, state.GlobalAlias)
	assert.Equal(t, testBucketAlias, *state.GlobalAlias)
}

func TestBucketUpdateChangesAlias(t *testing.T) {
	t.Parallel()

	var removed, added string
	api := &fakeBucketAPI{
		removeAlias: func(_ context.Context, _, alias string) (*garageclient.BucketInfo, error) {
			removed = alias
			return sampleBucketInfo(), nil
		},
		addAlias: func(_ context.Context, _, alias string) (*garageclient.BucketInfo, error) {
			added = alias
			info := sampleBucketInfo()
			info.GlobalAliases = []string{alias}
			return info, nil
		},
		getBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			info := sampleBucketInfo()
			info.GlobalAliases = []string{"new-alias"}
			return info, nil
		},
		updateBucket: func(context.Context, string, *garageclient.UpdateBucketInput) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
	}

	oldAlias := testBucketAlias
	newAlias := "new-alias"
	state, err := updateBucket(context.Background(), api, "bucket-id-123",
		BucketArgs{GlobalAlias: &oldAlias}, BucketArgs{GlobalAlias: &newAlias})
	require.NoError(t, err)

	assert.Equal(t, testBucketAlias, removed)
	assert.Equal(t, "new-alias", added)
	require.NotNil(t, state.GlobalAlias)
	assert.Equal(t, "new-alias", *state.GlobalAlias)
}

func TestBucketUpdateNoAliasChangeDoesNotTouchAlias(t *testing.T) {
	t.Parallel()

	aliasCalled := false
	api := &fakeBucketAPI{
		removeAlias: func(context.Context, string, string) (*garageclient.BucketInfo, error) {
			aliasCalled = true
			return sampleBucketInfo(), nil
		},
		addAlias: func(context.Context, string, string) (*garageclient.BucketInfo, error) {
			aliasCalled = true
			return sampleBucketInfo(), nil
		},
		updateBucket: func(context.Context, string, *garageclient.UpdateBucketInput) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
		getBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
	}

	alias := testBucketAlias
	website := &WebsiteArgs{IndexDocument: "index.html"}
	_, err := updateBucket(context.Background(), api, "bucket-id-123",
		BucketArgs{GlobalAlias: &alias}, BucketArgs{GlobalAlias: &alias, Website: website})
	require.NoError(t, err)
	assert.False(t, aliasCalled, "alias unchanged, so Add/RemoveGlobalBucketAlias must not be called")
}

func TestBucketDeleteIsIdempotent(t *testing.T) {
	t.Parallel()

	api := &fakeBucketAPI{
		deleteBucket: func(context.Context, string) error {
			return &garageclient.APIError{StatusCode: http.StatusNotFound, Code: "NoSuchBucket"}
		},
	}

	err := deleteBucket(context.Background(), api, "already-gone")
	require.NoError(t, err)
}

func TestBucketCreateWithLifecycleRulesRequiresAlias(t *testing.T) {
	t.Parallel()

	api := &fakeBucketAPI{
		createBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			info := sampleBucketInfo()
			info.GlobalAliases = nil
			return info, nil
		},
	}

	_, _, err := createBucket(context.Background(), api, BucketArgs{
		LifecycleRules: []LifecycleRuleArgs{{ID: "r1", Enabled: true}},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, errLifecycleRulesNeedAlias)
}

func TestBucketCreateAppliesLifecycleRules(t *testing.T) {
	t.Parallel()

	var gotBucketName string
	var gotRules []garageclient.LifecycleRule
	api := &fakeBucketAPI{
		createBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
		setLifecycle: func(_ context.Context, bucketName string, rules []garageclient.LifecycleRule) error {
			gotBucketName = bucketName
			gotRules = rules
			return nil
		},
	}

	alias := testBucketAlias
	days := 30
	_, _, err := createBucket(context.Background(), api, BucketArgs{
		GlobalAlias:    &alias,
		LifecycleRules: []LifecycleRuleArgs{{ID: "expire", Enabled: true, ExpirationDays: &days}},
	})
	require.NoError(t, err)

	assert.Equal(t, testBucketAlias, gotBucketName, "lifecycle rules must be set by alias, not by Admin API ID")
	require.Len(t, gotRules, 1)
	assert.Equal(t, "expire", gotRules[0].ID)
	require.NotNil(t, gotRules[0].ExpirationDays)
	assert.Equal(t, 30, *gotRules[0].ExpirationDays)
}

func TestBucketUpdateAppliesLifecycleRules(t *testing.T) {
	t.Parallel()

	var setCalled bool
	var gotRules []garageclient.LifecycleRule
	api := &fakeBucketAPI{
		updateBucket: func(context.Context, string, *garageclient.UpdateBucketInput) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
		setLifecycle: func(_ context.Context, bucketName string, rules []garageclient.LifecycleRule) error {
			setCalled = true
			assert.Equal(t, testBucketAlias, bucketName)
			gotRules = rules
			return nil
		},
		getBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
	}

	alias := testBucketAlias
	_, err := updateBucket(context.Background(), api, "bucket-id-123",
		BucketArgs{GlobalAlias: &alias, LifecycleRules: []LifecycleRuleArgs{{ID: "r1", Enabled: true}}},
		BucketArgs{GlobalAlias: &alias, LifecycleRules: nil})
	require.NoError(t, err)

	require.True(t, setCalled, "removing every rule must still call SetBucketLifecycle, to delete the config server-side")
	assert.Empty(t, gotRules)
}

func TestBucketUpdateWithoutLifecycleRulesDoesNotCallSetBucketLifecycle(t *testing.T) {
	t.Parallel()

	setCalled := false
	api := &fakeBucketAPI{
		updateBucket: func(context.Context, string, *garageclient.UpdateBucketInput) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
		setLifecycle: func(context.Context, string, []garageclient.LifecycleRule) error {
			setCalled = true
			return nil
		},
		getBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
	}

	alias := testBucketAlias
	_, err := updateBucket(context.Background(), api, "bucket-id-123",
		BucketArgs{GlobalAlias: &alias}, BucketArgs{GlobalAlias: &alias})
	require.NoError(t, err)
	assert.False(t, setCalled)
}

func TestBucketReadPopulatesLifecycleRules(t *testing.T) {
	t.Parallel()

	days := 14
	api := &fakeBucketAPI{
		getBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
		getLifecycle: func(_ context.Context, bucketName string) ([]garageclient.LifecycleRule, error) {
			assert.Equal(t, testBucketAlias, bucketName)
			return []garageclient.LifecycleRule{{ID: "expire", Enabled: true, ExpirationDays: &days}}, nil
		},
	}

	_, state, err := readBucket(context.Background(), api, "bucket-id-123")
	require.NoError(t, err)
	require.Len(t, state.LifecycleRules, 1)
	assert.Equal(t, "expire", state.LifecycleRules[0].ID)
}

func TestBucketReadIgnoresS3NotConfigured(t *testing.T) {
	t.Parallel()

	api := &fakeBucketAPI{
		getBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return sampleBucketInfo(), nil
		},
		// getLifecycle left nil: the fake returns ErrS3NotConfigured.
	}

	_, state, err := readBucket(context.Background(), api, "bucket-id-123")
	require.NoError(t, err)
	assert.Nil(t, state.LifecycleRules)
}

func TestBucketDeletePropagatesOtherErrors(t *testing.T) {
	t.Parallel()

	api := &fakeBucketAPI{
		deleteBucket: func(context.Context, string) error {
			return &garageclient.APIError{StatusCode: http.StatusBadRequest, Code: "BucketNotEmpty", Message: "bucket not empty"}
		},
	}

	err := deleteBucket(context.Background(), api, "non-empty-bucket")
	require.Error(t, err)
}

// Compile-time assertions that Bucket satisfies the infer resource
// interfaces we rely on, and that the real client satisfies bucketAPI.
var (
	_ infer.CustomResource[BucketArgs, BucketState] = Bucket{}
	_ infer.CustomRead[BucketArgs, BucketState]     = Bucket{}
	_ infer.CustomUpdate[BucketArgs, BucketState]   = Bucket{}
	_ infer.CustomDelete[BucketState]               = Bucket{}
	_ bucketAPI                                     = (*garageclient.Client)(nil)
)
