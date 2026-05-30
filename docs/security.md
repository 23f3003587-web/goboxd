# Security Holes Fixed

1. **Path traversal (Hole 1)** - `internal/security/filename.go:15` + validation in `WriteSource`
2. **Stale directories (Hole 7)** - `defer jail.Cleanup()` in `executor.go`
3. **Basic flag injection prep (Hole 3)** - allowlist structure in languages.yaml
4. **Basic unbounded output (Hole 6)**  
   Fixed in: `internal/sandbox/executor.go` (truncateOutput function)