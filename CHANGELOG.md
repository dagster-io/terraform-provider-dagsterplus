# Changelog

All notable changes to this provider will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). This project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.5] - 2026-05-01

### Added
- `pkg/provider` package exporting a top-level `New` constructor to support the Pulumi Terraform bridge.

## [0.1.4] - 2026-04-28

### Breaking Changes
- `dagsterplus_code_location`: `working_directory`, `executable_path`, `attribute`, and `agent_queue` are no longer `Computed`. These fields were previously `Optional+Computed`, meaning Terraform would persist whatever the API returned even if the field was omitted from config. They are now `Optional`-only and will be `null` in state when not set. Users with existing state that has these fields populated from a prior computed read should set the values explicitly in their config (or run `terraform state rm` and re-import) to avoid a persistent plan diff.

### Fixed
- Code location fields no longer incorrectly show as unknown after creation; optional fields now preserve state correctly.
- Provider now strips trailing slashes from `base_url` to prevent request failures when the URL is supplied with a trailing `/`.
- HTTP 500 responses containing a `PythonError` body are now handled gracefully instead of causing an opaque client error.
- `dagsterplus_user` resource no longer removes itself from state when the invited user is in a pending-invite state; state is preserved and refreshed once the invite is accepted.

### Changed
- Acceptance tests now run in CI against the live API on every pull request.

## [0.1.3] - 2026-04-14

### Added
- `container_context` support on `dagsterplus_code_location_from_document`, allowing container image and environment configuration.

### Fixed
- `dagsterplus_code_location_from_document` now accepts camelCase keys in the document (snake_case keys were removed as a simplification).
- Improved error handling: `InvalidLocationError` responses now surface the full message and error list.

## [0.1.2] - 2026-04-02

### Changed
- Provider documentation now includes a complete index template and improved formatting on the Terraform Registry.

## [0.1.1] - 2026-04-02

### Added
- Automated documentation generation via `tfplugindocs` in CI (docs are now auto-committed on changes to provider code or examples).

## [0.1.0] - 2026-04-02

### Added
- Initial public release.
- **21 resources:** `dagsterplus_deployment`, `dagsterplus_code_location`, `dagsterplus_team`, `dagsterplus_user`, `dagsterplus_alert_policy`, `dagsterplus_agent_token`, `dagsterplus_user_token`, `dagsterplus_role`, `dagsterplus_scim_settings`, `dagsterplus_atlan_integration`, `dagsterplus_github_integration`, `dagsterplus_secret`, `dagsterplus_deployment_settings`, `dagsterplus_service_user`, `dagsterplus_service_token`, `dagsterplus_organization_settings`, `dagsterplus_custom_metric`, `dagsterplus_external_asset_connection`, `dagsterplus_team_membership`, `dagsterplus_team_deployment_grant`, `dagsterplus_code_location_from_document`.
- **20 data sources** for all major resource types.
- Import support for all resources.
- Migration guide from `datarootsio/dagster` provider.
- Full CI pipeline (build, vet, format, unit tests, module tidy, generated code check).
- GoReleaser-based release pipeline with GPG-signed checksums.

[Unreleased]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.5...HEAD
[0.1.5]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/dagster-io/terraform-provider-dagsterplus/releases/tag/v0.1.0
