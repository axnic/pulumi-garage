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
	"errors"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/axnic/pulumi-garage/provider/internal/garageclient"
)

// permissionAPI is the subset of garageclient.Client the BucketKeyPermission
// resource needs. Satisfied by *garageclient.Client; tests inject a fake
// instead.
type permissionAPI interface {
	AllowBucketKey(
		ctx context.Context, bucketID, accessKeyID string, perm garageclient.BucketKeyPerm,
	) (*garageclient.BucketInfo, error)
	DenyBucketKey(
		ctx context.Context, bucketID, accessKeyID string, perm garageclient.BucketKeyPerm,
	) (*garageclient.BucketInfo, error)
	GetBucketInfo(ctx context.Context, id string) (*garageclient.BucketInfo, error)
}

// BucketKeyPermission grants an access Key read/write/owner permissions on
// a Bucket. Garage has no single natural ID for this grant, so the
// resource ID is a synthetic "<bucketId>/<accessKeyId>" composite.
type BucketKeyPermission struct{}

// PermissionsArgs is the read/write/owner permission triple to grant.
type PermissionsArgs struct {
	Read  bool `pulumi:"read,optional"`
	Write bool `pulumi:"write,optional"`
	Owner bool `pulumi:"owner,optional"`
}

// BucketKeyPermissionArgs are the inputs to the BucketKeyPermission resource.
type BucketKeyPermissionArgs struct {
	BucketID    string          `pulumi:"bucketId"`
	AccessKeyID string          `pulumi:"accessKeyId"`
	Permissions PermissionsArgs `pulumi:"permissions"`
}

// BucketKeyPermissionState is what's persisted in state for a
// BucketKeyPermission.
type BucketKeyPermissionState struct {
	BucketKeyPermissionArgs
}

var _ infer.Annotated = (*PermissionsArgs)(nil)

// Annotate provides schema descriptions for PermissionsArgs' fields.
func (p *PermissionsArgs) Annotate(a infer.Annotator) {
	a.Describe(&p.Read, "Whether the key can read objects (GetObject, ListObjects, ...) from the bucket.")
	a.Describe(&p.Write, "Whether the key can write objects (PutObject, DeleteObject, ...) to the bucket.")
	a.Describe(&p.Owner, "Whether the key has owner rights on the bucket (manage bucket-level settings "+
		"such as its website configuration or quotas via the S3 API).")
}

var _ infer.Annotated = (*BucketKeyPermissionArgs)(nil)

// Annotate provides schema descriptions for the BucketKeyPermission
// resource's fields.
func (a *BucketKeyPermissionArgs) Annotate(annotator infer.Annotator) {
	annotator.Describe(&a.BucketID, "The ID of the Bucket to grant permissions on.")
	annotator.Describe(&a.AccessKeyID, "The access key ID of the Key to grant permissions to.")
	annotator.Describe(&a.Permissions, "The read/write/owner permissions to grant.")
}

func (BucketKeyPermission) Create(
	ctx context.Context, req infer.CreateRequest[BucketKeyPermissionArgs],
) (infer.CreateResponse[BucketKeyPermissionState], error) {
	if req.DryRun {
		id := bucketKeyPermissionID(req.Inputs.BucketID, req.Inputs.AccessKeyID)
		return infer.CreateResponse[BucketKeyPermissionState]{
			ID:     id,
			Output: BucketKeyPermissionState{BucketKeyPermissionArgs: req.Inputs},
		}, nil
	}
	client := infer.GetConfig[Config](ctx).client
	state, id, err := createBucketKeyPermission(ctx, client, req.Inputs)
	if err != nil {
		return infer.CreateResponse[BucketKeyPermissionState]{}, err
	}
	return infer.CreateResponse[BucketKeyPermissionState]{ID: id, Output: state}, nil
}

func (BucketKeyPermission) Read(
	ctx context.Context, req infer.ReadRequest[BucketKeyPermissionArgs, BucketKeyPermissionState],
) (infer.ReadResponse[BucketKeyPermissionArgs, BucketKeyPermissionState], error) {
	client := infer.GetConfig[Config](ctx).client
	id, state, err := readBucketKeyPermission(ctx, client, req.Inputs.BucketID, req.Inputs.AccessKeyID)
	if err != nil {
		return infer.ReadResponse[BucketKeyPermissionArgs, BucketKeyPermissionState]{}, err
	}
	if id == "" {
		return infer.ReadResponse[BucketKeyPermissionArgs, BucketKeyPermissionState]{}, nil
	}
	return infer.ReadResponse[BucketKeyPermissionArgs, BucketKeyPermissionState]{
		ID: id, Inputs: state.BucketKeyPermissionArgs, State: state,
	}, nil
}

