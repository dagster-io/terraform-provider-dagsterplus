# Contributing a New Resource or Data Source

This file is instructions for Claude. Follow every section below whenever adding or updating a resource or data source in this provider.

---

## Environment

```bash
export PATH="$PATH:/usr/local/go/bin"   # Go is at /usr/local/go/bin, not in default PATH
go build ./...                           # must pass before any PR
go vet ./internal/provider/...           # must produce no new errors
```

All existing tests should pass before and after your changes. Run `go test ./internal/client/...` to verify.

### Credentials

Provider credentials are stored in `.env` at the repository root:

```
DAGSTER_CLOUD_ORGANIZATION=<org>
DAGSTER_CLOUD_API_TOKEN=<token>
```

The `make dev-*` targets source this file automatically. To use the credentials in other commands (e.g. acceptance tests), source the file first:

```bash
set -a && . ./.env && set +a
```

Or for a one-liner with any make/go command:

```bash
set -a && . ./.env && set +a && TF_ACC=1 go test ./internal/provider/... -run TestAccMyResource -v
```

### Local development workflow

`integration/main.tf` is a scratch Terraform config used to manually test resources against the real API. The provider binary is loaded from the repo root via `dev.tfrc` (a local dev override).

```bash
make dev-setup         # build binary + write dev.tfrc (run once, or after binary changes)
make dev-plan          # terraform plan against integration/main.tf (sources .env automatically)
make dev-apply         # terraform apply (sources .env automatically)
make dev-destroy       # terraform destroy (sources .env automatically)
make dev-integration   # apply → no-drift plan → destroy → no-drift plan (full integration cycle)
```

When adding a new resource, add a representative example block to `integration/main.tf` so it can be exercised with `make dev-plan` before writing acceptance tests.

---

## Breaking Change Protocol

**Applies automatically whenever you modify `internal/provider/resource_*.go`, `internal/provider/data_source_*.go`, or any file under `internal/client/schema/`.**

Before writing or finalising any change to those files, check your diff against this list. If **any** of the following are true, **stop immediately and ask the user to confirm before proceeding** — do not complete the change first:

| Change | Why it breaks production users |
|--------|-------------------------------|
| Removing an attribute from a schema | Users with that attribute in `.tf` config get a plan-time error |
| Renaming an attribute | Equivalent to removal from Terraform's perspective |
| Changing `Optional` → `Required` | Configs that omit the field now fail |
| Adding a new `Required` attribute without a default | Existing configs that don't supply it break |
| Adding `RequiresReplace` to an existing attribute | Resources that were stable now silently destroy/recreate |
| Removing a value from `stringvalidator.OneOf` | Configs using that value now fail validation |
| Changing an attribute's type (e.g. `StringAttribute` → `ListAttribute`) | State is incompatible; `terraform refresh` corrupts state |
| Changing a default value that alters behavior | Silent behavior change for users relying on the default |
| Changing the import ID format (field order, delimiter, or fields) | Existing `terraform import` scripts break |

When you stop, state clearly:
1. Which change triggered the alert and in which file/attribute
2. Why it is a breaking change for production users
3. Ask the user whether to proceed or find a non-breaking alternative

---

## Checklist: Adding a New Resource

Work through these in order. Do not skip steps.

### 1. Client layer (`internal/client/<resource>.go`)

- Define Go structs for the API response shape (raw GraphQL JSON) and the domain type the provider works with.
- Implement `Create` / `Get` / `List` / `Update` / `Delete` functions that call `c.doGraphQL(ctx, deployment, query, vars, &result)`.
  - Org-scoped calls: pass `""` as the deployment argument.
  - Deployment-scoped calls: pass the deployment name.
- Map API-specific values (e.g. `__typename`, ALL_CAPS enums) to user-friendly strings in the domain type; keep the mapping tables (`xToAPI` / `apiToX`) close to where they are used.
- Return `fmt.Errorf("<FuncName>: %w", err)` so errors carry call-site context.

### 2. Provider model and schema (`internal/provider/resource_<name>.go`)

#### Struct tags
- Every field in the model struct must have a `tfsdk:"<snake_case_name>"` tag that exactly matches the schema attribute/block name.

#### Schema attribute rules
| Situation | Setting |
|-----------|---------|
| User supplies on create/update | `Required: true` or `Optional: true` |
| Derived by provider / returned by API only | `Computed: true` only — **never** `Optional+Computed` unless the user can legitimately override it |
| Immutable (forces replacement) | `PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}` |
| Stable after creation (e.g. ID) | `PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}` |
| Secrets (tokens, keys, URLs) | `Sensitive: true` |

