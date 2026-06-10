# API Documentation

This document describes the HTTP endpoints implemented by goboxd.

## Routes

### GET /healthz and GET /health
Returns a lightweight liveness response. The /health alias exists for compatibility with older clients.

Response (200 OK):

```json
{"status":"ok"}
```

### GET /readyz and GET /ready
Performs a deeper readiness check. It verifies the nsjail binary, the nsjail config file, and every language configured in languages.yaml. The /ready alias is available as well.

Response:
- 200 OK when all checks pass
- 503 Service Unavailable when one or more checks fail

Example response:

```json
{
  "status": "ok",
  "nsjail": { "path": "/usr/local/bin/nsjail", "ok": true },
  "config": { "path": "/app/config/nsjail.cfg", "ok": true },
  "languages": [
    { "id": "py3", "name": "Python 3", "ready": true }
  ]
}
```

### GET /info
Returns build metadata, runtime details, supported language definitions, global limits, and live request statistics.

Response (200 OK) includes:
- build.version, commit, date, go_version
- nsjail.path, version, config_path, ok
- languages: language IDs, names, versions, run/build limits, and allowed build flags
- limits: source size, test count, concurrency, output limits
- stats: in-flight jobs, total jobs, failed internal jobs, last internal error time

### POST /run
Executes code in the sandbox and returns a structured run result.

Request body:

```json
{
  "language": "py3",
  "source": "print('hello')",
  "source_filename": "solution.py",
  "artifact_filename": "solution",
  "build": {
    "limits": {
      "wall_time_s": 5,
      "memory_kb": 1048576,
      "max_processes": 100
    },
    "flags": ["-O2"]
  },
  "run": {
    "limits": {
      "wall_time_s": 9,
      "memory_kb": 102400,
      "max_processes": 100
    },
    "flags": []
  },
  "tests": [
    {
      "stdin": "",
      "expected_stdout": "hello"
    }
  ]
}
```

Notes:
- The body must be valid JSON and must not contain unknown fields.
- language must exist in config/languages.yaml.
- source is required and must be within the configured source-size limit.
- tests must contain 1 to 50 entries.
- Each test stdin and expected_stdout must fit within the configured byte limits.

Response (200 OK):

```json
{
  "status": "accepted",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "",
    "duration_ms": 4
  },
  "tests": [
    {
      "status": "accepted",
      "stdout": "hello\n",
      "stderr": "",
      "duration_ms": 3,
      "memory_peak_kb": 2048
    }
  ]
}
```

Status values:
- accepted — all tests passed
- wrong_output — stdout did not match expected output
- output_whitespace_mismatch — output differs only by whitespace normalization
- time_exceeded — the sandbox exceeded the wall-time limit
- memory_exceeded — the sandbox exceeded the memory limit or was killed by the OOM path
- runtime_error — compilation or execution failed
- build_failed — the build phase failed and tests were not executed

Error responses:
- 400 Bad Request for invalid JSON, unknown fields, or request validation failures
- 503 Service Unavailable when the server is currently at its maximum concurrent job count
- 500 Internal Server Error when the sandbox fails unexpectedly

## Example: invalid language

```sh
curl -i -X POST http://localhost:8080/run \
  -H 'Content-Type: application/json' \
  -d '{"language":"nope","source":"print(1)","tests":[{"stdin":"","expected_stdout":"1"}]}'
```

This returns HTTP 400 with a JSON error object describing the validation issue.