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

// bucketAPI is the subset of garageclient.Client the Bucket resource needs.
// Satisfied by *garageclient.Client; tests inject a fake instead.
type bucketAPI interface {
	CreateBucket(ctx context.Context, globalAlias string) (*garageclient.BucketInfo, error)
	GetBucketInfo(ctx context.Context, id string) (*garageclient.BucketInfo, error)
	UpdateBucket(ctx context.Context, id string, in *garageclient.UpdateBucketInput) (*garageclient.BucketInfo, error)
	DeleteBucket(ctx context.Context, id string) error
	AddGlobalBucketAlias(ctx context.Context, bucketID, alias string) (*garageclient.BucketInfo, error)
	RemoveGlobalBucketAlias(ctx context.Context, bucketID, alias string) (*garageclient.BucketInfo, error)
}

// Bucket manages a Garage bucket: its (single) global alias, static-website
// hosting configuration, and storage quotas.
//
// Local aliases (scoped to one access key) and multiple global aliases are
// not supported - only one globalAlias per bucket, matching the common
// case. Deleting a non-empty bucket fails, mirroring Garage's own S3-like
// semantics; empty it first (e.g. with a bucket policy resource in your own
// program, or manually) before removing it from your Pulumi program.
type Bucket struct{}

// QuotasArgs caps a bucket's total size and/or object count. Leave a field
// nil for "no limit".
type QuotasArgs struct {
	MaxSize    *int `pulumi:"maxSize,optional"`
	MaxObjects *int `pulumi:"maxObjects,optional"`
}

var _ infer.Annotated = (*QuotasArgs)(nil)

// Annotate provides schema descriptions for QuotasArgs' fields.
func (q *QuotasArgs) Annotate(a infer.Annotator) {
	a.Describe(&q.MaxSize, "The maximum total size, in bytes, the bucket may hold. Unset means no limit.")
	a.Describe(&q.MaxObjects, "The maximum number of objects the bucket may hold. Unset means no limit.")
}

// WebsiteArgs turns a bucket into a static website.
type WebsiteArgs struct {
	IndexDocument string  `pulumi:"indexDocument"`
	ErrorDocument *string `pulumi:"errorDocument,optional"`
}

var _ infer.Annotated = (*WebsiteArgs)(nil)

// Annotate provides schema descriptions for WebsiteArgs' fields.
func (w *WebsiteArgs) Annotate(a infer.Annotator) {
	a.Describe(&w.IndexDocument, "The document served for requests to the bucket root or any \"directory\".")
	a.Describe(&w.ErrorDocument, "The document served for requests that don't match an existing object. "+
		"Defaults to Garage's built-in error page if unset.")
}

// BucketArgs are the inputs to the Bucket resource.
type BucketArgs struct {
	// GlobalAlias is the bucket's human-readable global alias, e.g.
	// "my-app-data". Buckets can be created without one and addressed only
	// by ID, but an alias is required to use the bucket over the S3 API
	// with most clients.
	GlobalAlias *string      `pulumi:"globalAlias,optional"`
	Website     *WebsiteArgs `pulumi:"website,optional"`
	Quotas      *QuotasArgs  `pulumi:"quotas,optional"`
}

// BucketState is what's persisted in state for a Bucket.
type BucketState struct {
	BucketArgs
	CreatedAt string `pulumi:"createdAt"`
	Objects   int    `pulumi:"objects"`
	Bytes     int    `pulumi:"bytes"`
}

var _ infer.Annotated = (*BucketArgs)(nil)

// Annotate provides schema descriptions for the Bucket resource's fields.
func (a *BucketArgs) Annotate(annotator infer.Annotator) {
	annotator.Describe(&a.GlobalAlias, "The bucket's human-readable global alias, e.g. \"my-app-data\". "+
		"Buckets can be created without one and addressed only by ID, but an alias is required to use the "+
		"bucket over the S3 API with most clients. Only a single global alias is supported; local "+
		"(per-key) aliases and multiple global aliases are not managed by this provider.")
	annotator.Describe(&a.Website, "Static-website hosting configuration for the bucket. Omit to leave "+
		"website hosting disabled.")
	annotator.Describe(&a.Quotas, "Storage quotas for the bucket. Omit either field, or the whole block, "+
		"to leave that limit unset.")
}

