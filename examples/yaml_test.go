//go:build yaml || all
// +build yaml all

package examples

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestYAMLExampleLifecycle drives the yaml example program's full
// create/read/update/delete lifecycle against a live Garage Admin API. It
// needs GARAGE_ADMIN_ENDPOINT/GARAGE_ADMIN_TOKEN pointing at one - see
// `make test_e2e` (Docker) or the repository README for how to run one
// locally.
func TestYAMLExampleLifecycle(t *testing.T) {
	requireGarage(t)

	pt := pulumitest.NewPulumiTest(t, "yaml",
		opttest.AttachProviderServer("garage", providerFactory),
		opttest.SkipInstall(),
	)

	pt.Preview(t)
	pt.Up(t)
	pt.Destroy(t)
}
