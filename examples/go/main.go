package main

import (
	garage "github.com/axnic/pulumi-garage/sdk/go/pulumi-garage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		key, err := garage.NewKey(ctx, "myKey", &garage.KeyArgs{
			Name: pulumi.String("my-app-key"),
		})
		if err != nil {
			return err
		}

		bucket, err := garage.NewBucket(ctx, "myBucket", &garage.BucketArgs{
			GlobalAlias: pulumi.String("my-app-bucket"),
		})
		if err != nil {
			return err
		}

		_, err = garage.NewBucketKeyPermission(ctx, "myPermission", &garage.BucketKeyPermissionArgs{
			BucketId:    bucket.ID(),
			AccessKeyId: key.AccessKeyId,
			Permissions: garage.PermissionsArgsArgs{
				Read:  pulumi.Bool(true),
				Write: pulumi.Bool(true),
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("bucketId", bucket.ID())
		ctx.Export("bucketName", bucket.GlobalAlias)
		ctx.Export("accessKeyId", key.AccessKeyId)
		ctx.Export("secretAccessKey", key.SecretAccessKey)
		return nil
	})
}