#### Validators — always use `terraform-plugin-framework-validators`
Add validators at the schema layer so errors surface at `terraform plan` before any API call.

```go
import (
    "github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
    "github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
    "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
)
```

Common patterns:
- Enum strings: `stringvalidator.OneOf("value_a", "value_b")`
- Mutually exclusive sibling attributes inside a nested block: `stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("other_field"))`
- Numeric bounds: `int64validator.AtLeast(1)`, `int64validator.Between(1, 100)`
- Validators are not needed on data source schemas (all attributes are `Computed`).

#### CRUD methods
Implement all four plus import:
- `Create` — read plan, call client, write state via `policyToModel` equivalent.
- `Read` — read state ID, call client Get, refresh state.
- `Update` — read plan, call client update/upsert, write state.
- `Delete` — read state, call client delete.
- `ImportState` — split `req.ID` on `/`, call client Get, write state. Include any fields that the API does not return (e.g. a `policy_type` discriminator) in `ImportStateVerifyIgnore` in acceptance tests.

Always check `resp.Diagnostics.HasError()` after appending diagnostics before proceeding.

### 3. Register in `internal/provider/provider.go`

Add to `Resources()`:
```go
NewMyResource,
```

### 4. Unit tests (`internal/client/<resource>_test.go`)

- Use `net/http/httptest` — no live credentials required.
- Cover: Create, Read (found), Read (not found → error), Update, Delete, and any error paths (HTTP errors, GraphQL errors, unexpected `__typename`).
- Follow the pattern in `internal/client/alert_policies_test.go` (or similar existing test file).
- Run with `go test ./internal/client/... -run TestMyResource -v`.

### 5. Acceptance tests (`internal/provider/resource_<name>_test.go`)

- Use `resource.Test` from `terraform-plugin-testing`.
- Every test must:
  - Call `testAccPreCheck(t)` in `PreCheck`.
  - Use `testAccProtoV6ProviderFactories`.
  - Set `CheckDestroy` to verify the remote resource no longer exists after destroy.
  - Name test resources with prefix `acc-tf-` to distinguish them from real resources and avoid collisions.
- Include steps for: Create, at least one meaningful Update (changing a significant attribute), and Import.
- `ImportStateVerifyIgnore` — list any attributes the API does not return (so they cannot round-trip through import).
- Run with `TF_ACC=1 DAGSTER_CLOUD_ORGANIZATION=... DAGSTER_CLOUD_API_TOKEN=... go test ./internal/provider/... -run TestAccMyResource -v`.

### 6. Data source — add when it makes sense

A data source is appropriate when users may need to reference an existing resource that was created outside Terraform (e.g. a deployment, an alert policy).

- Create `internal/provider/data_source_<name>.go`.
- Schema: lookup keys (`name`, `deployment`, etc.) are `Required`; everything else is `Computed`.
- Do not add validators to data source schemas.
- Reuse the resource's model struct and `*ToModel` helper if they are in the same package.
- Register in `provider.go` `DataSources()`.
- Add an acceptance test in `internal/provider/data_source_<name>_test.go` that creates a resource and then reads it back via the data source.
- Add an example at `examples/data-sources/dagsterplus_<name>/data-source.tf` (filename must be `data-source.tf`, not `main.tf` — `tfplugindocs` uses this for the "Example Usage" section).

### 7. Examples (`examples/resources/dagsterplus_<name>/`)

Create two files in the resource examples directory:

**`resource.tf`** — the main usage example (`tfplugindocs` picks this up for "Example Usage"):
- Include at least two examples if the resource has meaningfully different configurations (e.g. different policy types, different notification channels).
- Use realistic but obviously-fake values (`"my-org"`, `"oncall@example.com"`, `"#alerts"`).

**`import.sh`** — shell snippet showing how to import the resource (`tfplugindocs` picks this up for the "Import" section):
- Show the exact `terraform import` command with the ID format.
- Comment any attributes that are not returned by the API and cannot round-trip through import (e.g. `policy_type`, `token`).

Example structure:
```
examples/resources/dagsterplus_<name>/
  resource.tf
  import.sh
```

### 8. README (`README.md`)

Add a row to the **Status table** for the new resource/data source. The table has five columns:

| Entity | Resource | Data Source | Imports | Status |
|--------|----------|-------------|---------|--------|

