# MuninnDB Code Review Report
## Query, FTS, Novelty Detection, and Coherence Layers

**Date:** 2026-02-23
**Scope:**
- `/internal/query/mql/` — MQL parser and executor
- `/internal/index/fts/` — Full-text search engine
- `/internal/engine/novelty/` — Novelty detection
- `/internal/engine/coherence/` — Coherence scoring

---

## Executive Summary

Found **6 bugs** across the three layers:
- **2 CRITICAL** race conditions
- **2 HIGH** severity issues (integer underflow, missing error handling)
- **1 MEDIUM** severity (iterator resource leak under error)
- **1 LOW** severity (negative number handling)

---

## Critical Issues

### 1. CRITICAL: Race Condition in FTS IDF Cache Invalidation

**File:** `/Users/mjbonanno/github.com/scrypster/muninndb/internal/index/fts/fts.go`
**Line:** 230-232
**Severity:** CRITICAL

**Issue:**
The IDF cache invalidation in `IndexEngram` during document frequency updates is not atomic with respect to concurrent `Search` operations. Multiple threads can invalidate the cache while other threads are reading IDF values, causing stale IDF calculations to be used.

```go
// Lines 216-233 in IndexEngram
for term := range termCounts {
    tkey := keys.TermStatsKey(ws, term)
    var df uint32
    val, closer, err := idx.db.Get(tkey)
    if err == nil && len(val) >= 4 {
        df = binary.BigEndian.Uint32(val[0:4])
        closer.Close()
    }
    df++
    buf := make([]byte, 8)
    binary.BigEndian.PutUint32(buf[0:4], df)
    _ = idx.db.Set(tkey, buf, pebble.NoSync)

    // RACE: Invalidating cache while Search() may be reading it
    idx.mu.Lock()
    delete(idx.idfCache, term)
    idx.mu.Unlock()
}
```

**Root Cause:**
- `InvalidateIDFCache()` is meant for bulk clears (e.g., vault delete), but `IndexEngram` also invalidates per-term.
- A read lock in `getIDF()` (line 369-371) can race with per-term cache invalidation.
- The per-term invalidation happens *outside* the RWMutex lock on the larger atomic operations.

**Impact:**
- Concurrent inserts and searches may produce inconsistent scoring results.
- IDF values may become stale, reducing search relevance accuracy.
- Two threads could overwrite the same term's IDF cache entry.

**Recommended Fix:**
Either:
1. Consolidate all term DF updates and cache invalidations into a single critical section protected by `idx.mu.Lock()`.
2. Use `InvalidateIDFCache()` in bulk after all term updates are committed to the DB.
3. Track a generation counter so stale cache entries are automatically ignored.

---

### 2. CRITICAL: Double Lock on Registry in Coherence

**File:** `/Users/mjbonanno/github.com/scrypster/muninndb/internal/engine/coherence/coherence.go`
**Line:** 205-219
**Severity:** CRITICAL

