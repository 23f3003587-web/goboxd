## 2026-05-27 · Stage 1- Phase 1 plan and project initialization

**Prompt:**
I provided the full goboxd spec and asked for a realistic 5-day plan for Stage 1, focusing on getting /healthz and basic /run working with Python.

**Response summary:**
AI gave a clear 5-day breakdown, recommended project structure, and prioritized foundation (Dockerfile, Makefile, languages.yaml, health handler). Also reminded about AI logging requirements.

**What we used / didn't use:**
Used the full 5-day plan structure and Day 1 files as base. Did not use the initial executor suggestions directly (they had bugs), but used the structure.

---

## 2026-05-28 · Phase 2 - Python execution

**Prompt:**
Help fix executor.go and jail.go so Python /run works end-to-end with nsjail.

**Response summary:**
Provided multiple iterations of executor.go with chroot/bindmount fixes, template replacement, and status logic.

**What we used / didn't use:**
Used the improved jail creation and basic execution logic. Discarded early versions that had wrong --chroot usage and path handling.

---

## 2026-05-29 · Phase 3 - Adding C++ support

**Prompt:**
Implement compiled language pipeline for C++ (build + run steps) based on updated languages.yaml structure.

**Response summary:**
Provided updates to languages.go, languages.yaml, and a major refactor of executor.go to support build/run steps, template replacement for {{artifact}}, and per-test execution.

**What we used / didn't use:**
Used the core runStep abstraction and template replacement logic. Iterated multiple times to fix {{artifact}} in cmd, flag expansion, whitespace comparison, and nil safety on errors. Discarded versions that didn't handle build failure correctly.

---

## 2026-05-30 · Phase 4 - Testing, Validation & Polish

**Prompt:**
We have basic Python and C++ working. Help with Day 4: add global limits, improved validation, unit + integration tests, better error messages, and documentation (README, api.md, security.md).

**Response summary:**
Provided config.go with global limits, validator.go, filename_test.go, integration tests, improved run.go handler with proper JSON error responses, and basic documentation files.

**What we used / didn't use:**
Used the structure for tests and validation logic. Fixed several test failures caused by import issues and handler registration problems. Improved error handling and added request logging.

---

## 2026-05-30 · Phase 5 - Final Stabilization & Demo Readiness

**Prompt:**
Stage 1 is almost done. Help complete Day 5: add panic recovery middleware, /info and /readyz endpoints, output truncation (Hole 6), final documentation, and fix 404 errors on health endpoints.

**Response summary:**
Provided middleware.go, updated health.go with proper chi middleware, main.go with correct route registration, security.md updates, and final polish commands.

**What we used / didn't use:**
Used the middleware and endpoint implementations. Fixed bugs where /info and /readyz returned 404 (caused by incorrect middleware attachment and missing imports). Resolved chi handler registration issues and added RecoveryMiddleware. Also fixed docker-compose version warning.

---

## 2026-06-01 · Final Stage 1 Submission Preparation

**Prompt:**  
Give me a full checklist for Stage 1 submission. Also help verify all functional requirements, Docker build, tests, and documentation. Review recent automated security changes.

**Response summary:**  
Provided complete Stage 1 checklist, helped verify tests and Docker, reviewed security patches (kept most of them), fixed multi-test comparison logic, and prepared final documentation and PR description.

**What we used / didn't use:**  
Used the checklist to systematically verify the project. Kept most recent security hardening changes (request limits, output truncation, path validation) as they passed all tests. Fixed bugs discovered during verification related to truncation affecting normal outputs.

---