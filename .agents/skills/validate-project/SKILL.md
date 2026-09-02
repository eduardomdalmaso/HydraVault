---
name: validate-project
description: >-
  Procedure for running full project validation, DDD layer compliance audits,
  modular frontend file size checks, Go unit tests, build checks, HTTP endpoint verification, and documentation synchronization on HydraForge & HydraStream.
  Use when validating changes, performing code review, or before completing tasks.
---

# 🛡️ HydraForge & HydraStream Project Validation Skill

Follow this multi-step runbook to validate the integrity, DDD compliance, modular frontend rules (<100 lines), build health, runtime status, and documentation sync of HydraForge.

## Step 1: DDD Domain Purity Audit

Verify that `internal/domain` does not contain forbidden infrastructure imports (`net/http`, `database/sql`, etc.):

```bash
grep -rn "net/http" internal/domain/ && echo "❌ Violations found!" || echo "✅ Domain layer is pure!"
```

## Step 2: Modular Frontend Line Count Audit

Verify that no web file under `web/` exceeds 100 lines:

```bash
wc -l web/src/**/*.jsx web/src/**/*.js web/src/**/*.css
```

*Expected result:* All web files have fewer than 100 lines.

## Step 3: Run Unit & Concurrency Tests

Execute fast Go unit tests across all packages:

```bash
make test
```

*Expected result:* All domain and application tests pass cleanly.

## Step 4: Compile Production Binary

Ensure the application builds without syntax or dependency errors:

```bash
make build
```

## Step 5: Endpoint & Runtime Smoke Test

Verify HydraForge endpoints (:8081):

```bash
curl -s http://localhost:8081/api/v1/training/telemetry
curl -s http://localhost:8081/api/v1/training/jobs
curl -s http://localhost:8081/api/v1/training/models
curl -s http://localhost:8081/api/v1/datasets
```

## Step 6: Architecture Documentation Audit

Confirm that mandatory architectural docs are synchronized with codebase changes:

- Ensure [`GEMINI.md`](file:///home/hades/Documents/HydraForge/GEMINI.md) reflects current rules and package layout.
- Ensure [`README.md`](file:///home/hades/Documents/HydraForge/README.md) and [`README.pt-BR.md`](file:///home/hades/Documents/HydraForge/README.pt-BR.md) reflect all current subsystems and views.

## Step 7: Final Sanity Check

Check that all components follow the Ports & Adapters structure:

- `internal/domain` -> Entities & Domain Logic (Pure)
- `internal/ports` -> Interfaces (`JobRepository`, `TrainingWorker`, `GPUProvider`)
- `internal/application` -> Services (`TrainingService`)
- `internal/adapters` -> HTTP, SQLite, GPU, and Python Worker implementations
