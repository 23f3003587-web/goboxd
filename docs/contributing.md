# Contributing

Thank you for helping improve goboxd.

## Development workflow

1. Fork the repository and create a feature branch.
2. Run the tests with `make test` and `make integration`.
3. Update docs or examples if the execution behavior changes.
4. Open a pull request with a concise summary of the change.

## Project layout

- `cmd/` contains the main binary.
- `internal/` contains service logic and sandbox integration.
- `pkg/` contains reusable helpers and extension points.
- `deployments/` contains container and orchestration artifacts.
