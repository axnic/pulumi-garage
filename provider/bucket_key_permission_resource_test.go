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

type grantFunc func(
	ctx context.Context, bucketID, accessKeyID string, perm garageclient.BucketKeyPerm,
) (*garageclient.BucketInfo, error)

type fakePermissionAPI struct {
	allow     grantFunc
	deny      grantFunc
	getBucket func(ctx context.Context, id string) (*garageclient.BucketInfo, error)
}

func (f *fakePermissionAPI) AllowBucketKey(
	ctx context.Context, bucketID, accessKeyID string, perm garageclient.BucketKeyPerm,
) (*garageclient.BucketInfo, error) {
	return f.allow(ctx, bucketID, accessKeyID, perm)
}

func (f *fakePermissionAPI) DenyBucketKey(
	ctx context.Context, bucketID, accessKeyID string, perm garageclient.BucketKeyPerm,
) (*garageclient.BucketInfo, error) {
	return f.deny(ctx, bucketID, accessKeyID, perm)
}

func (f *fakePermissionAPI) GetBucketInfo(ctx context.Context, id string) (*garageclient.BucketInfo, error) {
	return f.getBucket(ctx, id)
}

func TestBucketKeyPermissionCreateGrantsRequestedPermissions(t *testing.T) {
	t.Parallel()

	var gotPerm garageclient.BucketKeyPerm
	api := &fakePermissionAPI{
		allow: func(_ context.Context, bucketID, accessKeyID string, perm garageclient.BucketKeyPerm) (
			*garageclient.BucketInfo, error,
		) {
			assert.Equal(t, "bucket-id", bucketID)
			assert.Equal(t, "GK123", accessKeyID)
			gotPerm = perm
			return &garageclient.BucketInfo{ID: bucketID}, nil
		},
	}

	args := BucketKeyPermissionArgs{
		BucketID:    "bucket-id",
		AccessKeyID: "GK123",
		Permissions: PermissionsArgs{Read: true, Write: true},
	}
	state, id, err := createBucketKeyPermission(context.Background(), api, args)
	require.NoError(t, err)

	assert.Equal(t, "bucket-id/GK123", id)
	assert.True(t, gotPerm.Read)
	assert.True(t, gotPerm.Write)
	assert.False(t, gotPerm.Owner)
	assert.True(t, state.Permissions.Read)
}

func TestBucketKeyPermissionReadFindsCurrentGrant(t *testing.T) {
	t.Parallel()

	api := &fakePermissionAPI{
		getBucket: func(_ context.Context, id string) (*garageclient.BucketInfo, error) {
			assert.Equal(t, "bucket-id", id)
			return &garageclient.BucketInfo{
				ID: "bucket-id",
				Keys: []garageclient.BucketInfoKey{
					{AccessKeyID: "GK123", Permissions: garageclient.BucketKeyPerm{Read: true, Write: true}},
				},
			}, nil
		},
	}

	id, state, err := readBucketKeyPermission(context.Background(), api, "bucket-id", "GK123")
	require.NoError(t, err)
	assert.Equal(t, "bucket-id/GK123", id)
	assert.True(t, state.Permissions.Read)
	assert.True(t, state.Permissions.Write)
	assert.False(t, state.Permissions.Owner)
}

func TestBucketKeyPermissionReadMissingBucketSignalsDeletion(t *testing.T) {
	t.Parallel()

	api := &fakePermissionAPI{
		getBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return nil, &garageclient.APIError{StatusCode: http.StatusNotFound, Code: "NoSuchBucket"}
		},
	}

	id, _, err := readBucketKeyPermission(context.Background(), api, "gone", "GK123")
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestBucketKeyPermissionReadMissingGrantSignalsDeletion(t *testing.T) {
	t.Parallel()

	api := &fakePermissionAPI{
		getBucket: func(context.Context, string) (*garageclient.BucketInfo, error) {
			return &garageclient.BucketInfo{ID: "bucket-id", Keys: []garageclient.BucketInfoKey{}}, nil
		},
	}

	id, _, err := readBucketKeyPermission(context.Background(), api, "bucket-id", "GK123")
	require.NoError(t, err)
	assert.Empty(t, id, "key no longer has any grant on this bucket")
}

