# Terraform Provider for Dagster+

Manage [Dagster+](https://dagster.cloud) (Dagster Cloud) resources declaratively with Terraform.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) ≥ 1.0
- [Go](https://golang.org/) ≥ 1.21 (to build from source)

## Authentication

The provider requires a Dagster+ API token and your organization name. Both can be
supplied via HCL attributes or environment variables:

| Attribute       | Environment variable         |
|-----------------|------------------------------|
| `organization`  | `DAGSTER_CLOUD_ORGANIZATION` |
| `api_token`     | `DAGSTER_CLOUD_API_TOKEN`    |

Generate a token at **Dagster+ → Account Settings → API Tokens**.

## Usage

### Managing resources

You can manage resources using the `terraform apply` command. For example, to create a new deployment named `production`:

```hcl
resource "dagsterplus_deployment" "production" {
  name = "production"
}
```

### Data sources

You can use data sources to retrieve information about existing resources. For example, to list all deployments in your organization:

```hcl
data "dagsterplus_deployments" "all" {}

output "deployments" {
  value = data.dagsterplus_deployments.all
}
```

### Importing existing resources

You can import existing resources into your Terraform state using the `terraform import` command. For example, to import an existing deployment:

```hcl
resource "dagsterplus_deployment" "production" {
  name = "production"
}
```

More examples are available in the [`examples`](./examples/) directory and in the provider documentation on the Terraform Registry.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute to this provider.

## License

This provider is distributed under the [Mozilla Public License, Version 2.0](LICENSE).
