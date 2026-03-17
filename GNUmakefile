default: generate build

.PHONY: generate
generate:
	GOFLAGS="-mod=mod" go run github.com/Khan/genqlient internal/client/schema/genqlient.yaml

BINARY_NAME   = terraform-provider-dagsterplus
REGISTRY      = registry.terraform.io
NAMESPACE     = dagster-io
PROVIDER_NAME = dagsterplus
VERSION       = 0.1.0

OS_ARCH = $(shell go env GOOS)_$(shell go env GOARCH)

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------

.PHONY: build
build:
	go build -o $(BINARY_NAME)

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/$(REGISTRY)/$(NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/$(OS_ARCH)
	mv $(BINARY_NAME) ~/.terraform.d/plugins/$(REGISTRY)/$(NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/$(OS_ARCH)/$(BINARY_NAME)

# ---------------------------------------------------------------------------
# Testing
# ---------------------------------------------------------------------------

.PHONY: test
test:
	go test ./... -v -timeout 120s

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./internal/provider/... -v -timeout 120s

# ---------------------------------------------------------------------------
# Code quality
# ---------------------------------------------------------------------------

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: vet
vet:
	go vet ./...

# ---------------------------------------------------------------------------
# Docs (requires tfplugindocs)
# ---------------------------------------------------------------------------

.PHONY: docs
docs:
	tfplugindocs generate --provider-name dagsterplus

# ---------------------------------------------------------------------------
# Housekeeping
# ---------------------------------------------------------------------------

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

.PHONY: tidy
tidy:
	go mod tidy

# ---------------------------------------------------------------------------
# Local development
# ---------------------------------------------------------------------------

DEV_TFRC = $(CURDIR)/dev.tfrc
TF       = set -a && . ./.env && set +a && TF_CLI_CONFIG_FILE=$(DEV_TFRC) terraform

.PHONY: dev-setup
dev-setup: build
	@printf 'provider_installation {\n  dev_overrides {\n    "dagster-io/dagsterplus" = "%s"\n  }\n  direct {}\n}\n' \
		"$(CURDIR)" > dev.tfrc
	@echo "dev.tfrc written. Set DAGSTER_CLOUD_ORGANIZATION and DAGSTER_CLOUD_API_TOKEN, then run: make dev-plan"

.PHONY: dev-plan
dev-plan:
	$(TF) -chdir=local plan

.PHONY: dev-apply
dev-apply:
	$(TF) -chdir=local apply -auto-approve

.PHONY: dev-destroy
dev-destroy:
	$(TF) -chdir=local destroy
