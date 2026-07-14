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

// Package provider implements a Pulumi native provider for Garage,
// an S3-compatible self-hosted distributed object storage system.
package provider

import (
	"fmt"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

// Version is initialized by the Go linker to contain the semver of this build.
var Version string

// Name controls how this provider is referenced in package names and elsewhere.
const Name string = "garage"

// Provider creates a new instance of the provider.
func Provider() p.Provider {
	p, err := infer.NewProviderBuilder().
		WithDisplayName("pulumi-garage").
		WithDescription("A Pulumi provider to manage Garage S3-compatible object storage.").
		WithHomepage("https://github.com/axnic/pulumi-garage").
		WithRepository("https://github.com/axnic/pulumi-garage").
		WithPublisher("axnic").
		WithLicense("Apache-2.0").
		WithKeywords("garage", "s3", "object-storage", "self-hosted").
		WithNamespace("axnic").
		WithGoImportPath("github.com/axnic/pulumi-garage/sdk/go/pulumi-garage").
		// Lets `pulumi plugin install`/the engine's automatic plugin
		// resolution fetch the provider binary straight from this repo's
		// GitHub Releases - no Pulumi Registry listing required for the
		// provider binary itself to be usable.
		WithPluginDownloadURL("github://api.github.com/axnic/pulumi-garage").
		WithLanguageMap(map[string]any{
			"nodejs": map[string]any{
				"packageName":          "@axnic/pulumi-garage",
				"respectSchemaVersion": true,
			},
			"python": map[string]any{
				"packageName":          "pulumi_garage",
				"respectSchemaVersion": true,
				"pyproject": map[string]any{
					"enabled": true,
				},
			},
			"csharp": map[string]any{
				"rootNamespace":        "Axnic.Pulumi",
				"respectSchemaVersion": true,
			},
			"go": map[string]any{
				"generateResourceContainerTypes": true,
				"importBasePath":                 "github.com/axnic/pulumi-garage/sdk/go/pulumi-garage",
				"respectSchemaVersion":           true,
			},
		}).
		WithResources(
			infer.Resource(Bucket{}),
			infer.Resource(Key{}),
			infer.Resource(BucketKeyPermission{}),
		).
		WithConfig(infer.Config(&Config{})).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{
			"provider": "index",
		}).Build()
	if err != nil {
		panic(fmt.Errorf("unable to build provider: %w", err))
	}
	return p
}
