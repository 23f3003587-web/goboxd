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