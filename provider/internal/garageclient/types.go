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

// KeyPermissions are the global permissions granted to an access key,
// independent of any particular bucket.
type KeyPermissions struct {
	CreateBucket bool `json:"createBucket"`
}

// KeyInfo is the shape returned by CreateKey, GetKeyInfo, and UpdateKey.
// The "buckets" field the API also returns is intentionally not modelled
// here: this provider treats bucket/key permissions as a separate resource
// (BucketKeyPermission) and never needs to read that list back.
type KeyInfo struct {
	AccessKeyID     string         `json:"accessKeyId"`
	Name            string         `json:"name"`
	Created         string         `json:"created"`
	Expiration      *string        `json:"expiration"`
	Expired         bool           `json:"expired"`
	SecretAccessKey string         `json:"secretAccessKey,omitempty"`
	Permissions     KeyPermissions `json:"permissions"`
}

// KeyListEntry is the (lighter) shape returned by ListKeys. Note the access
// key ID comes back as "id" here, unlike every other endpoint which uses
// "accessKeyId" - this isn't a typo, it's genuinely what the Admin API sends.
type KeyListEntry struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Created    string  `json:"created"`
	Expiration *string `json:"expiration"`
	Expired    bool    `json:"expired"`
}

// BucketKeyPerm is the read/write/owner permission triple a key can hold on
// a bucket. In responses it reflects the current grants. As a request to
// AllowBucketKey/DenyBucketKey, only fields set to true have any effect -
// AllowBucketKey grants exactly the permissions marked true and leaves the
// rest untouched, DenyBucketKey revokes exactly the permissions marked true
// and leaves the rest untouched. Sending false (or omitting a field) is
// always a no-op, verified empirically against a live Garage v2.3.0 node.
type BucketKeyPerm struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
	Owner bool `json:"owner"`
}

// BucketQuotas caps a bucket's total size and/or object count. A nil
// pointer field means "no limit".
type BucketQuotas struct {
	MaxSize    *int64 `json:"maxSize"`
	MaxObjects *int64 `json:"maxObjects"`
}

// BucketWebsiteConfig is the static-website configuration for a bucket, as
// returned (nested under BucketInfo.WebsiteConfig) when websiteAccess is
// enabled. Note this is NOT the shape UpdateBucket expects on the way in -
// see WebsiteAccessUpdate for that.
type BucketWebsiteConfig struct {
	IndexDocument string `json:"indexDocument"`
	ErrorDocument string `json:"errorDocument,omitempty"`
}

// WebsiteAccessUpdate is the request shape UpdateBucket expects for its
// "websiteAccess" field: unlike the response shape (BucketWebsiteConfig),
// it carries an explicit "enabled" flag rather than being nil/non-nil.
type WebsiteAccessUpdate struct {
	Enabled       bool   `json:"enabled"`
	IndexDocument string `json:"indexDocument,omitempty"`
	ErrorDocument string `json:"errorDocument,omitempty"`
}

// BucketInfoKey describes one key's permissions on a bucket, as embedded in
// BucketInfo.Keys.
type BucketInfoKey struct {
	AccessKeyID string        `json:"accessKeyId"`
	Name        string        `json:"name"`
	Permissions BucketKeyPerm `json:"permissions"`
}

// BucketInfo is the shape returned by CreateBucket, GetBucketInfo,
// UpdateBucket, AddBucketAlias, and RemoveBucketAlias.
type BucketInfo struct {
	ID            string               `json:"id"`
	Created       string               `json:"created"`
	GlobalAliases []string             `json:"globalAliases"`
	WebsiteAccess bool                 `json:"websiteAccess"`
	WebsiteConfig *BucketWebsiteConfig `json:"websiteConfig"`
	Keys          []BucketInfoKey      `json:"keys"`
	Objects       int64                `json:"objects"`
	Bytes         int64                `json:"bytes"`
	Quotas        BucketQuotas         `json:"quotas"`
}

// BucketListEntry is the (lighter) shape returned by ListBuckets.
type BucketListEntry struct {
	ID            string   `json:"id"`
	Created       string   `json:"created"`
	GlobalAliases []string `json:"globalAliases"`
	LocalAliases  []string `json:"localAliases"`
}
