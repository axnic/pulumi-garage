# YAML Example Program

A minimal Pulumi program exercising the three resources this provider
implements: `Key`, `Bucket`, and `BucketKeyPermission`. Used both as a
hand-editable local playground and as the program driven by
`TestYAMLExampleLifecycle` in `examples/yaml_test.go`.

The provider needs a running Garage cluster to talk to. Configure it either
via stack config:

```bash
pulumi config set garage:endpoint http://localhost:3903
pulumi config set garage:adminToken <token> --secret
```

or via the `GARAGE_ADMIN_ENDPOINT` / `GARAGE_ADMIN_TOKEN` environment
variables (see the repository README for how to start a local Garage with
Docker). Then:

```bash
pulumi login
pulumi stack init local
pulumi up
```