- **Entity**: short human name (e.g. `Alert Policy`, `Deployment`).
- **Resource**: the full resource type name (e.g. `dagsterplus_alert_policy`), or `—` if none.
- **Data Source**: the full data source type name, or `—` if none.
- **Imports**: `Yes` if the resource implements `ImportState`, otherwise `No`.
- **Status**: `**Ready**` only when GraphQL API behaviour has been verified against the live API; otherwise `Experimental`.

Resources and data sources are self-documenting via the Terraform registry (generated from schema descriptions and `examples/` files by `tfplugindocs`) — do not add reference tables or usage snippets to the README.

### 9. CHANGELOG (`CHANGELOG.md`)

Update the `## [Unreleased]` section for every change made. Use the appropriate heading:

- `### Added` — new resources, data sources, attributes, or features
- `### Fixed` — bug fixes
- `### Changed` — backward-compatible behavior or implementation changes
- `### Breaking Changes` — anything flagged by the Breaking Change Protocol above

Breaking changes **must** use `### Breaking Changes` (not `### Changed`). Each bullet must name the resource, the attribute, and what the user needs to do to migrate.

### 10. Final checks

These mirror the CI jobs that run on every PR. All must pass before merging.

```bash
go build ./...                                   # CI: Build
go vet ./...                                     # CI: Vet
gofmt -l . && [ -z "$(gofmt -l .)" ]            # CI: Format — output must be empty
go test ./internal/client/... ./internal/resources/... ./internal/datasources/... -v -timeout 120s  # CI: Unit Tests
go mod tidy && git diff --exit-code go.mod go.sum          # CI: Module Tidy
make generate && git diff --exit-code internal/client/schema/generated.go  # CI: Generated Code
make dev-plan                                    # smoke-test against the real API using integration/main.tf
```

If you modified any `.graphql` file, `make generate` is required — `generated.go` must not be dirty.
If you added or removed any Go dependency, `go mod tidy` is required — `go.mod`/`go.sum` must not be dirty.

**Never commit** `.env` or `dev.tfrc` — both are gitignored and contain credentials or local paths.

---

## Patterns to Follow

### Consult the GraphQL schema first

Before writing any client code, read the schema to verify field names, argument types, union members, and available mutations:

- **`internal/client/schema/schema.graphql`** — full GraphQL schema. Look up the exact type definitions, input types, and union result types for the resource you are implementing.
- **`internal/client/schema/queries/`** — existing `.graphql` operation files. Follow these as the pattern for new query/mutation files.
- **`internal/client/schema/generated.go`** — auto-generated Go types from `genqlient`. Read this to understand the exact Go struct names produced from the schema before referencing them in client code.

After adding or modifying any `.graphql` file in `queries/`, run `make generate` to regenerate `generated.go`. **Never edit `generated.go` by hand** — it is overwritten on every `make generate`.

To determine whether a resource is org-scoped or deployment-scoped, check whether the relevant query/mutation is on the root `Query`/`Mutation` type (org-scoped) or requires a deployment-specific endpoint. Org-scoped operations use `c.gqlClient("")`; deployment-scoped use `c.gqlClient(deploymentName)`.

### PythonError in mutation unions

Many mutation result unions include `PythonError` as a possible member (e.g. `CreateCustomMetricResult = CreateCustomMetricSuccess | CustomMetricError | PythonError`). The `default` branch in the client's type switch will catch it and return an error with the Go type name. You do not need to add an explicit `PythonError` case, but you **must** always include a `default` branch — never write an exhaustive switch without one.

### GraphQL client: deployment-scoped vs org-scoped

```go
// Deployment-scoped (most resource operations)
c.doGraphQL(ctx, deploymentName, query, vars, &result)

// Org-scoped (deployments, users, teams)
c.doGraphQL(ctx, "", query, vars, &result)
```

### Computed-only fields that the provider derives

If the provider always calculates a field (e.g. `event_types` derived from a policy type block), make it `Computed: true` only — do not add `Optional: true`. This prevents user-supplied stale values from causing plan noise.

### Fields the API does not return

Some fields (like `policy_type` in alert policies) are used by the provider to interpret the API response but are not returned by the API. For these:
- Keep them in the model struct with `tfsdk` tags.
- In `Read`, preserve the existing state value (do not overwrite with an empty string).
- In acceptance tests, add to `ImportStateVerifyIgnore`.
- Note the limitation in README under the resource's Import section.

### Sensitive fields

Mark any field containing a credential, secret, or private URL as `Sensitive: true` in both the resource and data source schema. Do not log or print sensitive values in test output.

### Naming test resources

Prefix all acceptance test resource names with `acc-tf-` (e.g. `acc-tf-alert-health`). This makes it easy to identify and clean up leaked test resources in the real API.
