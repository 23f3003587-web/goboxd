# Example requests

Run the service locally, then try the examples below.

## Health

```sh
curl http://localhost:8080/healthz
```

## Execute Python

```sh
curl -X POST http://localhost:8080/run \
  -H 'Content-Type: application/json' \
  -d '{
    "language": "py3",
    "source": "print(\"hello\")",
    "tests": [{ "stdin": "", "expected_stdout": "hello" }]
  }'
```
