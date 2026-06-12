# Architecture

This project is organized as a small HTTP service around a sandbox execution pipeline.

## High-level flow

1. Startup in cmd/goboxd/main.go
   - Loads the language registry from languages.yaml.
   - Loads global runtime config.
   - Initializes the concurrency limiter.
   - Creates a chi router and registers the HTTP endpoints.
   - Starts the server with graceful shutdown.

2. Request handling in internal/api/handlers
   - /healthz and /readyz report liveness and readiness.
   - /info exposes build metadata, language capabilities, limits, and runtime stats.
   - /run validates JSON, enforces request-size limits, and applies the global job limiter before sandbox execution.

3. Validation and safety in internal/security
   - The request validator checks that the language exists, that the source is present, and that test payloads fit the configured byte limits.
   - Build and run flags are checked against each language's allowlist.
   - Filename validation protects the jail path from traversal issues.

4. Sandbox execution in internal/sandbox
   - Execute() creates a temporary jail directory and writes the source file into it.
   - If the language has a build phase, it runs the compiler or build command first.
   - The run phase is launched through nsjail with the configured time, memory, process, and output limits.
   - Output is captured, sanitized, and truncated to the configured maximum size.

5. Result reporting
   - The handler converts the sandbox result into a JSON response.
   - Metrics are updated for request count, in-flight jobs, and failed internal executions.

## Core components

### cmd/goboxd/main.go
The main entry point wires configuration, middleware, routes, and the HTTP server. It also attaches SIGINT/SIGTERM handlers so shutdown is graceful.

### internal/config
- Loads the YAML language registry from languages.yaml.
- Stores global runtime limits and the current environment override for concurrent jobs.
- Validates the request shape before execution.

### internal/api/handlers
Contains the HTTP-layer logic:
- RecoveryMiddleware for panic recovery
- Healthz/Readyz/Info handlers for diagnostics
- Run for the main execution path

### internal/security
Protects the execution path by validating filenames, build/run flags, and request content before any sandbox process is launched.

### internal/sandbox
Implements the real sandbox behavior:
- jail creation and cleanup
- source writing into the isolated working directory
- nsjail command construction
- execution status mapping and output truncation

### internal/model and internal/metrics
- model defines the JSON request and response contracts.
- metrics tracks request throughput and internal failure counts for /info.

## Data flow for /run

1. The client sends JSON to POST /run.
2. The handler rejects invalid JSON or unknown fields and enforces global request-size caps.
3. The security layer validates language, source, tests, and allowed flags.
4. The concurrency gate ensures the server does not exceed the configured maximum number of in-flight executions.
5. The sandbox layer writes the source into a temporary jail and executes the language-specific build/run commands through nsjail.
6. The handler returns a JSON object describing build status, test outcomes, and the aggregate top-level status.

## Why the architecture matters

- The YAML language registry makes the service extensible without changing Go code for each language.
- The nsjail-backed sandbox keeps untrusted code isolated from the host environment.
- The handler and security layers enforce strict validation before execution, which reduces the risk of malformed or unsafe input.
- The readiness and info endpoints make the service easier to run in containers and observe in production-like environments.

## Notes for contributors

- Update config/languages.yaml when adding or changing supported languages.
- Keep the sandbox limits and allowed flags in sync with the language definitions.
- Prefer validating at the handler/security level before the sandbox path is touched.
