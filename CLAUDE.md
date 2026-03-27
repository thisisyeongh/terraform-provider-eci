# CLAUDE.md

## Overview

Elice Cloud Infrastructure (ECI) Terraform Provider. Built with Go 1.25, terraform-plugin-framework v1.19.0.

## Build & Run

- `make build` — Build binary (bin/terraform-provider-eci)
- `make format` — Run golangci-lint + gofmt + golines
- `make check` — Validate formatting and linting
- `make test` — Run unit tests
- `make testacc` — Run acceptance tests (requires TF_ACC=1)

## Code Conventions

### Package Structure

- `internal/api/` — API client and response types (file: entity_type.go)
- `internal/resource/` — Resources (file: resource_entity_type.go)
- `internal/datasource/` — Data sources (file: entity_type.go)
- `internal/provider/` — Provider configuration
- `internal/utils/` — Utility helpers
- `internal/acctest/` — Shared test helpers

### Naming

- API responses: `Resource[Entity][Method]Response` (e.g., `ResourceVirtualMachineGetResponse`)
- API methods: `Get/Post/Patch/Delete[Entity]` (uses PATCH, not PUT)
- Models: `Resource[Entity]Model` / `[Entity]DataSourceModel`
- Conversion functions: `resource[Entity]GetResponseTo[Entity]Model(ctx, response, data)`
- Constructors: `NewResource[Entity]()` / `New[Entity]DataSource()`

### Import Order

1. stdlib
2. internal (terraform-provider-eci/internal/*)
3. external (github.com/*)

### Import Aliases

- `ds "terraform-provider-eci/internal/datasource"`
- `res "terraform-provider-eci/internal/resource"`
- `. "terraform-provider-eci/internal/utils"` (dot import)

### Language

- All code, comments, variable names, commit messages, and documentation must be written in English.

### Resource Implementation Patterns

- All fields use `types.*` (`types.String`, `types.Map`, etc.)
- All resources support `tags` (Required, MapAttribute)
- Optional fields use pointers (`*string`, `*bool`); double pointers for unset-able fields (`**string`)
- Status polling: `waitStatus()` (exponential backoff, max 15s interval)
- Error reporting: `addResourceError()` → `diag.Diagnostics`
- Method receivers: `r` for resources, `d` for data sources, `api` for APIClient

### Testing Conventions

- Test files are colocated with source files (Go standard)
- Shared test helpers: `internal/acctest/` package (exported)
- Unit tests: `TestXxx` — no API calls, table-driven
- Acceptance tests: `TestAcc[Resource]_[scenario]` — real API calls, requires TF_ACC=1
- Config helpers: `testAcc[Resource]Config[Purpose](params)` (unexported)
- Check helpers: `testAccCheck[Resource]Destroy` (unexported)
- Use `resource.ParallelTest` (only `resource.Test` when serial execution is required)
- Resource names: generate with `acctest.RandomName(prefix)`
- All tests use randomized names to allow safe parallel execution across multiple developers

### Development Workflow

1. Write code and tests
2. Run locally: `source .env.test && make testacc` — verify all tests pass
3. Push and create PR
4. CI runs unit tests automatically
5. Maintainer reviews code, adds `run-acceptance-tests` label to trigger acceptance tests in CI
6. Maintainer verifies acceptance test results, then approves and merges

### CI Workflow

- **Unit test** (`test.yml`) — runs automatically on every push and PR
- **Acceptance test** (`acceptance-test.yml`) — runs only when a maintainer adds the `run-acceptance-tests` label to a PR
- External contributors only trigger unit tests; maintainers review the code first, then trigger acceptance tests via label

### Acceptance Test Environment

- Local: create `.env.test` (gitignored) with required env vars, then `source .env.test && make testacc`
- CI: `ECI_API_ACCESS_TOKEN` is stored in GitHub Secrets; all other config is hardcoded in `acceptance-test.yml`
- If pricing plan names or other infrastructure names change, update `.github/workflows/acceptance-test.yml` and `.env.test.example`
- See `.env.test.example` for the full list of required variables