var _ infer.Annotated = (*BucketState)(nil)

// Annotate provides schema descriptions for the Bucket resource's outputs.
func (s *BucketState) Annotate(annotator infer.Annotator) {
	annotator.Describe(&s.CreatedAt, "The RFC 3339 timestamp at which the bucket was created.")
	annotator.Describe(&s.Objects, "The number of objects currently stored in the bucket.")
	annotator.Describe(&s.Bytes, "The total size, in bytes, of all objects currently stored in the bucket.")
}

func (Bucket) Create(
	ctx context.Context, req infer.CreateRequest[BucketArgs],
) (infer.CreateResponse[BucketState], error) {
	if req.DryRun {
		return infer.CreateResponse[BucketState]{ID: req.Name, Output: BucketState{BucketArgs: req.Inputs}}, nil
	}
	client := infer.GetConfig[Config](ctx).client
	state, id, err := createBucket(ctx, client, req.Inputs)
	if err != nil {
		return infer.CreateResponse[BucketState]{}, err
	}
	return infer.CreateResponse[BucketState]{ID: id, Output: state}, nil
}

func (Bucket) Read(
	ctx context.Context, req infer.ReadRequest[BucketArgs, BucketState],
) (infer.ReadResponse[BucketArgs, BucketState], error) {
	client := infer.GetConfig[Config](ctx).client
	id, state, err := readBucket(ctx, client, req.ID)
	if err != nil {
		return infer.ReadResponse[BucketArgs, BucketState]{}, err
	}
	if id == "" {
		return infer.ReadResponse[BucketArgs, BucketState]{}, nil
	}
	return infer.ReadResponse[BucketArgs, BucketState]{ID: id, Inputs: state.BucketArgs, State: state}, nil
}

func (Bucket) Update(
	ctx context.Context, req infer.UpdateRequest[BucketArgs, BucketState],
) (infer.UpdateResponse[BucketState], error) {
	if req.DryRun {
		out := BucketState{BucketArgs: req.Inputs, CreatedAt: req.State.CreatedAt}
		return infer.UpdateResponse[BucketState]{Output: out}, nil
	}
	client := infer.GetConfig[Config](ctx).client
	state, err := updateBucket(ctx, client, req.ID, req.State.BucketArgs, req.Inputs)
	if err != nil {
		return infer.UpdateResponse[BucketState]{}, err
	}
	return infer.UpdateResponse[BucketState]{Output: state}, nil
}

func (Bucket) Delete(ctx context.Context, req infer.DeleteRequest[BucketState]) (infer.DeleteResponse, error) {
	client := infer.GetConfig[Config](ctx).client
	return infer.DeleteResponse{}, deleteBucket(ctx, client, req.ID)
}

func createBucket(ctx context.Context, api bucketAPI, args BucketArgs) (BucketState, string, error) {
	alias := ""
	if args.GlobalAlias != nil {
		alias = *args.GlobalAlias
	}
	info, err := api.CreateBucket(ctx, alias)
	if err != nil {
		return BucketState{}, "", err
	}

	if args.Website != nil || args.Quotas != nil {
		info, err = api.UpdateBucket(ctx, info.ID, updateBucketInputFrom(args))
		if err != nil {
			return BucketState{}, "", err
		}
	}

	return bucketStateFrom(info, args), info.ID, nil
}

func readBucket(ctx context.Context, api bucketAPI, id string) (string, BucketState, error) {
	info, err := api.GetBucketInfo(ctx, id)
	if err != nil {
		var apiErr *garageclient.APIError
		if errors.As(err, &apiErr) && apiErr.NotFound() {
			return "", BucketState{}, nil
		}
		return "", BucketState{}, err
	}
	return info.ID, bucketStateFrom(info, argsFromBucketInfo(info)), nil
}

