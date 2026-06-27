# Evaluation Harness & Benchmarking

Slash measures compression on two axes: **token reduction** and **task pass-rate**. Neither is meaningful without the other — aggressive compression that breaks functionality is useless.

## Philosophy

1. **No hand-waved numbers.** Every savings claim is reproducible via the published harness.
2. **Measured blended reductions.** Averaging across real sessions, not cherry-picking best-case surfaces.
3. **Quality = pass-rate.** A task is "pass" if the agent solves it end-to-end (tests pass, PR is valid, etc.).
4. **Holdout validation.** ~10% of tasks run uncompressed as a regression detector.
5. **Per-host + per-surface breakdown.** Where are savings coming from? (code reads, JSON, logs, re-reads?)

## Task Set

The harness runs on:
- **SWE-Bench subset** (100 tasks from the full 2k set) — semi-random selection of Python/JavaScript repos, varying complexity from easy to hard.
- **Internal task library** (optional) — your own tickets/issues, if available.

Tasks are frozen for reproducibility. Adding new tasks requires consensus (to avoid benchmark garden-hening).

## Session Structure

Each task execution:
1. **Init**: clean workspace, seed initial repo files.
2. **Prompt**: agent gets the task description + context.
3. **Loop**: agent makes tool calls, system routes them → compressor → response.
4. **End**: check if task is solved (tests pass, expected files exist).

Tool calls are logged with:
- `tool` (read, bash, apply_patch, etc.)
- `tool_input` (path, command, etc.)
- `uncompressed_tokens` (actual size before compression)
- `compressed_tokens` (size after compression, or -1 if not compressed)
- `compression_method` (diff_only, skeleton, json_crush, dedup, etc.)
- `latency_ms` (hook round-trip time)

## Measurement

### Token Reduction

For each tool call:
```
savings_ratio = (uncompressed_tokens - compressed_tokens) / uncompressed_tokens
```

Aggregate metrics:
```
blended_reduction = sum(uncompressed_tokens - compressed_tokens) / sum(uncompressed_tokens)
per_surface = grouped by {code_read, json, logs, re_read, bash, patch}
confidence_range = 95th percentile margin (typically ±5%)
```

### Pass-Rate

For each task:
```
result = "pass" if tests/validation succeed, else "fail"
pass_rate(cohort) = count(pass) / count(tasks)
```

Compare across:
- Uncompressed baseline (control group, ~10% of runs).
- Compressed (the treatment).
- Per-host (Claude Code, Codex, etc.).
- Per-surface (what types of compression help most).

### Latency

For each hook call:
```
latency = time(adapter.DecodeHookEvent) 
        + time(core.Compress) 
        + time(adapter.EncodeHookResult)
```

Track p50, p95, p99. Hot-path budget: p95 < ~50ms (fail-open if exceeded).

## Running Evals

### Setup

```bash
# Install dependencies
go install ./cmd/slash

# Download task set (or use your own)
slash eval download-tasks --set swe-bench-100

# Pre-generate repo snapshots for faster iteration
slash eval prepare-snapshots
```

### Execute

```bash
# Run full eval: 90 compressed + 10 holdout
slash eval run \
  --task-set swe-bench-100 \
  --compressed-fraction 0.9 \
  --output eval-results-$(date +%Y%m%d).json

# Subset (e.g., quick smoke test)
slash eval run \
  --task-set swe-bench-100 \
  --max-tasks 20 \
  --compressed-fraction 0.9 \
  --output eval-smoke-test.json
```

### Analyze

```bash
# Generate report: text + CSV tables + plots
slash eval report \
  --input eval-results-20240101.json \
  --format text,csv,html \
  --output eval-report-20240101/
```

Output includes:
- Summary stats: reduction %, pass-rate delta, latency p50/p95.
- Per-host breakdown (Claude Code vs. Codex vs. etc.).
- Per-surface breakdown (code vs. JSON vs. logs).
- Pass/fail breakdown by task (find regressions).
- Confidence ranges and statistical significance tests.

## Expected Results (V1)

Based on the design:
| Metric | Expected | Confidence |
|---|---|---|
| **Token reduction** | 40–55% | ±5% |
| **Pass-rate (compressed)** | 67–70% vs. 69–72% (baseline) | ±1–2% |
| **Latency overhead** | p50 <5ms, p95 <40ms | High |
| **Per-surface savings** | Code 60%, JSON 45%, logs 65%, re-reads 90% | Medium |

**Quality risk factors:**
- Diff-only re-reads: very safe; skip if model asks for full context.
- Output compression: safe with retrieve; quality loss <1% empirically.
- Skeleton reads: depends on language; Go/TS are well-supported.

## Regression Testing

The holdout (10% of tasks uncompressed) serves as a canary:
- If holdout pass-rate drops >2%, investigate.
- If compressed pass-rate drops >3% below holdout, enable kill-switch.
- Per-task regression: if a specific task fails compressed but passes uncompressed, debug the compression.

## Interpreting Results

**Good results:**
- Token reduction 40%+ with pass-rate within 1% of baseline.
- Per-surface breakdown shows expected patterns (re-reads >80%, code 50–70%).
- p95 latency < 40ms (no user-visible slowdown).

**Concerning results:**
- Token reduction >60% with pass-rate drop >3% → compression too aggressive.
- p95 latency >100ms → daemon bottleneck, investigate.
- One host consistently underperforms → adapter bug or hook-schema mismatch.
- Specific surface regresses (e.g., JSON drops to 20%) → compressor tuning needed.

## Continuous Eval

The harness is integrated into CI/CD:
- Every commit runs a smoke test (20 tasks, 90% compressed).
- Weekly full eval on all 100 tasks (published to dashboard).
- On release, full eval + comparison to previous version.

## Custom Tasks

To add your own tasks:

```bash
# Create a task file
cat > my-task.json << 'EOF'
{
  "id": "my-custom-task-1",
  "title": "Fix the auth bug in my app",
  "description": "...",
  "repo": "myrepo",
  "initial_files": [...],
  "expected_outcome": "tests pass",
  "validation": {
    "command": "pytest tests/",
    "success_pattern": "passed"
  }
}
EOF

# Add to your eval set
slash eval add-task --file my-task.json --set my-tasks

# Run eval
slash eval run --task-set my-tasks
```

## Sharing Results

When publishing eval results:
1. Commit the raw JSON to `eval/results/`.
2. Generate the report: `slash eval report --input results/json`.
3. Upload plots to the website / documentation.
4. Describe the methodology and caveat about hyperparameter sensitivity.
5. Invite reproduction: "Run `slash eval run` to verify these results on your hardware."

## Limitations & Caveats

- **Hardware variability:** Latency measurements depend on disk/CPU; report on consistent hardware (CI, not laptop).
- **Model variance:** Different agent versions may compress differently. Note the model version (Opus, Sonnet, etc.) in results.
- **Task-set bias:** 100 tasks may not reflect your use case. Encourage users to eval on their own repos.
- **No real user sessions:** Evals are structured; real sessions have longer context windows and may compress differently.
- **Compression hyperparam sensitivity:** Skeleton depth, truncation length, etc., are tuned on this task set. May not generalize.

---

**TL;DR:** Measure both axes (tokens + pass-rate), holdout for regression detection, publish the harness so anyone can reproduce, and be honest about confidence ranges.
