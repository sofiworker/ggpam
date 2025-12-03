# Repository Guidelines

Always respond in Chinese-simplified

## Coding Style & Naming Conventions
- Follow idiomatic Go 1.18 style: tabs for indentation, `UpperCamelCase` for exported identifiers, `lowerCamelCase` for locals.
- Naming: Use `camelCase` for variables, `PascalCase` for types, and `snake_case` for file names.
- Linting: fixes must satisfy `golangci-lint` (same config used in CI).
- Errors: wrap with context; return early; avoid panics in library code.
- Always run `gofmt` (or `goimports`) before committing; CI assumes formatted sources.
- Implementations for different platforms rely on Golang build tags
- Expose constants or errors as much as possible to facilitate user-side judgment
- Prefer defining small interfaces and then composing them, rather than directly defining complete interfaces.
- Most configurations in config can be set via WithFunc.
- Use `go mod tidy` to check or add dependencies
- Context: First parameter for blocking operations
- Most strings need to support i18n, but log and error messages should remain in English for programmatic handling. Only when displaying information to users should i18n translation be performed.

## Testing Guidelines
- Place new tests alongside the code (`*_test.go`) or under `test/` when they span multiple modules.
- Name tests with `TestXxx` and prefer table-driven cases for coverage clarity.
- Smoke-test locally via `go test ./...`; run `make test` before pushing changes that affect Docker packaging.

## Commit & Pull Request Guidelines
- Match the existing conventional prefixes (`fix:`, `feat:`, `chore:`) followed by a short imperative summary.

## Tool Priority
- Exclude directory：`bin`, `.git`, `node_modules`, `dist`, `coverage`