//go:build go || all
// +build go all

package examples

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoExampleLifecycle drives the go example program's full
// create/read/update/delete lifecycle against a live Garage Admin API, then
// proves the granted BucketKeyPermission actually works by uploading and
// downloading a real object through the S3 API using the generated key -
// not just that the admin API calls succeeded. Needs
// GARAGE_ADMIN_ENDPOINT/GARAGE_ADMIN_TOKEN pointing at a live Garage
// cluster - see `make test_e2e` (Docker) or the repository README for how
// to run one locally.
func TestGoExampleLifecycle(t *testing.T) {
	requireGarage(t)

	cwd, err := os.Getwd()
	require.NoError(t, err)

	module := filepath.Join(cwd, "../sdk/go/pulumi-garage")
	pt := pulumitest.NewPulumiTest(t, "go",
		opttest.GoModReplacement("github.com/axnic/pulumi-garage/sdk/go/pulumi-garage", module),
		opttest.AttachProviderServer("garage", providerFactory),
		opttest.SkipInstall(),
	)

	pt.Preview(t)
	up := pt.Up(t)
	verifyS3Access(t, up.Outputs)
	pt.Destroy(t)
}

// verifyS3Access uses the bucket/key the example program created to
// actually exercise the S3 API - a real PutObject/GetObject round trip -
// confirming the BucketKeyPermission grant is not just recorded but
// functional.
func verifyS3Access(t *testing.T, outputs auto.OutputMap) {
	t.Helper()

	bucketName, accessKeyID, secretAccessKey := stackOutputStrings(t, outputs)

	endpoint := os.Getenv("GARAGE_S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:3900"
	}

	client, err := minio.New(strings.TrimPrefix(endpoint, "http://"), &minio.Options{
		Creds:        credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:       false,
		Region:       "garage",
		BucketLookup: minio.BucketLookupPath,
	})
	require.NoError(t, err)

	ctx := context.Background()
	const key = "hello.txt"
	const content = "hello from the pulumi-garage e2e test"

	_, err = client.PutObject(ctx, bucketName, key, strings.NewReader(content), int64(len(content)),
		minio.PutObjectOptions{ContentType: "text/plain"})
	require.NoError(t, err, "PutObject must succeed with the permissions granted by BucketKeyPermission")

	obj, err := client.GetObject(ctx, bucketName, key, minio.GetObjectOptions{})
	require.NoError(t, err)
	defer obj.Close()

	got, err := io.ReadAll(obj)
	require.NoError(t, err)
	assert.Equal(t, content, string(got))

	require.NoError(t, client.RemoveObject(ctx, bucketName, key, minio.RemoveObjectOptions{}))
}

// stackOutputStrings reads bucketName (the bucket's S3-addressable global
// alias, not its internal Garage ID - S3 bucket names are capped at 63
// characters, shorter than Garage's 64-hex-char bucket IDs), accessKeyId,
// and secretAccessKey from the example program's stack outputs.
func stackOutputStrings(t *testing.T, outputs auto.OutputMap) (bucketName, accessKeyID, secretAccessKey string) {
	t.Helper()
	get := func(name string) string {
		v, ok := outputs[name]
		require.True(t, ok, "missing stack output %q", name)
		s, ok := v.Value.(string)
		require.True(t, ok, "stack output %q is not a string", name)
		return s
	}
	return get("bucketName"), get("accessKeyId"), get("secretAccessKey")
}
