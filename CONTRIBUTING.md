# Contributing

## Developing the provider

### Building the provider

Clone the repository:

```bash
git clone https://github.com/dagster-io/terraform-provider-dagsterplus.git
cd terraform-provider-dagsterplus
```

Build and install the provider binary locally:

```bash
make dev-setup
```

This compiles the binary and writes a `dev.tfrc` file that points Terraform at the local binary instead of the registry.

### Credentials

Copy your Dagster+ credentials into a `.env` file at the repository root:

```
DAGSTER_CLOUD_ORGANIZATION=<your-org>
DAGSTER_CLOUD_API_TOKEN=<your-api-token>
```

The `make dev-*` targets source this file automatically.

### Local development workflow

`integration/main.tf` is a scratch Terraform config used to manually test resources against the real API.

```bash
make dev-plan      # terraform plan against integration/main.tf
make dev-apply     # terraform apply
make dev-destroy   # terraform destroy
make dev-integration  # full cycle: apply → no-drift plan → destroy → no-drift plan
```

### Code formatting

Before submitting a pull request, ensure your Go code is properly formatted:

```bash
gofmt -s -w .
```

### Generating documentation

Documentation is generated from the provider schema and the `examples/` directory:

```bash
make docs
```

### Regenerating GraphQL types

If you modify any `.graphql` file under `internal/client/schema/queries/`, regenerate the Go types:

```bash
make generate
```

Never edit `internal/client/schema/generated.go` by hand — it is overwritten on every `make generate`.

## Testing the provider

### Running unit tests

Unit tests use `net/http/httptest` and require no credentials:

```bash
make test
```

### Running acceptance tests

Acceptance tests run real Terraform operations against the Dagster+ API. Source your credentials first:

```bash
set -a && . ./.env && set +a && TF_ACC=1 go test ./internal/provider/... -v -timeout 120s
```

To run a single test:

```bash
set -a && . ./.env && set +a && TF_ACC=1 go test ./internal/provider/... -run TestAccMyResource -v
```

### Debugging

Enable detailed Terraform logs by setting `TF_LOG`:

```bash
TF_LOG=DEBUG terraform plan
```

Logs appear on `stderr`.

## Adding a new resource

See [CLAUDE.md](CLAUDE.md) for the full step-by-step checklist. The short version:

1. **Client layer** — add `internal/client/<resource>.go` with `Create`/`Get`/`List`/`Update`/`Delete` functions calling `c.doGraphQL`.
2. **Resource** — add `internal/provider/resource_<name>.go` with schema, model struct, and CRUD methods.
3. **Register** — add `NewMyResource` to `Resources()` in `internal/provider/provider.go`.
4. **Unit tests** — add `internal/client/<resource>_test.go` using `httptest`.
5. **Acceptance tests** — add `internal/provider/resource_<name>_test.go`.
6. **Data source** (if applicable) — add `internal/provider/data_source_<name>.go` and register in `DataSources()`.
7. **Examples** — add `examples/resources/dagsterplus_<name>/resource.tf` and `import.sh`.
8. **README** — add a row to the status table.

Before writing any client code, consult `internal/client/schema/schema.graphql` to verify field names, argument types, and available mutations.

## Cutting a release

To cut a new release of the provider, create a new annotated tag and push it. This triggers a GitHub Action that builds binaries for all supported platforms, signs the checksums, and publishes the artifacts as a GitHub Release — which the Terraform Registry picks up automatically.

```bash
git tag -a vX.Y.Z -m vX.Y.Z
git push origin vX.Y.Z
```

The release is created as a draft. Review it on GitHub before publishing so the Terraform Registry picks it up.
