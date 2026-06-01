# Security Fixes

This document lists the major security holes from the reference implementation that were addressed in this `goboxd` codebase.

## Closed Holes

**1. Path traversal via filename**  
Fixed in:
- `internal/security/filename.go:9-25` — `ValidateFilename()` with `filepath.Clean` + `filepath.Base`
- `internal/sandbox/jail.go:55-64` — safe path joining and escape check in `WriteSource()`

**2. Shell-style directory commands**  
Fixed in:
- `internal/sandbox/executor.go:170-176` — using `exec.CommandContext` with argument arrays (no shell)

**4. No request size limits**  
Fixed in:
- `internal/handler/run.go:25` — `http.MaxBytesReader`
- `internal/config/config.go:58-82` — `ValidateRequest()` with source, stdin, and expected output limits

**5. UID collisions under load**  
Fixed in:
- `internal/sandbox/jail.go:47-52` — `os.MkdirTemp()`
- `internal/sandbox/jail.go:23-45` — startup orphan cleanup

**6. Unbounded child output**  
Fixed in:
- `internal/sandbox/executor.go:172-190` — `newTruncatingWriter` with truncation markers

**7. Stale jail directories**  
Fixed in:
- `internal/sandbox/jail.go:23-45` — `defer j.Cleanup()` + `init()` orphan cleanup