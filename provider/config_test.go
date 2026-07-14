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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureUsesExplicitFields(t *testing.T) {
	c := &Config{Endpoint: "http://explicit:3903", AdminToken: "explicit-token"}
	err := c.Configure(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "http://explicit:3903", c.Endpoint)
	assert.Equal(t, "explicit-token", c.AdminToken)
	assert.NotNil(t, c.client)
}

func TestConfigureFallsBackToEnv(t *testing.T) {
	t.Setenv("GARAGE_ADMIN_ENDPOINT", "http://from-env:3903")
	t.Setenv("GARAGE_ADMIN_TOKEN", "env-token")

	c := &Config{}
	err := c.Configure(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "http://from-env:3903", c.Endpoint)
	assert.Equal(t, "env-token", c.AdminToken)
	assert.NotNil(t, c.client)
}

func TestConfigureExplicitFieldsTakePrecedenceOverEnv(t *testing.T) {
	t.Setenv("GARAGE_ADMIN_ENDPOINT", "http://from-env:3903")
	t.Setenv("GARAGE_ADMIN_TOKEN", "env-token")

	c := &Config{Endpoint: "http://explicit:3903", AdminToken: "explicit-token"}
	err := c.Configure(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "http://explicit:3903", c.Endpoint)
	assert.Equal(t, "explicit-token", c.AdminToken)
}

func TestConfigureFailsWithoutEndpoint(t *testing.T) {
	// t.Setenv to "" (rather than leaving these ambient) so this passes
	// regardless of the environment it runs in - e.g. the devcontainer sets
	// both for its own always-on Garage instance, which would otherwise
	// mask the "neither set" case this test exists to check.
	t.Setenv(envAdminEndpoint, "")
	t.Setenv(envAdminToken, "")

	c := &Config{AdminToken: "some-token"}
	err := c.Configure(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}

func TestConfigureFailsWithoutAdminToken(t *testing.T) {
	t.Setenv(envAdminEndpoint, "")
	t.Setenv(envAdminToken, "")

	c := &Config{Endpoint: "http://localhost:3903"}
	err := c.Configure(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "adminToken")
}
