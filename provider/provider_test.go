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
	"testing"

	"github.com/stretchr/testify/require"
)

// TestProviderBuilds is a smoke test for the infer.NewProviderBuilder chain
// in Provider(): it panics on a misconfigured builder (e.g. a resource that
// doesn't satisfy the interfaces its Annotate/Configure wiring implies), so
// simply calling it exercises that wiring end to end.
func TestProviderBuilds(t *testing.T) {
	t.Parallel()

	prov := Provider()
	require.NotNil(t, prov)
}
