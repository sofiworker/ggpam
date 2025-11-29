# Repository Guidelines

Always respond in Chinese-simplified

## Target
- Use golang to rewrite `google-authenticator-libpam` project
- `google-authenticator-libpam` the project is located in the `/root/workspace/c/google-authenticator-libpam` directory or on internet https://github.com/google/google-authenticator-libpam
- First, you should give me your plan, then I confirm, you can do
- Anything logic should keep as the C project

## Coding Style & Naming Conventions
- Follow idiomatic Go 1.18 style: tabs for indentation, `UpperCamelCase` for exported identifiers, `lowerCamelCase` for locals.
- Always run `gofmt` (or `goimports`) before committing; CI assumes formatted sources.

## Testing Guidelines
- Place new tests alongside the code (`*_test.go`) or under `test/` when they span multiple modules.
- Name tests with `TestXxx` and prefer table-driven cases for coverage clarity.
- Smoke-test locally via `go test ./...`; run `make test` before pushing changes that affect Docker packaging.
- Document any required external services (ClickHouse, Kafka, Redis) in test helpers and guard network calls with fakes when possible.

## Commit & Pull Request Guidelines
- Match the existing conventional prefixes (`fix:`, `feat:`, `chore:`) followed by a short imperative summary.

## Tool Priority
- Exclude directory：`.git`, `node_modules`, `dist`, `coverage`