func TestBucketKeyPermissionUpdateAllowsNewlyTrueAndDeniesNewlyFalse(t *testing.T) {
	t.Parallel()

	var allowed, denied garageclient.BucketKeyPerm
	var allowCalled, denyCalled bool
	api := &fakePermissionAPI{
		allow: func(_ context.Context, _, _ string, perm garageclient.BucketKeyPerm) (*garageclient.BucketInfo, error) {
			allowCalled = true
			allowed = perm
			return &garageclient.BucketInfo{ID: "bucket-id"}, nil
		},
		deny: func(_ context.Context, _, _ string, perm garageclient.BucketKeyPerm) (*garageclient.BucketInfo, error) {
			denyCalled = true
			denied = perm
			return &garageclient.BucketInfo{ID: "bucket-id"}, nil
		},
	}

	oldArgs := BucketKeyPermissionArgs{
		BucketID: "bucket-id", AccessKeyID: "GK123",
		Permissions: PermissionsArgs{Read: true, Write: false, Owner: true},
	}
	newArgs := BucketKeyPermissionArgs{
		BucketID: "bucket-id", AccessKeyID: "GK123",
		Permissions: PermissionsArgs{Read: true, Write: true, Owner: false},
	}
	state, err := updateBucketKeyPermission(context.Background(), api, oldArgs, newArgs)
	require.NoError(t, err)

	require.True(t, allowCalled)
	assert.True(t, allowed.Write, "write went false->true, must be allowed")
	assert.False(t, allowed.Read, "read was unchanged, must not be re-allowed")

	require.True(t, denyCalled)
	assert.True(t, denied.Owner, "owner went true->false, must be denied")
	assert.False(t, denied.Write, "write was not revoked, must not be denied")

	assert.True(t, state.Permissions.Write)
	assert.False(t, state.Permissions.Owner)
}

func TestBucketKeyPermissionUpdateNoOpWhenUnchanged(t *testing.T) {
	t.Parallel()

	called := false
	api := &fakePermissionAPI{
		allow: func(context.Context, string, string, garageclient.BucketKeyPerm) (*garageclient.BucketInfo, error) {
			called = true
			return &garageclient.BucketInfo{}, nil
		},
		deny: func(context.Context, string, string, garageclient.BucketKeyPerm) (*garageclient.BucketInfo, error) {
			called = true
			return &garageclient.BucketInfo{}, nil
		},
	}

	args := BucketKeyPermissionArgs{
		BucketID: "bucket-id", AccessKeyID: "GK123",
		Permissions: PermissionsArgs{Read: true},
	}
	_, err := updateBucketKeyPermission(context.Background(), api, args, args)
	require.NoError(t, err)
	assert.False(t, called, "no permission changed, so neither Allow nor Deny should be called")
}

func TestBucketKeyPermissionDeleteRevokesAllCurrentPermissions(t *testing.T) {
	t.Parallel()

	var gotPerm garageclient.BucketKeyPerm
	api := &fakePermissionAPI{
		deny: func(_ context.Context, bucketID, accessKeyID string, perm garageclient.BucketKeyPerm) (
			*garageclient.BucketInfo, error,
		) {
			assert.Equal(t, "bucket-id", bucketID)
			assert.Equal(t, "GK123", accessKeyID)
			gotPerm = perm
			return &garageclient.BucketInfo{}, nil
		},
	}

	args := BucketKeyPermissionArgs{
		BucketID: "bucket-id", AccessKeyID: "GK123",
		Permissions: PermissionsArgs{Read: true, Write: true, Owner: true},
	}
	err := deleteBucketKeyPermission(context.Background(), api, args)
	require.NoError(t, err)
	assert.True(t, gotPerm.Read)
	assert.True(t, gotPerm.Write)
	assert.True(t, gotPerm.Owner)
}

func TestBucketKeyPermissionDeleteIsIdempotentWhenBucketGone(t *testing.T) {
	t.Parallel()

	api := &fakePermissionAPI{
		deny: func(context.Context, string, string, garageclient.BucketKeyPerm) (*garageclient.BucketInfo, error) {
			return nil, &garageclient.APIError{StatusCode: http.StatusNotFound, Code: "NoSuchBucket"}
		},
	}

	args := BucketKeyPermissionArgs{BucketID: "gone", AccessKeyID: "GK123", Permissions: PermissionsArgs{Read: true}}
	err := deleteBucketKeyPermission(context.Background(), api, args)
	require.NoError(t, err)
}

var (
	_ infer.CustomResource[BucketKeyPermissionArgs, BucketKeyPermissionState] = BucketKeyPermission{}
	_ infer.CustomRead[BucketKeyPermissionArgs, BucketKeyPermissionState]     = BucketKeyPermission{}
	_ infer.CustomUpdate[BucketKeyPermissionArgs, BucketKeyPermissionState]   = BucketKeyPermission{}
	_ infer.CustomDelete[BucketKeyPermissionState]                            = BucketKeyPermission{}
	_ permissionAPI                                                           = (*garageclient.Client)(nil)
)
