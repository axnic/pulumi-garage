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
	"fmt"
	"os"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/axnic/pulumi-garage/provider/internal/garageclient"
)

// Environment variables used as a fallback when the corresponding provider
// config field isn't set, mirroring how the `garage` CLI itself reads
// GARAGE_ADMIN_TOKEN.
const (
	envAdminEndpoint = "GARAGE_ADMIN_ENDPOINT"
	envAdminToken    = "GARAGE_ADMIN_TOKEN"
)

// Config is the provider-level configuration: where to find a Garage
// cluster's Admin API, and how to authenticate to it.
type Config struct {
	// Endpoint is the base URL of the Garage Admin API, e.g.
	// "http://localhost:3903". Falls back to GARAGE_ADMIN_ENDPOINT.
	Endpoint string `pulumi:"endpoint,optional"`
	// AdminToken authenticates to the Admin API as a bearer token. Falls
	// back to GARAGE_ADMIN_TOKEN.
	AdminToken string `pulumi:"adminToken,optional" provider:"secret"`

	client *garageclient.Client
}

var _ infer.CustomConfigure = (*Config)(nil)
var _ infer.Annotated = (*Config)(nil)

// Annotate provides schema descriptions for the provider configuration.
func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.Endpoint, "The base URL of the Garage Admin API, e.g. \"http://localhost:3903\". "+
		"Falls back to the "+envAdminEndpoint+" environment variable if not set.")
	a.Describe(&c.AdminToken, "A bearer token authorized against the Garage Admin API. "+
		"Falls back to the "+envAdminToken+" environment variable if not set.")
}

// Configure resolves endpoint/adminToken (applying the environment variable
// fallback) and builds the shared Admin API client used by every resource.
func (c *Config) Configure(_ context.Context) error {
	if c.Endpoint == "" {
		c.Endpoint = os.Getenv(envAdminEndpoint)
	}
	if c.AdminToken == "" {
		c.AdminToken = os.Getenv(envAdminToken)
	}

	if c.Endpoint == "" {
		return fmt.Errorf("garage provider: 'endpoint' must be set, or the %s environment variable", envAdminEndpoint)
	}
	if c.AdminToken == "" {
		return fmt.Errorf("garage provider: 'adminToken' must be set, or the %s environment variable", envAdminToken)
	}

	c.client = garageclient.New(c.Endpoint, c.AdminToken)
	return nil
}
