# Traversal Profiles — Design Document

**Date:** 2026-02-23
**Status:** Approved
**Scope:** BFS Phase 5 traversal profiles, auto-inference engine, MCP tool improvements

---

## Problem Statement

MuninnDB has 16 named RelTypes stored in every 40-byte ERF association record. Today they contribute **zero** to retrieval quality — the BFS Phase 5 formula (`score = parent_score × Weight × hopPenalty^depth`) ignores RelType entirely. The types are decorative.

Additionally, the `muninn_link` MCP tool only documents 5 of 16 relation types, and `muninn_recall` has no way to express retrieval intent.

---

## Solution: Traversal Profiles (C-B-A Layered Resolution)

### Resolution Chain

```
per-request profile override → auto-inferred profile → vault Plasticity default → "default"
```

Each layer solves a specific failure mode:
- **Auto-inference (C)** — works automatically on every call without agent or admin action. Mines the context array that already exists on every Activate call.
- **Vault default via Plasticity (B)** — safety net for specialized vaults (e.g., a decisions vault that should always use causal traversal).
- **Per-request override (A)** — escape hatch for power users and agents that know their intent explicitly.

---

## Profile Definitions

Five built-in profiles, hardcoded as package-level constants. No custom profiles, no editor.

| Profile | Include | Exclude | Boost |
|---|---|---|---|
| `default` | all edges | none | `{RelContradicts: 0.3, RelSupersedes: 0.5}` |
| `causal` | Causes, DependsOn, Blocks, PrecededBy, FollowedBy | none | `{RelCauses: 1.3}` |
| `confirmatory` | Supports, Implements, Refines, References | Contradicts | `{RelSupports: 1.2}` |
| `adversarial` | Contradicts, Supersedes, Blocks | none | `{RelContradicts: 1.5}` |
| `structural` | IsPartOf, BelongsToProject, CreatedByPerson | none | none |

---

## Auto-Inference Engine

**Not keyword matching — pattern matching on phrases with weighted scoring.**

A score ≥ 2 is required to override Default. Ambiguous queries fall through safely to vault default or Default profile.

### Pattern Rules

| Profile | Patterns | Score |
|---|---|---|
| `causal` | `^why\b`, `what caused`, `what led to`, `depends on.*?`, `root cause`, `blocked by`, `because of what` | 2 |
| `adversarial` | `contradict`, `conflict(s\|ing)? with`, `inconsisten`, `wrong about`, `disagree` | 3 |
| `structural` | `part of`, `belongs to`, `structure of`, `components? of`, `organiz` | 1 |
| `confirmatory` | `supports? .*(claim\|idea\|theory\|decision)`, `validat`, `evidence (for\|that)`, `confirm` | 2 |

### Inference Algorithm

```
1. Concatenate all context[] strings
2. Run all rules against concatenated string
3. Sum scores per profile
4. If max score >= 2: return that profile
5. Else: return vault Plasticity.DefaultProfile (or "default")
```

---

## Codebase Changes

### New File: `engine/activation/profiles.go`
- `TraversalProfile` struct (Include []RelType, Exclude []RelType, Boost map[RelType]float32)
- `builtinProfiles` map — five hardcoded profiles
- `InferProfile(contexts []string, vaultDefault string) string`
- Profile lookup helper

### Modified: `engine/activation/engine.go`
- `phase5Traverse` adds profile parameter
- BFS loop: skip excluded edges, multiply by Boost modifier
- New formula: `score = parent_score × Weight × modifier × hopPenalty^depth`
- `resolveProfile()` implements C-B-A resolution chain
- Log resolved profile on every activation (non-negotiable for tuning)

### Modified: `transport/mbp/types.go`
- Add `Profile string` to `ActivateRequest`

### Modified: Plasticity config (wherever it lives)
- Add `DefaultProfile string` to Plasticity struct

### Modified: `mcp/tools.go`
- `muninn_recall`: add optional `profile` param (enum: default, causal, confirmatory, adversarial, structural) with descriptions of when to use each
- `muninn_link`: expand `relation` description to all 16 types with "when to use" guidance for each

### Modified: `mcp/handlers.go`
- Pass `Profile` field through from MCP request to `ActivateRequest`
- Expand `relTypeFromString` to handle all 16 types (currently only handles 5, defaults unknown to `relates_to`)

---

## Testing Requirements

This feature requires comprehensive test coverage before ship. Tests are not optional.

### Unit Tests
- `InferProfile()`: one test per pattern rule (positive and negative cases)
- `resolveProfile()`: all branches of C-B-A resolution chain
- BFS with each profile: verify Include filtering removes correct edges
- BFS with each profile: verify Exclude filtering removes correct edges
- BFS Boost multipliers: verify score is correctly modified
- Boost ×0.3 on RelContradicts: verify low-scoring propagation
- Empty Include (default): verify all edges traversed
- Non-empty Include (causal): verify only allowed edges traversed
- Profile with Exclude: verify excluded edges not traversed even if Include is empty

### Integration Tests
- End-to-end Activate with `profile="causal"`: results differ from default on a vault with mixed edge types
- End-to-end Activate with `profile="adversarial"`: surfaces contradiction-linked engrams
- Auto-inference fires correctly on causal query string
- Auto-inference fires correctly on adversarial query string
- Auto-inference falls through to Default on ambiguous query
- Vault Plasticity DefaultProfile respected when no inference fires
- Per-request Profile override beats vault default
- Per-request Profile override beats inferred profile

### MCP Tool Tests
- `muninn_recall` with `profile` param: correct profile passed through to engine
- `muninn_link` with all 16 relation type strings: none should default to `relates_to` erroneously
- `muninn_link` with unknown relation string: still defaults to `relates_to` with no panic
- Activation response logs `profile_used` (Phase 2: also returns it in response body)

### Regression Tests
- Existing Activate calls with no `Profile` field: behavior identical to pre-feature (resolves to "default", same BFS formula with new default boost table applied)
- `DisableHops: true` still works with profiles present
- `Weights` override still works with profiles present

---

## Phase 2 (Post-Ship)

After production data is available:
- Add `inferred_profile string` to activation response so agents can observe what fired
- Tune inference patterns based on logs
- Adjust Boost multiplier values based on retrieval quality feedback

---

## What We Are NOT Building

- Custom user-defined profiles
- Profile editor or API
- Profile composition
- ML-based inference
- Embeddings-based intent detection

YAGNI. Five built-ins + pattern inference covers 95% of use cases.