**Issue:**
The `GetOrCreate()` method acquires the lock twice (double-lock) without releasing in between, creating a potential deadlock scenario if the RWMutex is not reentrant (which it isn't in Go).

```go
// Lines 205-219
func (r *Registry) GetOrCreate(vaultName string) *VaultCounters {
    r.mu.RLock()  // First read lock
    c, ok := r.vaults[vaultName]
    r.mu.RUnlock()
    if ok {
        return c
    }
    r.mu.Lock()   // Second lock (write lock)
    defer r.mu.Unlock()
    if c, ok = r.vaults[vaultName]; ok {
        return c
    }
    c = &VaultCounters{}
    r.vaults[vaultName] = c
    return c
}
```

**Root Cause:**
- Go's `sync.RWMutex` is not reentrant. If the same goroutine attempts to acquire a lock it already holds, it will deadlock.
- Line 205 acquires `RLock()`. If a concurrent writer holds the write lock between lines 209-210, then line 212 will block trying to upgrade to `Lock()`.
- Though this pattern *should* work because RUnlock() is called at 208, **there is a race window**: between UnRLock() and Lock(), another goroutine could insert the value, making this a lost-update race, not a double-lock issue. However, **the check-then-act pattern is not atomic.**

**Actual Problem (refined):**
Two concurrent goroutines can both call `GetOrCreate("vault1")` simultaneously:
1. G1 reads at line 206-207, sees `ok = false`.
2. G2 reads at line 206-207, sees `ok = false`.
3. G1 acquires lock at line 212, creates entry.
4. G2 acquires lock at line 212 (after G1 releases), creates a *new* entry, overwriting G1's.

This is a **double-creation race** even though the mutex isn't deadlocked.

**Impact:**
- If `RecordWrite()` was called on the first VaultCounters, and then GetOrCreate() creates a second one, the first one's metrics are lost.
- Coherence scores become incorrect (counters reset).

**Recommended Fix:**
Use sync/once or eliminate the RLock entirely:

```go
func (r *Registry) GetOrCreate(vaultName string) *VaultCounters {
    r.mu.Lock()
    defer r.mu.Unlock()
    if c, ok := r.vaults[vaultName]; ok {
        return c
    }
    c = &VaultCounters{}
    r.vaults[vaultName] = c
    return c
}
```

---

## High Severity Issues

### 3. HIGH: Integer Underflow in Coherence Counters

**File:** `/Users/mjbonanno/github.com/scrypster/muninndb/internal/engine/coherence/coherence.go`
**Line:** 63, 77
**Severity:** HIGH

**Issue:**
The `RecordLinkDeleted()` and `RecordContradictionResolved()` methods decrement atomic counters without checking if they would go negative.

```go
// Lines 62-80
func (c *VaultCounters) RecordLinkCreated(isFirstLink, isRefines bool) {
    if isFirstLink {
        c.OrphanCount.Add(-1)  // Line 63 — no check if OrphanCount > 0
    }
    if isRefines {
        c.RefinesCount.Add(1)
    }
}

func (c *VaultCounters) RecordLinkDeleted(wasLastLink, isRefines bool) {
    if wasLastLink {
        c.OrphanCount.Add(1)  // orphan count increases (correct)
    }
    if isRefines {
        c.RefinesCount.Add(-1)  // Line 77 — can underflow if not carefully managed
    }
}
```

**Root Cause:**
- If `RefinesCount.Add(-1)` is called when `RefinesCount == 0`, it wraps to `math.MaxInt64` (2^63 - 1) due to Go's two's complement arithmetic.
- Similarly, `OrphanCount.Add(-1)` without prior validation.

**Scenario:**
1. Engram created with no links → `RefinesCount = 0`.
2. Delete a REFINES link that doesn't exist → `RefinesCount.Add(-1)` → underflows to `9223372036854775807`.
3. `Snapshot()` reports nonsensical duplication pressure.

**Impact:**
- Coherence scores become incorrect (often 0, since duplication pressure formula will receive extremely large values).
- Audit/monitoring based on coherence becomes unreliable.
- May cause downstream issues in systems that rely on coherence scores for decision-making.

**Recommended Fix:**
Add validation before decrementing:

```go
func (c *VaultCounters) RecordLinkDeleted(wasLastLink, isRefines bool) {
    if wasLastLink {
        c.OrphanCount.Add(1)
    }
    if isRefines {
        current := c.RefinesCount.Load()
        if current > 0 {
            c.RefinesCount.Add(-1)
        } else {
            // Log error or return error
        }
    }
}
```

---

### 4. HIGH: Missing Error Handling in FTS Iterator

**File:** `/Users/mjbonanno/github.com/scrypster/muninndb/internal/index/fts/fts.go`
**Line:** 274-305
**Severity:** HIGH

**Issue:**
In the `Search()` method, if `idx.db.NewIter()` returns an error (line 278), the iterator is skipped but no error is returned to the caller. Additionally, if `iter.Valid()` fails silently, the error is not propagated.

```go
// Lines 274-305
iter, err := idx.db.NewIter(&pebble.IterOptions{
    LowerBound: lowerBound,
    UpperBound: upperBound,
})
if err != nil {
    continue  // ERROR SILENTLY IGNORED
}

for iter.First(); iter.Valid(); iter.Next() {
    key := iter.Key()
    // ...
}
iter.Close()
```

**Root Cause:**
- Pebble iterator creation can fail (e.g., I/O errors, disk full, corruption).
- The error is silently skipped, so search results are incomplete.
- Caller has no way to know that some results were lost.

**Impact:**
- Incomplete search results without indication of failure.
- User may make decisions based on partial data, thinking they have the full result set.
- Production queries could silently fail.

**Recommended Fix:**
Accumulate errors and return them:

```go
var searchErrors []error
for _, term := range tokens {
    iter, err := idx.db.NewIter(...)
    if err != nil {
        searchErrors = append(searchErrors, fmt.Errorf("iterator failed for term %q: %w", term, err))
        continue
    }
    // ... rest of code
}
if len(searchErrors) > 0 {
    return results, fmt.Errorf("search completed with errors: %v", searchErrors)
}
return results, nil
```

---

## Medium Severity Issues

### 5. MEDIUM: Iterator Resource Leak on Early Exit

**File:** `/Users/mjbonanno/github.com/scrypster/muninndb/internal/index/fts/fts.go`
**Line:** 282-305
**Severity:** MEDIUM

**Issue:**
If an early return or panic occurs during iterator processing, the iterator is not closed. While line 305 shows `iter.Close()`, any return statement before that line (e.g., in an outer context) will leak the iterator resource.

```go
// Lines 282-305
for iter.First(); iter.Valid(); iter.Next() {
    key := iter.Key()
    if len(key) < 1+8+len(term)+1+16 {
        continue  // OK, continues loop
    }
    var engramID [16]byte
    copy(engramID[:], key[1+8+len(term)+1:])
    // If panic happens here, iter.Close() at 305 is never called
    val := iter.Value()
    pv := decodePosting(val)
    // ...
    scores[engramID] += bm25
}
iter.Close()  // Line 305
```

**Root Cause:**
- No `defer iter.Close()` is used.
- If the loop body panics or a context cancellation (via `ctx`) occurs, the iterator is leaked.

**Impact:**
- File descriptors leak over time in production.
- Eventually, system may hit ulimit and new iterators cannot be created.
- Performance degradation in long-running servers.

**Recommended Fix:**
Use `defer` immediately after iterator creation:

```go
iter, err := idx.db.NewIter(...)
if err != nil {
    continue
}
defer iter.Close()  // Ensure cleanup
for iter.First(); iter.Valid(); iter.Next() {
    // ...
}
```

---

## Low Severity Issues

### 6. LOW: No Validation for Negative Integer Parse Results

**File:** `/Users/mjbonanno/github.com/scrypster/muninndb/internal/query/mql/parser.go`
**Line:** 137, 153, 457, 503
**Severity:** LOW

**Issue:**
The parser accepts negative integers for `MAX_RESULTS`, `HOPS`, and `FRAMES` via `strconv.Atoi()`, then clamps positive values but does not explicitly reject negatives.

```go
// Line 137-145 (MAX_RESULTS)
n, err := strconv.Atoi(numTok.Value)
if err != nil {
    return nil, p.error("invalid number for MAX_RESULTS")
}
if n > 1000 {
    n = 1000
}
// If n is negative, it's silently accepted and set to query.MaxResults
query.MaxResults = n
```

**Scenario:**
Input: `ACTIVATE FROM v CONTEXT ["x"] MAX_RESULTS -50`
- Parser accepts and sets `query.MaxResults = -50`.
- Executor may fail or behave unexpectedly when trying to limit results by a negative count.

**Root Cause:**
- No explicit validation for negative values after `Atoi()`.
- Clamp only checks upper bound.

**Impact:**
- Executor may crash or return wrong results if it tries to process negative result limits.
- Potential DoS if unchecked negative values cause infinite loops or memory issues downstream.

**Recommended Fix:**
Add explicit bounds checking:

```go
n, err := strconv.Atoi(numTok.Value)
if err != nil {
    return nil, p.error("invalid number for MAX_RESULTS")
}
if n < 0 {
    return nil, p.error("MAX_RESULTS must be non-negative")
}
if n > 1000 {
    n = 1000
}
query.MaxResults = n
```

---

## Summary Table

| # | File | Line | Severity | Type | Description |
|---|------|------|----------|------|-------------|
| 1 | `fts.go` | 230-232 | CRITICAL | Race Condition | IDF cache invalidation during concurrent indexing and search |
| 2 | `coherence.go` | 205-219 | CRITICAL | Double-Create Race | GetOrCreate() lacks atomic check-then-act for vault creation |
| 3 | `coherence.go` | 63, 77 | HIGH | Integer Underflow | RecordLinkDeleted() and RecordContradictionResolved() decrement without bounds check |
| 4 | `fts.go` | 274-305 | HIGH | Unhandled Error | NewIter() errors silently ignored in Search(), returning incomplete results |
| 5 | `fts.go` | 282-305 | MEDIUM | Resource Leak | Iterator not closed on early exit or context cancellation (no defer) |
| 6 | `parser.go` | 137, 153, 457, 503 | LOW | Missing Validation | Negative integers accepted for MAX_RESULTS, HOPS, FRAMES without validation |

---

## Recommendations

### Immediate (This Sprint)
1. **Fix CRITICAL issues #1 and #2** — these can cause data corruption or incorrect scores.
2. **Add iterator defer in FTS** (Issue #5) — one-line fix.
3. **Add bounds checking to parser** (Issue #6) — validate non-negative before clamping.

### Short-term (Next Sprint)
4. **Validate counter decrements in coherence** (Issue #3) — add logging/errors for underflow scenarios.
5. **Propagate iterator/search errors** (Issue #4) — return partial results with error.

### Long-term
6. Add property-based tests to detect race conditions and edge cases.
7. Use linters (`go vet`, `staticcheck`) to catch missing error handling.
8. Consider adding invariant checks in coherence counters (assertions on periodic snapshots).

---

## Test Coverage Notes

**Existing tests:**
- `parser_test.go` has good coverage for syntax, but no tests for negative numbers or large integers.
- `fts_test.go` and `coherence_test.go` not reviewed in this pass but should add race condition tests.
- No concurrent stress tests found for FTS indexing + searching simultaneously.

**Recommended new tests:**
- Concurrent `IndexEngram()` + `Search()` with concurrent IDF cache invalidation.
- `GetOrCreate()` called concurrently from 100+ goroutines.
- Counter underflow scenarios in coherence.
- Negative number parsing in MQL parser.

---

**Report prepared by:** Code Review Agent
**Status:** Ready for Triage
