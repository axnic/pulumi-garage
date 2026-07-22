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
	"errors"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
)

// ErrS3NotConfigured is returned by SetBucketLifecycle/GetBucketLifecycle
// when the provider wasn't given S3 API credentials - lifecycle rules are
// only reachable over Garage's S3 API, never the Admin API.
var ErrS3NotConfigured = errors.New(
	"garage provider: lifecycleRules requires the provider's S3 API config " +
		"(s3Endpoint/accessKeyId/secretAccessKey) to be set",
)

// LifecycleRule is one rule in a bucket's S3 lifecycle configuration:
// automatic object expiration and/or incomplete-multipart-upload cleanup.
type LifecycleRule struct {
	ID                                 string
	Prefix                             string
	Enabled                            bool
	ExpirationDays                     *int
	AbortIncompleteMultipartUploadDays *int
}

// NewS3Client builds the S3 API client used for lifecycle management.
// Region defaults to "garage" (Garage's own default S3 region) if empty.
func NewS3Client(endpoint, region, accessKeyID, secretAccessKey string) (*minio.Client, error) {
	if region == "" {
		region = "garage"
	}
	secure := strings.HasPrefix(endpoint, "https://")
	host := strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")

	return minio.New(host, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:       secure,
		Region:       region,
		BucketLookup: minio.BucketLookupPath,
	})
}

// WithS3 attaches an S3 API client to c, required before SetBucketLifecycle
// or GetBucketLifecycle can be used.
func (c *Client) WithS3(s3Client *minio.Client) *Client {
	c.s3 = s3Client
	return c
}

// SetBucketLifecycle replaces bucketName's lifecycle configuration. An empty
// rules slice removes the lifecycle configuration entirely. bucketName must
// be the bucket's S3-facing name (its global alias), not its Admin API ID.
func (c *Client) SetBucketLifecycle(ctx context.Context, bucketName string, rules []LifecycleRule) error {
	if c.s3 == nil {
		return ErrS3NotConfigured
	}
	return c.s3.SetBucketLifecycle(ctx, bucketName, lifecycleConfigFrom(rules))
}

// GetBucketLifecycle fetches bucketName's current lifecycle configuration.
// Returns (nil, nil) if no lifecycle configuration is set.
func (c *Client) GetBucketLifecycle(ctx context.Context, bucketName string) ([]LifecycleRule, error) {
	if c.s3 == nil {
		return nil, ErrS3NotConfigured
	}
	cfg, err := c.s3.GetBucketLifecycle(ctx, bucketName)
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchLifecycleConfiguration" {
			return nil, nil
		}
		return nil, err
	}
	return lifecycleRulesFrom(cfg), nil
}

func lifecycleConfigFrom(rules []LifecycleRule) *lifecycle.Configuration {
	cfg := lifecycle.NewConfiguration()
	for _, r := range rules {
		status := "Disabled"
		if r.Enabled {
			status = "Enabled"
		}
		rule := lifecycle.Rule{
			ID:         r.ID,
			Status:     status,
			RuleFilter: lifecycle.Filter{Prefix: r.Prefix},
		}
		if r.ExpirationDays != nil {
			rule.Expiration = lifecycle.Expiration{Days: lifecycle.ExpirationDays(*r.ExpirationDays)}
		}
		if r.AbortIncompleteMultipartUploadDays != nil {
			rule.AbortIncompleteMultipartUpload = lifecycle.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: lifecycle.ExpirationDays(*r.AbortIncompleteMultipartUploadDays),
			}
		}
		cfg.Rules = append(cfg.Rules, rule)
	}
	return cfg
}

func lifecycleRulesFrom(cfg *lifecycle.Configuration) []LifecycleRule {
	if cfg.Empty() {
		return nil
	}
	rules := make([]LifecycleRule, 0, len(cfg.Rules))
	for _, r := range cfg.Rules {
		rule := LifecycleRule{
			ID:      r.ID,
			Prefix:  r.RuleFilter.Prefix,
			Enabled: r.Status == "Enabled",
		}
		if !r.Expiration.IsDaysNull() {
			days := int(r.Expiration.Days)
			rule.ExpirationDays = &days
		}
		if !r.AbortIncompleteMultipartUpload.IsDaysNull() {
			days := int(r.AbortIncompleteMultipartUpload.DaysAfterInitiation)
			rule.AbortIncompleteMultipartUploadDays = &days
		}
		rules = append(rules, rule)
	}
	return rules
}
