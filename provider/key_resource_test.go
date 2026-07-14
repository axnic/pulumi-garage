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

type fakeKeyAPI struct {
	createKey func(ctx context.Context, name string) (*garageclient.KeyInfo, error)
	getKey    func(ctx context.Context, id string) (*garageclient.KeyInfo, error)
	updateKey func(ctx context.Context, id, name string) (*garageclient.KeyInfo, error)
	deleteKey func(ctx context.Context, id string) error
}

func (f *fakeKeyAPI) CreateKey(ctx context.Context, name string) (*garageclient.KeyInfo, error) {
	return f.createKey(ctx, name)
}

func (f *fakeKeyAPI) GetKeyInfo(ctx context.Context, id string) (*garageclient.KeyInfo, error) {
	return f.getKey(ctx, id)
}

func (f *fakeKeyAPI) UpdateKey(ctx context.Context, id, name string) (*garageclient.KeyInfo, error) {
	return f.updateKey(ctx, id, name)
}

func (f *fakeKeyAPI) DeleteKey(ctx context.Context, id string) error {
	return f.deleteKey(ctx, id)
}

func TestKeyCreate(t *testing.T) {
	t.Parallel()

	var gotName string
	api := &fakeKeyAPI{
		createKey: func(_ context.Context, name string) (*garageclient.KeyInfo, error) {
			gotName = name
			return &garageclient.KeyInfo{
				AccessKeyID:     "GK123",
				Name:            name,
				Created:         "2026-07-14T10:04:46.018Z",
				SecretAccessKey: "supersecret",
			}, nil
		},
	}

	name := "my-key"
	state, id, err := createKey(context.Background(), api, KeyArgs{Name: &name})
	require.NoError(t, err)

	assert.Equal(t, "my-key", gotName)
	assert.Equal(t, "GK123", id)
	assert.Equal(t, "GK123", state.AccessKeyID)
	assert.Equal(t, "supersecret", state.SecretAccessKey)
}

func TestKeyReadPreservesSecretFromPriorState(t *testing.T) {
	t.Parallel()

	api := &fakeKeyAPI{
		getKey: func(_ context.Context, id string) (*garageclient.KeyInfo, error) {
			return &garageclient.KeyInfo{AccessKeyID: id, Name: "renamed", Created: "2026-07-14T10:04:46.018Z"}, nil
		},
	}

	id, state, err := readKey(context.Background(), api, "GK123", "previously-captured-secret")
	require.NoError(t, err)
	assert.Equal(t, "GK123", id)
	assert.Equal(t, "previously-captured-secret", state.SecretAccessKey)
	assert.Equal(t, "renamed", *state.Name)
}

func TestKeyReadNotFoundSignalsDeletion(t *testing.T) {
	t.Parallel()

	api := &fakeKeyAPI{
		getKey: func(context.Context, string) (*garageclient.KeyInfo, error) {
			return nil, &garageclient.APIError{StatusCode: http.StatusNotFound, Code: "NoSuchAccessKey"}
		},
	}

	id, _, err := readKey(context.Background(), api, "gone", "secret")
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestKeyUpdateRenames(t *testing.T) {
	t.Parallel()

	var gotName string
	api := &fakeKeyAPI{
		updateKey: func(_ context.Context, id, name string) (*garageclient.KeyInfo, error) {
			gotName = name
			return &garageclient.KeyInfo{AccessKeyID: id, Name: name, Created: "2026-07-14T10:04:46.018Z"}, nil
		},
	}

	newName := "renamed-key"
	state, err := updateKey(context.Background(), api, "GK123", KeyArgs{Name: &newName}, "existing-secret")
	require.NoError(t, err)
	assert.Equal(t, "renamed-key", gotName)
	assert.Equal(t, "existing-secret", state.SecretAccessKey, "update must not lose the secret captured at create time")
}

func TestKeyDeleteIsIdempotent(t *testing.T) {
	t.Parallel()

	api := &fakeKeyAPI{
		deleteKey: func(context.Context, string) error {
			return &garageclient.APIError{StatusCode: http.StatusNotFound, Code: "NoSuchAccessKey"}
		},
	}

	err := deleteKey(context.Background(), api, "already-gone")
	require.NoError(t, err)
}

var (
	_ infer.CustomResource[KeyArgs, KeyState] = Key{}
	_ infer.CustomRead[KeyArgs, KeyState]     = Key{}
	_ infer.CustomUpdate[KeyArgs, KeyState]   = Key{}
	_ infer.CustomDelete[KeyState]            = Key{}
	_ keyAPI                                  = (*garageclient.Client)(nil)
)
