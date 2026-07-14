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

// keyAPI is the subset of garageclient.Client the Key resource needs.
// Satisfied by *garageclient.Client; tests inject a fake instead.
type keyAPI interface {
	CreateKey(ctx context.Context, name string) (*garageclient.KeyInfo, error)
	GetKeyInfo(ctx context.Context, accessKeyID string) (*garageclient.KeyInfo, error)
	UpdateKey(ctx context.Context, accessKeyID, name string) (*garageclient.KeyInfo, error)
	DeleteKey(ctx context.Context, accessKeyID string) error
}

// Key manages a Garage S3 access key. The secret is only ever readable
// right after creation - Garage's Admin API doesn't return it on
// subsequent reads, so this provider captures it once at Create time and
// carries it forward across Read/Update rather than re-fetching it.
//
// The key's global "createBucket" permission is not modelled in v1; keys
// are scoped to buckets exclusively via the BucketKeyPermission resource.
type Key struct{}

// KeyArgs are the inputs to the Key resource.
type KeyArgs struct {
	// Name is a human-readable label for the key. If unset, Garage assigns
	// a default name.
	Name *string `pulumi:"name,optional"`
}

// KeyState is what's persisted in state for a Key.
type KeyState struct {
	KeyArgs
	AccessKeyID     string `pulumi:"accessKeyId"`
	SecretAccessKey string `pulumi:"secretAccessKey" provider:"secret"`
	CreatedAt       string `pulumi:"createdAt"`
}

var _ infer.Annotated = (*KeyArgs)(nil)

// Annotate provides schema descriptions for the Key resource's fields.
func (a *KeyArgs) Annotate(annotator infer.Annotator) {
	annotator.Describe(&a.Name, "A human-readable label for the key. If unset, Garage assigns a default name.")
}

var _ infer.Annotated = (*KeyState)(nil)

// Annotate provides schema descriptions for the Key resource's outputs.
func (s *KeyState) Annotate(annotator infer.Annotator) {
	annotator.Describe(&s.AccessKeyID, "The S3 access key ID, e.g. the value of AWS_ACCESS_KEY_ID.")
	annotator.Describe(&s.SecretAccessKey, "The S3 secret access key, e.g. the value of AWS_SECRET_ACCESS_KEY. "+
		"Only ever readable at creation time - Garage's Admin API does not return it again afterwards, so it "+
		"is captured once here and carried forward in state.")
	annotator.Describe(&s.CreatedAt, "The RFC 3339 timestamp at which the key was created.")
}

func (Key) Create(ctx context.Context, req infer.CreateRequest[KeyArgs]) (infer.CreateResponse[KeyState], error) {
	if req.DryRun {
		return infer.CreateResponse[KeyState]{ID: req.Name, Output: KeyState{KeyArgs: req.Inputs}}, nil
	}
	client := infer.GetConfig[Config](ctx).client
	state, id, err := createKey(ctx, client, req.Inputs)
	if err != nil {
		return infer.CreateResponse[KeyState]{}, err
	}
	return infer.CreateResponse[KeyState]{ID: id, Output: state}, nil
}

func (Key) Read(
	ctx context.Context, req infer.ReadRequest[KeyArgs, KeyState],
) (infer.ReadResponse[KeyArgs, KeyState], error) {
	client := infer.GetConfig[Config](ctx).client
	id, state, err := readKey(ctx, client, req.ID, req.State.SecretAccessKey)
	if err != nil {
		return infer.ReadResponse[KeyArgs, KeyState]{}, err
	}
	if id == "" {
		return infer.ReadResponse[KeyArgs, KeyState]{}, nil
	}
	return infer.ReadResponse[KeyArgs, KeyState]{ID: id, Inputs: state.KeyArgs, State: state}, nil
}

func (Key) Update(
	ctx context.Context, req infer.UpdateRequest[KeyArgs, KeyState],
) (infer.UpdateResponse[KeyState], error) {
	if req.DryRun {
		out := req.State
		out.KeyArgs = req.Inputs
		return infer.UpdateResponse[KeyState]{Output: out}, nil
	}
	client := infer.GetConfig[Config](ctx).client
	state, err := updateKey(ctx, client, req.ID, req.Inputs, req.State.SecretAccessKey)
	if err != nil {
		return infer.UpdateResponse[KeyState]{}, err
	}
	return infer.UpdateResponse[KeyState]{Output: state}, nil
}

func (Key) Delete(ctx context.Context, req infer.DeleteRequest[KeyState]) (infer.DeleteResponse, error) {
	client := infer.GetConfig[Config](ctx).client
	return infer.DeleteResponse{}, deleteKey(ctx, client, req.ID)
}

func createKey(ctx context.Context, api keyAPI, args KeyArgs) (KeyState, string, error) {
	name := ""
	if args.Name != nil {
		name = *args.Name
	}
	info, err := api.CreateKey(ctx, name)
	if err != nil {
		return KeyState{}, "", err
	}
	return keyStateFrom(info, info.SecretAccessKey), info.AccessKeyID, nil
}

func readKey(ctx context.Context, api keyAPI, id, existingSecret string) (string, KeyState, error) {
	info, err := api.GetKeyInfo(ctx, id)
	if err != nil {
		var apiErr *garageclient.APIError
		if errors.As(err, &apiErr) && apiErr.NotFound() {
			return "", KeyState{}, nil
		}
		return "", KeyState{}, err
	}
	return info.AccessKeyID, keyStateFrom(info, existingSecret), nil
}

func updateKey(ctx context.Context, api keyAPI, id string, args KeyArgs, existingSecret string) (KeyState, error) {
	name := ""
	if args.Name != nil {
		name = *args.Name
	}
	info, err := api.UpdateKey(ctx, id, name)
	if err != nil {
		return KeyState{}, err
	}
	return keyStateFrom(info, existingSecret), nil
}

func deleteKey(ctx context.Context, api keyAPI, id string) error {
	err := api.DeleteKey(ctx, id)
	var apiErr *garageclient.APIError
	if errors.As(err, &apiErr) && apiErr.NotFound() {
		return nil
	}
	return err
}

func keyStateFrom(info *garageclient.KeyInfo, secret string) KeyState {
	name := info.Name
	return KeyState{
		KeyArgs:         KeyArgs{Name: &name},
		AccessKeyID:     info.AccessKeyID,
		SecretAccessKey: secret,
		CreatedAt:       info.Created,
	}
}