func (BucketKeyPermission) Update(
	ctx context.Context, req infer.UpdateRequest[BucketKeyPermissionArgs, BucketKeyPermissionState],
) (infer.UpdateResponse[BucketKeyPermissionState], error) {
	if req.DryRun {
		return infer.UpdateResponse[BucketKeyPermissionState]{
			Output: BucketKeyPermissionState{BucketKeyPermissionArgs: req.Inputs},
		}, nil
	}
	client := infer.GetConfig[Config](ctx).client
	state, err := updateBucketKeyPermission(ctx, client, req.State.BucketKeyPermissionArgs, req.Inputs)
	if err != nil {
		return infer.UpdateResponse[BucketKeyPermissionState]{}, err
	}
	return infer.UpdateResponse[BucketKeyPermissionState]{Output: state}, nil
}

func (BucketKeyPermission) Delete(
	ctx context.Context, req infer.DeleteRequest[BucketKeyPermissionState],
) (infer.DeleteResponse, error) {
	client := infer.GetConfig[Config](ctx).client
	return infer.DeleteResponse{}, deleteBucketKeyPermission(ctx, client, req.State.BucketKeyPermissionArgs)
}

func bucketKeyPermissionID(bucketID, accessKeyID string) string {
	return bucketID + "/" + accessKeyID
}

func createBucketKeyPermission(
	ctx context.Context, api permissionAPI, args BucketKeyPermissionArgs,
) (BucketKeyPermissionState, string, error) {
	perm := toBucketKeyPerm(args.Permissions)
	if _, err := api.AllowBucketKey(ctx, args.BucketID, args.AccessKeyID, perm); err != nil {
		return BucketKeyPermissionState{}, "", err
	}
	id := bucketKeyPermissionID(args.BucketID, args.AccessKeyID)
	return BucketKeyPermissionState{BucketKeyPermissionArgs: args}, id, nil
}

func readBucketKeyPermission(
	ctx context.Context, api permissionAPI, bucketID, accessKeyID string,
) (string, BucketKeyPermissionState, error) {
	info, err := api.GetBucketInfo(ctx, bucketID)
	if err != nil {
		var apiErr *garageclient.APIError
		if errors.As(err, &apiErr) && apiErr.NotFound() {
			return "", BucketKeyPermissionState{}, nil
		}
		return "", BucketKeyPermissionState{}, err
	}

	for _, key := range info.Keys {
		if key.AccessKeyID != accessKeyID {
			continue
		}
		args := BucketKeyPermissionArgs{
			BucketID:    bucketID,
			AccessKeyID: accessKeyID,
			Permissions: fromBucketKeyPerm(key.Permissions),
		}
		return bucketKeyPermissionID(bucketID, accessKeyID), BucketKeyPermissionState{BucketKeyPermissionArgs: args}, nil
	}
	// The key holds no permissions on this bucket (anymore): the grant is gone.
	return "", BucketKeyPermissionState{}, nil
}

func updateBucketKeyPermission(
	ctx context.Context, api permissionAPI, oldArgs, newArgs BucketKeyPermissionArgs,
) (BucketKeyPermissionState, error) {
	toAllow := garageclient.BucketKeyPerm{
		Read:  newArgs.Permissions.Read && !oldArgs.Permissions.Read,
		Write: newArgs.Permissions.Write && !oldArgs.Permissions.Write,
		Owner: newArgs.Permissions.Owner && !oldArgs.Permissions.Owner,
	}
	toDeny := garageclient.BucketKeyPerm{
		Read:  !newArgs.Permissions.Read && oldArgs.Permissions.Read,
		Write: !newArgs.Permissions.Write && oldArgs.Permissions.Write,
		Owner: !newArgs.Permissions.Owner && oldArgs.Permissions.Owner,
	}

	if toAllow.Read || toAllow.Write || toAllow.Owner {
		if _, err := api.AllowBucketKey(ctx, newArgs.BucketID, newArgs.AccessKeyID, toAllow); err != nil {
			return BucketKeyPermissionState{}, err
		}
	}
	if toDeny.Read || toDeny.Write || toDeny.Owner {
		if _, err := api.DenyBucketKey(ctx, newArgs.BucketID, newArgs.AccessKeyID, toDeny); err != nil {
			return BucketKeyPermissionState{}, err
		}
	}

	return BucketKeyPermissionState{BucketKeyPermissionArgs: newArgs}, nil
}

func deleteBucketKeyPermission(ctx context.Context, api permissionAPI, args BucketKeyPermissionArgs) error {
	_, err := api.DenyBucketKey(ctx, args.BucketID, args.AccessKeyID, toBucketKeyPerm(args.Permissions))
	var apiErr *garageclient.APIError
	if errors.As(err, &apiErr) && apiErr.NotFound() {
		return nil
	}
	return err
}

func toBucketKeyPerm(p PermissionsArgs) garageclient.BucketKeyPerm {
	return garageclient.BucketKeyPerm{Read: p.Read, Write: p.Write, Owner: p.Owner}
}

func fromBucketKeyPerm(p garageclient.BucketKeyPerm) PermissionsArgs {
	return PermissionsArgs{Read: p.Read, Write: p.Write, Owner: p.Owner}
}
