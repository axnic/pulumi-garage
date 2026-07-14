package examples

import (
	"os"
	"testing"

	"github.com/pulumi/providertest/providers"
	goprovider "github.com/pulumi/pulumi-go-provider"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/axnic/pulumi-garage/provider"
)

var providerFactory = func(_ providers.PulumiTest) (pulumirpc.ResourceProviderServer, error) { //nolint:unused
	return goprovider.RawServer("garage", "1.0.0", provider.Provider())(nil)
}

// requireGarage skips the calling test unless GARAGE_ADMIN_ENDPOINT is set,
// keeping `make test` fast/hermetic while still letting `make test_e2e`
// (or a developer with a local Garage running) exercise the full
// create/read/update/delete lifecycle against a real Admin API.
func requireGarage(t *testing.T) { //nolint:unused
	t.Helper()
	if os.Getenv("GARAGE_ADMIN_ENDPOINT") == "" {
		t.Skip("set GARAGE_ADMIN_ENDPOINT (and GARAGE_ADMIN_TOKEN) to run this test against a live Garage instance; " +
			"see `make test_e2e`")
	}
}
