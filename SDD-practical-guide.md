# Spec-Driven Development — Practical Guide

---

## 1. Design First (Problem & Contract Definition)

### Goal
Define **what the system does** before writing any code.

### Actions
- [x] Identify system boundaries and responsibilities
- [x] Define external interfaces (APIs, events)
- [x] Decide non-functional requirements (rate limits, retries, SLAs)

### Artifacts
- [x] OpenAPI / AsyncAPI
- [N/A] Event schemas (JSON Schema / Avro)
- [x] Functional specs (load balancing, rate limiting, caching, retries)

### Output
- Clear, reviewable contracts
- No implementation yet

---

## 2. Write the Specs (Source of Truth)

### Goal
Create explicit, versioned contracts that describe system behavior.

### Actions
- [x] Write OpenAPI definitions
- [N/A] Define event schemas with versioning
- [x] Document critical behaviors (errors, limits, observability)

### Artifacts
- [x] `/specs/openapi.yaml`
- [N/A] `/specs/events/*.json`
- [x] Markdown behavior specs

### Output
- Specs stored in git
- Specs independent from code

---

## 3. Validate Specs Automatically (CI Enforcement)

### Goal
Prevent invalid or incomplete specs from entering the codebase.

### Actions
- [x] Add CI jobs to validate specs
- [x] Fail builds if specs are invalid or missing
- [x] Enforce required sections and versions

### Artifacts
- [x] CI pipelines (GitHub Actions, GitLab CI)
- [x] Validation tools (swagger-cli, ajv)

### Output
- Specs become mandatory
- No broken contracts allowed

---

## 4. Generate Code from Specs (Optional but Recommended)

### Goal
Eliminate contract drift and boilerplate.

### Actions
- [x] Generate controllers, DTOs, clients
- [x] Keep generated code isolated
- [x] Do not modify generated files manually

### Artifacts
- [x] Generated API interfaces
- [x] Generated models/events

### Output
- Code aligned with specs
- Faster and safer development

---

## 5. Implement Behavior Behind the Contract

### Goal
Implement business logic without changing contracts.

### Actions
- [ ] Implement service logic
- [ ] Connect infrastructure (DB, Kafka, Redis)
- [ ] Respect spec-defined behavior

### Artifacts
- [ ] Application code (services, handlers)
- [ ] Infrastructure adapters

### Output
- Working system
- Behavior matches spec exactly

---

## 6. Contract & Spec Testing

### Goal
Verify that implementations respect the contracts.

### Actions
- [ ] Validate API responses against OpenAPI
- [ ] Validate events against schemas
- [ ] Add consumer-driven contract tests if needed

### Artifacts
- [ ] Contract tests
- [ ] Schema validation tests

### Output
- Guaranteed compatibility
- Early detection of breaking changes

---

## 7. Versioning, Release & Operation

### Goal
Evolve the system safely over time.

### Actions
- [ ] Version APIs and events (v1, v2)
- [ ] Deprecate explicitly
- [ ] Monitor based on spec-defined metrics

### Artifacts
- [ ] Versioned specs
- [ ] Changelogs
- [ ] Dashboards & alerts

### Output
- Safe evolution
- Production-grade system

---

## Golden Rules

- Specs are first-class citizens
- Specs live in git and block CI
- No code change without a spec change
- Code is an implementation detail

---

## One-line Summary

> **Spec → Validate → Generate → Implement → Verify → Evolve**