func updateBucket(ctx context.Context, api bucketAPI, id string, oldArgs, newArgs BucketArgs) (BucketState, error) {
	oldAlias, newAlias := "", ""
	if oldArgs.GlobalAlias != nil {
		oldAlias = *oldArgs.GlobalAlias
	}
	if newArgs.GlobalAlias != nil {
		newAlias = *newArgs.GlobalAlias
	}
	if oldAlias != newAlias {
		if oldAlias != "" {
			if _, err := api.RemoveGlobalBucketAlias(ctx, id, oldAlias); err != nil {
				return BucketState{}, err
			}
		}
		if newAlias != "" {
			if _, err := api.AddGlobalBucketAlias(ctx, id, newAlias); err != nil {
				return BucketState{}, err
			}
		}
	}

	if _, err := api.UpdateBucket(ctx, id, updateBucketInputFrom(newArgs)); err != nil {
		return BucketState{}, err
	}

	info, err := api.GetBucketInfo(ctx, id)
	if err != nil {
		return BucketState{}, err
	}
	return bucketStateFrom(info, newArgs), nil
}

func deleteBucket(ctx context.Context, api bucketAPI, id string) error {
	err := api.DeleteBucket(ctx, id)
	var apiErr *garageclient.APIError
	if errors.As(err, &apiErr) && apiErr.NotFound() {
		return nil
	}
	return err
}

// updateBucketInputFrom always sends website/quotas (rather than nil-ing
// them out when unset) so that clearing a previously-set website config or
// quota by removing it from the program takes effect.
func updateBucketInputFrom(args BucketArgs) *garageclient.UpdateBucketInput {
	website := &garageclient.WebsiteAccessUpdate{}
	if args.Website != nil {
		website.Enabled = true
		website.IndexDocument = args.Website.IndexDocument
		if args.Website.ErrorDocument != nil {
			website.ErrorDocument = *args.Website.ErrorDocument
		}
	}

	quotas := &garageclient.BucketQuotas{}
	if args.Quotas != nil {
		if args.Quotas.MaxSize != nil {
			size := int64(*args.Quotas.MaxSize)
			quotas.MaxSize = &size
		}
		if args.Quotas.MaxObjects != nil {
			objects := int64(*args.Quotas.MaxObjects)
			quotas.MaxObjects = &objects
		}
	}

	return &garageclient.UpdateBucketInput{Website: website, Quotas: quotas}
}

func bucketStateFrom(info *garageclient.BucketInfo, args BucketArgs) BucketState {
	return BucketState{
		BucketArgs: args,
		CreatedAt:  info.Created,
		Objects:    int(info.Objects),
		Bytes:      int(info.Bytes),
	}
}

// argsFromBucketInfo normalizes a BucketInfo read back from the API into
// BucketArgs, used by Read to reconcile out-of-band changes.
func argsFromBucketInfo(info *garageclient.BucketInfo) BucketArgs {
	args := BucketArgs{}
	if len(info.GlobalAliases) > 0 {
		alias := info.GlobalAliases[0]
		args.GlobalAlias = &alias
	}
	if info.WebsiteAccess && info.WebsiteConfig != nil {
		website := &WebsiteArgs{IndexDocument: info.WebsiteConfig.IndexDocument}
		if info.WebsiteConfig.ErrorDocument != "" {
			errDoc := info.WebsiteConfig.ErrorDocument
			website.ErrorDocument = &errDoc
		}
		args.Website = website
	}
	if info.Quotas.MaxSize != nil || info.Quotas.MaxObjects != nil {
		quotas := &QuotasArgs{}
		if info.Quotas.MaxSize != nil {
			size := int(*info.Quotas.MaxSize)
			quotas.MaxSize = &size
		}
		if info.Quotas.MaxObjects != nil {
			objects := int(*info.Quotas.MaxObjects)
			quotas.MaxObjects = &objects
		}
		args.Quotas = quotas
	}
	return args
}
