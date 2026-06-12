<div align="center">

# goboxd

**A Go HTTP service for executing untrusted code in isolated sandboxes.**

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23-00ADD8.svg?logo=go&logoColor=white)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-Required-2496ED.svg?logo=docker&logoColor=white)](https://www.docker.com)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/thesouldev/goboxd/pulls)

</div>

---

## Overview

goboxd is a Go HTTP service that loads language definitions from [languages.yaml](languages.yaml), validates incoming requests, executes untrusted code inside an nsjail sandbox, and returns structured execution results. The API supports optional test cases, per-language build/run limits, and health/readiness probes for orchestration.

## What the service actually does

- Loads the language registry from [config/languages.yaml](config/languages.yaml) at startup.
- Exposes HTTP endpoints for liveness, readiness, system info, and code execution.
- Uses the Go chi router with request logging, panic recovery, and graceful shutdown.
- Runs each request in a temporary jail directory with nsjail, with per-language time, memory, and process limits.
- Enforces request-size limits and rejects invalid or unsafe requests before sandbox execution.

## Supported languages

The language registry currently defines these IDs:

- py3 — Python 3
- cpp — C++
- c — C
- java — Java
- js — JavaScript (Node)
- bash — Bash
- verilog — Verilog
- lua — Lua

## Quick start

### Prerequisites

- Docker with Compose v2
- A local Docker daemon

No host Go toolchain is required to run the service itself, but the project still includes Go-based tests and linting through the Docker tool container.

### Build and run

```sh
git clone https://github.com/thesouldev/goboxd.git
cd goboxd
make build
make run
```

The service listens on port 8080 by default.

### Basic health check

```sh
curl http://localhost:8080/healthz
curl http://localhost:8080/health
curl http://localhost:8080/readyz
curl http://localhost:8080/ready
curl http://localhost:8080/info
```

### Execute code

```sh
curl -X POST http://localhost:8080/run \
  -H 'Content-Type: application/json' \
  -d '{
    "language": "py3",
    "source": "print(\"hello\")",
    "tests": [{ "stdin": "", "expected_stdout": "hello" }]
  }'
```

## Runtime behavior

- POST /run validates the request body, rejects unknown fields, and applies global limits such as a 2 MiB request body cap and a maximum of 50 test cases.
- The sandbox pipeline writes the source file, optionally runs a build phase, executes the run phase with nsjail, and summarizes the result.
- The response includes a top-level status plus per-test details and build metadata.
- The readiness endpoint performs real probes for the nsjail binary, its config file, and each registered language.

## Development commands

```sh
make test         # unit tests in the Docker tools container
make integration  # end-to-end execution tests for all supported languages
make lint         # golangci-lint
make logs         # follow service logs
make stop         # stop the stack
```

## Project structure

```text
.
├── cmd/goboxd/        # main HTTP server entry point
├── internal/          # handlers, sandboxing, config, security, metrics, models
├── config/
│   ├── languages.yaml # language registry and limits
│   └── nsjail.cfg     # nsjail sandbox configuration
├── deployments/docker/ # runtime and tools images
├── docs/              # API, architecture, benchmarks, and security notes
└── tests/             # unit and integration coverage
```

## Contributing

Contributions are welcome. Please open an issue or discuss substantial changes before submitting a pull request.

## License

This project is distributed under the GNU General Public License v3.0. See [LICENSE](LICENSE) for the full text.
