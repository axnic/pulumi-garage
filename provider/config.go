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
	envS3Endpoint    = "GARAGE_S3_ENDPOINT"
	envS3Region      = "GARAGE_S3_REGION"
	envS3AccessKeyID = "GARAGE_S3_ACCESS_KEY_ID"
	envS3SecretKey   = "GARAGE_S3_SECRET_ACCESS_KEY"
)

// Config is the provider-level configuration: where to find a Garage
// cluster's Admin API (and, optionally, its S3 API), and how to
// authenticate to each.
type Config struct {
	// Endpoint is the base URL of the Garage Admin API, e.g.
	// "http://localhost:3903". Falls back to GARAGE_ADMIN_ENDPOINT.
	Endpoint string `pulumi:"endpoint,optional"`
	// AdminToken authenticates to the Admin API as a bearer token. Falls
	// back to GARAGE_ADMIN_TOKEN.
	AdminToken string `pulumi:"adminToken,optional" provider:"secret"`

	// S3Endpoint, AccessKeyID, and SecretAccessKey are only required for
	// features reachable exclusively through Garage's S3 API (currently:
	// Bucket's lifecycleRules). Leave all three unset if you don't use
	// those features.
	S3Endpoint string `pulumi:"s3Endpoint,optional"`
	// S3Region falls back to GARAGE_S3_REGION, then to "garage" (Garage's
	// own default S3 region) if that's unset too.
	S3Region        string `pulumi:"s3Region,optional"`
	AccessKeyID     string `pulumi:"accessKeyId,optional"`
	SecretAccessKey string `pulumi:"secretAccessKey,optional" provider:"secret"`

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
	a.Describe(&c.S3Endpoint, "The base URL of the Garage S3 API, e.g. \"http://localhost:3900\". "+
		"Only required to manage a Bucket's lifecycleRules. Falls back to the "+envS3Endpoint+
		" environment variable if not set.")
	a.Describe(&c.S3Region, "The S3 region to sign requests for. Falls back to the "+envS3Region+
		" environment variable, then to \"garage\" (Garage's own default) if neither is set.")
	a.Describe(&c.AccessKeyID, "An access key ID authorized against the Garage S3 API. "+
		"Only required to manage a Bucket's lifecycleRules. Falls back to the "+envS3AccessKeyID+
		" environment variable if not set.")
	a.Describe(&c.SecretAccessKey, "The secret access key paired with accessKeyId. "+
		"Only required to manage a Bucket's lifecycleRules. Falls back to the "+envS3SecretKey+
		" environment variable if not set.")
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
	if c.S3Endpoint == "" {
		c.S3Endpoint = os.Getenv(envS3Endpoint)
	}
	if c.S3Region == "" {
		c.S3Region = os.Getenv(envS3Region)
	}
	if c.AccessKeyID == "" {
		c.AccessKeyID = os.Getenv(envS3AccessKeyID)
	}
	if c.SecretAccessKey == "" {
		c.SecretAccessKey = os.Getenv(envS3SecretKey)
	}

	if c.Endpoint == "" {
		return fmt.Errorf("garage provider: 'endpoint' must be set, or the %s environment variable", envAdminEndpoint)
	}
	if c.AdminToken == "" {
		return fmt.Errorf("garage provider: 'adminToken' must be set, or the %s environment variable", envAdminToken)
	}

	c.client = garageclient.New(c.Endpoint, c.AdminToken)

	s3Set := c.S3Endpoint != "" || c.AccessKeyID != "" || c.SecretAccessKey != ""
	s3Complete := c.S3Endpoint != "" && c.AccessKeyID != "" && c.SecretAccessKey != ""
	switch {
	case !s3Set:
		// S3 API not configured; a Bucket using lifecycleRules will fail
		// with an actionable error, rather than silently doing nothing.
	case !s3Complete:
		return fmt.Errorf("garage provider: 's3Endpoint', 'accessKeyId', and 'secretAccessKey' must all be set " +
			"together to enable S3 API features (e.g. a Bucket's lifecycleRules)")
	default:
		s3Client, err := garageclient.NewS3Client(c.S3Endpoint, c.S3Region, c.AccessKeyID, c.SecretAccessKey)
		if err != nil {
			return fmt.Errorf("garage provider: building S3 API client: %w", err)
		}
		c.client.WithS3(s3Client)
	}

	return nil
}
