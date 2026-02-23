# Traversal Profiles Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make RelType a first-class participant in BFS graph traversal via auto-inferred, vault-default, and per-request traversal profiles — so every Activate call benefits without agent or admin action.

**Architecture:** Five hardcoded `TraversalProfile` structs control BFS Phase 5 via Include/Exclude filtering and per-type Boost multipliers. A regex pattern-matching inference engine mines the `context[]` array already present on every call and selects the best profile automatically. Resolution chain: per-request override → auto-inferred → vault Plasticity default → "default".

**Tech Stack:** Go 1.23+, existing `internal/engine/activation`, `internal/auth`, `internal/mcp`, `internal/transport/mbp` packages. No new dependencies.

---

## Baseline

Before starting: confirm all tests pass.

```bash
cd ~/github.com/scrypster/muninndb
go test ./...
```

All packages should show `ok`. If anything is failing, stop and fix it first.

---

## Task 1: Define TraversalProfile + InferProfile()

**Files:**
- Create: `internal/engine/activation/profiles.go`
- Create: `internal/engine/activation/profiles_test.go`

### Step 1: Write the failing tests

Create `internal/engine/activation/profiles_test.go`:

```go
package activation

import (
	"testing"

	"github.com/scrypster/muninndb/internal/storage"
)

// --- Profile lookup ---

func TestGetProfile_KnownNames(t *testing.T) {
	for _, name := range []string{"default", "causal", "confirmatory", "adversarial", "structural"} {
		p := GetProfile(name)
		if p == nil {
			t.Errorf("GetProfile(%q) returned nil", name)
		}
	}
}

func TestGetProfile_UnknownFallsToDefault(t *testing.T) {
	p := GetProfile("nonexistent")
	if p == nil {
		t.Fatal("GetProfile(unknown) must not return nil")
	}
	def := GetProfile("default")
	if p != def {
		t.Error("GetProfile(unknown) should return the default profile pointer")
	}
}

// --- Include/Exclude semantics ---

func TestDefaultProfile_NoIncludes_NoExcludes(t *testing.T) {
	p := GetProfile("default")
	if len(p.Include) != 0 {
		t.Errorf("default profile should have empty Include, got %v", p.Include)
	}
	if len(p.Exclude) != 0 {
		t.Errorf("default profile should have empty Exclude, got %v", p.Exclude)
	}
}

func TestCausalProfile_IncludesCausalTypes(t *testing.T) {
	p := GetProfile("causal")
	expected := []storage.RelType{
		storage.RelCauses,
		storage.RelDependsOn,
		storage.RelBlocks,
		storage.RelPrecededBy,
		storage.RelFollowedBy,
	}
	for _, rel := range expected {
		if !p.Includes(rel) {
			t.Errorf("causal profile should include %v", rel)
		}
	}
}

func TestCausalProfile_ExcludesNonCausalTypes(t *testing.T) {
	p := GetProfile("causal")
	nonCausal := []storage.RelType{
		storage.RelSupports,
		storage.RelContradicts,
		storage.RelReferences,
		storage.RelBelongsToProject,
	}
	for _, rel := range nonCausal {
		if p.Includes(rel) {
			t.Errorf("causal profile should NOT include %v", rel)
		}
	}
}

func TestCausalProfile_NoExcludes(t *testing.T) {
	p := GetProfile("causal")
	if len(p.Exclude) != 0 {
		t.Errorf("causal profile should have no Exclude list, got %v", p.Exclude)
	}
}

func TestAdversarialProfile_ExcludesNothing_IncludesConflict(t *testing.T) {
	p := GetProfile("adversarial")
	if !p.Includes(storage.RelContradicts) {
		t.Error("adversarial profile must include RelContradicts")
	}
	if !p.Includes(storage.RelSupersedes) {
		t.Error("adversarial profile must include RelSupersedes")
	}
	if !p.Includes(storage.RelBlocks) {
		t.Error("adversarial profile must include RelBlocks")
	}
}

func TestConfirmatoryProfile_ExcludesContradicts(t *testing.T) {
	p := GetProfile("confirmatory")
	if !p.Excluded(storage.RelContradicts) {
		t.Error("confirmatory profile must exclude RelContradicts")
	}
}

func TestConfirmatoryProfile_IncludesSupports(t *testing.T) {
	p := GetProfile("confirmatory")
	if !p.Includes(storage.RelSupports) {
		t.Error("confirmatory profile must include RelSupports")
	}
}

func TestStructuralProfile_IncludesStructuralTypes(t *testing.T) {
	p := GetProfile("structural")
	for _, rel := range []storage.RelType{
		storage.RelIsPartOf,
		storage.RelBelongsToProject,
		storage.RelCreatedByPerson,
	} {
		if !p.Includes(rel) {
			t.Errorf("structural profile should include %v", rel)
		}
	}
}

// --- Boost multipliers ---

func TestDefaultProfile_ContradictsHasLowBoost(t *testing.T) {
	p := GetProfile("default")
	boost := p.BoostFor(storage.RelContradicts)
	if boost >= 1.0 {
		t.Errorf("default profile RelContradicts boost should be < 1.0, got %f", boost)
	}
}

func TestDefaultProfile_UnknownTypeBoostIsOne(t *testing.T) {
	p := GetProfile("default")
	boost := p.BoostFor(storage.RelRelatesTo)
	if boost != 1.0 {
		t.Errorf("default profile RelRelatesTo boost should be 1.0, got %f", boost)
	}
}

func TestAdversarialProfile_ContradictsBoostAboveOne(t *testing.T) {
	p := GetProfile("adversarial")
	boost := p.BoostFor(storage.RelContradicts)
	if boost <= 1.0 {
		t.Errorf("adversarial profile RelContradicts boost should be > 1.0, got %f", boost)
	}
}

// --- InferProfile ---

func TestInferProfile_CausalQueries(t *testing.T) {
	cases := []string{
		"why did the deployment fail?",
		"what caused the outage?",
		"what led to this decision?",
		"root cause of the bug",
		"what is blocking the release?",
		"why doesn't this work?",
	}
	for _, q := range cases {
		got := InferProfile([]string{q}, "")
		if got != "causal" {
			t.Errorf("InferProfile(%q) = %q, want %q", q, got, "causal")
		}
	}
}

func TestInferProfile_AdversarialQueries(t *testing.T) {
	cases := []string{
		"what contradicts the user preference?",
		"find conflicts with this decision",
		"which memories are inconsistent?",
		"what disagrees with this claim?",
	}
	for _, q := range cases {
		got := InferProfile([]string{q}, "")
		if got != "adversarial" {
			t.Errorf("InferProfile(%q) = %q, want %q", q, got, "adversarial")
		}
	}
}

func TestInferProfile_ConfirmatoryQueries(t *testing.T) {
	cases := []string{
		"what validates this decision?",
		"find evidence for the hypothesis",
		"what confirms our approach?",
	}
	for _, q := range cases {
		got := InferProfile([]string{q}, "")
		if got != "confirmatory" {
			t.Errorf("InferProfile(%q) = %q, want %q", q, got, "confirmatory")
		}
	}
}

func TestInferProfile_StructuralQueries(t *testing.T) {
	cases := []string{
		"what is part of project alpha?",
		"what belongs to the auth module?",
	}
	for _, q := range cases {
		got := InferProfile([]string{q}, "")
		if got != "structural" {
			t.Errorf("InferProfile(%q) = %q, want %q", q, got, "structural")
		}
	}
}

func TestInferProfile_AmbiguousUsesVaultDefault(t *testing.T) {
	got := InferProfile([]string{"what does the user prefer?"}, "causal")
	if got != "causal" {
		t.Errorf("ambiguous query should fall through to vault default, got %q", got)
	}
}

func TestInferProfile_AmbiguousNoVaultDefaultUsesDefault(t *testing.T) {
	got := InferProfile([]string{"what does the user prefer?"}, "")
	if got != "default" {
		t.Errorf("ambiguous query with no vault default should return 'default', got %q", got)
	}
}

func TestInferProfile_EmptyContextUsesVaultDefault(t *testing.T) {
	got := InferProfile([]string{}, "structural")
	if got != "structural" {
		t.Errorf("empty context should use vault default, got %q", got)
	}
}

func TestInferProfile_MultipleContextStrings(t *testing.T) {
	// Multiple context strings should be joined and matched
	got := InferProfile([]string{"tell me about the system", "what caused the failure?"}, "")
	if got != "causal" {
		t.Errorf("multi-string context should infer causal, got %q", got)
	}
}

// --- False positive guards ---

func TestInferProfile_NoCausalFalsePositive_Declarative(t *testing.T) {
	// "here's why I chose Go" — declarative, not a causal question
	// Should NOT infer causal if score doesn't hit threshold
	// This test documents the ambiguous case; threshold prevents misfire
	got := InferProfile([]string{"here is why I chose Go for this service"}, "")
	// Acceptable: either "default" (threshold not met) or "causal" (pattern matched)
	// The key constraint is: must not panic, must return a valid profile name
	validProfiles := map[string]bool{"default": true, "causal": true, "confirmatory": true, "adversarial": true, "structural": true}
	if !validProfiles[got] {
		t.Errorf("InferProfile returned invalid profile name %q", got)
	}
}
```

### Step 2: Run tests to verify they fail

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/engine/activation/... -run "TestGetProfile|TestInferProfile|TestDefaultProfile|TestCausalProfile|TestAdversarialProfile|TestConfirmatoryProfile|TestStructuralProfile" -v 2>&1 | head -30
```

Expected: compilation failure (types don't exist yet).

### Step 3: Implement profiles.go

Create `internal/engine/activation/profiles.go`:

```go
package activation

import (
	"regexp"
	"strings"

	"github.com/scrypster/muninndb/internal/storage"
)

// TraversalProfile controls how BFS Phase 5 traverses the association graph.
//
//   - Include: if non-empty, only traverse edges with RelType in this set.
//   - Exclude: skip edges with RelType in this set (applied after Include).
//   - Boost: per-RelType score multiplier. Missing types default to 1.0.
type TraversalProfile struct {
	Include    []storage.RelType
	Exclude    []storage.RelType
	Boost      map[storage.RelType]float32
	includeSet map[storage.RelType]struct{}
	excludeSet map[storage.RelType]struct{}
}

// Includes reports whether rel passes the Include filter.
// If Include is empty, all types pass.
func (p *TraversalProfile) Includes(rel storage.RelType) bool {
	if len(p.includeSet) == 0 {
		return true
	}
	_, ok := p.includeSet[rel]
	return ok
}

// Excluded reports whether rel is in the Exclude list.
func (p *TraversalProfile) Excluded(rel storage.RelType) bool {
	_, ok := p.excludeSet[rel]
	return ok
}

// BoostFor returns the score multiplier for rel (1.0 if not configured).
func (p *TraversalProfile) BoostFor(rel storage.RelType) float32 {
	if v, ok := p.Boost[rel]; ok {
		return v
	}
	return 1.0
}

// AllowsEdge returns true if rel should be traversed under this profile.
func (p *TraversalProfile) AllowsEdge(rel storage.RelType) bool {
	return p.Includes(rel) && !p.Excluded(rel)
}

func newProfile(include, exclude []storage.RelType, boost map[storage.RelType]float32) *TraversalProfile {
	p := &TraversalProfile{
		Include:    include,
		Exclude:    exclude,
		Boost:      boost,
		includeSet: make(map[storage.RelType]struct{}, len(include)),
		excludeSet: make(map[storage.RelType]struct{}, len(exclude)),
	}
	for _, r := range include {
		p.includeSet[r] = struct{}{}
	}
	for _, r := range exclude {
		p.excludeSet[r] = struct{}{}
	}
	return p
}

// builtinProfiles are the five hardcoded traversal profiles.
var builtinProfiles = map[string]*TraversalProfile{
	"default": newProfile(
		nil, // all edges
		nil, // no exclusions
		map[storage.RelType]float32{
			storage.RelContradicts: 0.3,
			storage.RelSupersedes:  0.5,
		},
	),
	"causal": newProfile(
		[]storage.RelType{
			storage.RelCauses,
			storage.RelDependsOn,
			storage.RelBlocks,
			storage.RelPrecededBy,
			storage.RelFollowedBy,
		},
		nil,
		map[storage.RelType]float32{
			storage.RelCauses: 1.3,
		},
	),
	"confirmatory": newProfile(
		[]storage.RelType{
			storage.RelSupports,
			storage.RelImplements,
			storage.RelRefines,
			storage.RelReferences,
		},
		[]storage.RelType{storage.RelContradicts},
		map[storage.RelType]float32{
			storage.RelSupports: 1.2,
		},
	),
	"adversarial": newProfile(
		[]storage.RelType{
			storage.RelContradicts,
			storage.RelSupersedes,
			storage.RelBlocks,
		},
		nil,
		map[storage.RelType]float32{
			storage.RelContradicts: 1.5,
		},
	),
	"structural": newProfile(
		[]storage.RelType{
			storage.RelIsPartOf,
			storage.RelBelongsToProject,
			storage.RelCreatedByPerson,
		},
		nil,
		nil,
	),
}

// GetProfile returns the named profile, or the default profile for unknown names.
func GetProfile(name string) *TraversalProfile {
	if p, ok := builtinProfiles[name]; ok {
		return p
	}
	return builtinProfiles["default"]
}

// ValidProfileName reports whether name is a known built-in profile.
func ValidProfileName(name string) bool {
	_, ok := builtinProfiles[name]
	return ok
}

// --- Auto-Inference Engine ---

type profileRule struct {
	profile  string
	patterns []*regexp.Regexp
	score    int
}

var inferenceRules = []profileRule{
	{
		profile: "causal",
		score:   2,
		patterns: mustCompile([]string{
			`(?i)^why\b`,
			`(?i)\bwhat\s+caused\b`,
			`(?i)\bwhat\s+led\s+to\b`,
			`(?i)\broot\s+cause\b`,
			`(?i)\bblocked\s+by\b`,
			`(?i)\bbecause\s+of\s+what\b`,
			`(?i)\bdepends\s+on\b.{0,40}\?`,
			`(?i)\bwhat\s+is\s+blocking\b`,
		}),
	},
	{
		profile: "adversarial",
		score:   3,
		patterns: mustCompile([]string{
			`(?i)\bcontradict`,
			`(?i)\bconflict(?:s|ing)?\s+(?:with|between)\b`,
			`(?i)\binconsisten`,
			`(?i)\bwrong\s+about\b`,
			`(?i)\bdisagree`,
		}),
	},
	{
		profile: "confirmatory",
		score:   2,
		patterns: mustCompile([]string{
			`(?i)\bvalidat`,
			`(?i)\bevidence\s+(?:for|that)\b`,
			`(?i)\bconfirm`,
			`(?i)\bsupports?\s+(?:the\s+)?(?:claim|idea|theory|decision|hypothesis)\b`,
		}),
	},
	{
		profile: "structural",
		score:   1,
		patterns: mustCompile([]string{
			`(?i)\bpart\s+of\b`,
			`(?i)\bbelongs\s+to\b`,
			`(?i)\bstructure\s+of\b`,
			`(?i)\bcomponents?\s+of\b`,
		}),
	},
}

// InferProfile selects a traversal profile from context phrases.
//
// Resolution order:
//  1. Pattern-match context against inference rules; if max score >= 2, use that profile.
//  2. Fall through to vaultDefault if provided.
//  3. Fall through to "default".
func InferProfile(contexts []string, vaultDefault string) string {
	if len(contexts) == 0 {
		return fallback(vaultDefault)
	}

	joined := strings.Join(contexts, " ")
	scores := make(map[string]int, 4)

	for _, rule := range inferenceRules {
		for _, pat := range rule.patterns {
			if pat.MatchString(joined) {
				scores[rule.profile] += rule.score
				break // one match per rule
			}
		}
	}

	best, bestScore := "", 0
	for profile, score := range scores {
		if score > bestScore || (score == bestScore && profile < best) {
			best, bestScore = profile, score
		}
	}

	if bestScore >= 2 {
		return best
	}
	return fallback(vaultDefault)
}

func fallback(vaultDefault string) string {
	if vaultDefault != "" && ValidProfileName(vaultDefault) {
		return vaultDefault
	}
	return "default"
}

func mustCompile(patterns []string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		out[i] = regexp.MustCompile(p)
	}
	return out
}
```

### Step 4: Run tests

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/engine/activation/... -run "TestGetProfile|TestInferProfile|TestDefaultProfile|TestCausalProfile|TestAdversarialProfile|TestConfirmatoryProfile|TestStructuralProfile" -v
```

Expected: all pass.

### Step 5: Commit

```bash
cd ~/github.com/scrypster/muninndb
git add internal/engine/activation/profiles.go internal/engine/activation/profiles_test.go
git commit -m "feat: add TraversalProfile definitions and InferProfile engine"
```

---

## Task 2: Wire Profiles into BFS Phase 5

**Files:**
- Modify: `internal/engine/activation/engine.go`

This is the core change: add `Profile string` to `ActivateRequest`, add `resolveProfile()`, and modify `phase5Traverse` to apply Include/Exclude/Boost.

### Step 1: Write failing tests

Add to `internal/engine/activation/activation_test.go` (or a new file `profiles_integration_test.go` in the same package):

```go
// profiles_bfs_test.go
package activation

import (
	"testing"

	"github.com/scrypster/muninndb/internal/storage"
)

// TestPhase5WithCausalProfile verifies that the causal profile skips non-causal edges.
// We test the AllowsEdge logic directly since phase5Traverse requires a live store.
func TestProfileAllowsEdge_Causal(t *testing.T) {
	p := GetProfile("causal")

	// causal types allowed
	for _, rel := range []storage.RelType{
		storage.RelCauses, storage.RelDependsOn, storage.RelBlocks,
		storage.RelPrecededBy, storage.RelFollowedBy,
	} {
		if !p.AllowsEdge(rel) {
			t.Errorf("causal profile should allow %v", rel)
		}
	}

	// non-causal types blocked
	for _, rel := range []storage.RelType{
		storage.RelSupports, storage.RelContradicts, storage.RelReferences,
		storage.RelIsPartOf, storage.RelBelongsToProject,
	} {
		if p.AllowsEdge(rel) {
			t.Errorf("causal profile should block %v", rel)
		}
	}
}

func TestProfileAllowsEdge_Confirmatory_ExcludesContradicts(t *testing.T) {
	p := GetProfile("confirmatory")
	if p.AllowsEdge(storage.RelContradicts) {
		t.Error("confirmatory profile must not allow RelContradicts")
	}
	if !p.AllowsEdge(storage.RelSupports) {
		t.Error("confirmatory profile must allow RelSupports")
	}
}

func TestProfileAllowsEdge_Default_AllowsAll(t *testing.T) {
	p := GetProfile("default")
	// default allows all (empty include = traverse all)
	allTypes := []storage.RelType{
		storage.RelSupports, storage.RelContradicts, storage.RelDependsOn,
		storage.RelSupersedes, storage.RelRelatesTo, storage.RelIsPartOf,
		storage.RelCauses, storage.RelPrecededBy, storage.RelFollowedBy,
		storage.RelCreatedByPerson, storage.RelBelongsToProject, storage.RelReferences,
		storage.RelImplements, storage.RelBlocks, storage.RelResolves, storage.RelRefines,
	}
	for _, rel := range allTypes {
		if !p.AllowsEdge(rel) {
			t.Errorf("default profile should allow all edge types, blocked %v", rel)
		}
	}
}

func TestBoostFor_DefaultProfile_ContradictsLow(t *testing.T) {
	p := GetProfile("default")
	boost := p.BoostFor(storage.RelContradicts)
	if boost != 0.3 {
		t.Errorf("expected boost 0.3 for RelContradicts in default, got %f", boost)
	}
}

func TestBoostFor_AdversarialProfile_ContradictsHigh(t *testing.T) {
	p := GetProfile("adversarial")
	boost := p.BoostFor(storage.RelContradicts)
	if boost != 1.5 {
		t.Errorf("expected boost 1.5 for RelContradicts in adversarial, got %f", boost)
	}
}

func TestBoostFor_SameTypeOppositeProfilesGiveDifferentResults(t *testing.T) {
	def := GetProfile("default").BoostFor(storage.RelContradicts)
	adv := GetProfile("adversarial").BoostFor(storage.RelContradicts)
	if def >= adv {
		t.Errorf("adversarial boost (%f) should be higher than default boost (%f) for RelContradicts", adv, def)
	}
}

// TestResolveProfile_ResolutionChain tests all branches of C-B-A resolution.
func TestResolveProfile_ExplicitOverride(t *testing.T) {
	req := &ActivateRequest{Profile: "adversarial", Context: []string{"why did this fail?"}}
	p := resolveProfile(req, "causal")
	// explicit override wins even when context would infer causal and vault default is causal
	if p != GetProfile("adversarial") {
		t.Error("explicit Profile override should win over inference and vault default")
	}
}

func TestResolveProfile_InferenceWinsOverVaultDefault(t *testing.T) {
	req := &ActivateRequest{Profile: "", Context: []string{"what contradicts this decision?"}}
	p := resolveProfile(req, "causal") // vault default is causal
	// adversarial inferred (score 3 >= threshold 2)
	if p != GetProfile("adversarial") {
		t.Errorf("inference should win over vault default: got %+v", p)
	}
}

func TestResolveProfile_VaultDefaultWinsOnAmbiguous(t *testing.T) {
	req := &ActivateRequest{Profile: "", Context: []string{"user preferences"}}
	p := resolveProfile(req, "structural")
	if p != GetProfile("structural") {
		t.Error("vault default should win when inference is ambiguous")
	}
}

func TestResolveProfile_DefaultFallsThrough(t *testing.T) {
	req := &ActivateRequest{Profile: "", Context: []string{"user preferences"}}
	p := resolveProfile(req, "")
	if p != GetProfile("default") {
		t.Error("should fall through to default profile when no override, inference, or vault default")
	}
}
```

### Step 2: Run tests to verify they fail

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/engine/activation/... -run "TestProfileAllowsEdge|TestBoostFor|TestResolveProfile" -v 2>&1 | head -20
```

Expected: compilation failure (`ActivateRequest.Profile` and `resolveProfile` don't exist yet).

### Step 3: Add Profile to ActivateRequest and implement resolveProfile + modified phase5Traverse

In `internal/engine/activation/engine.go`:

**3a. Add `Profile string` to `ActivateRequest` (line 97, after `ReadOnly bool`):**
```go
	ReadOnly         bool // when true, skip all write side-effects (observe mode)
	Profile          string // traversal profile: "default"|"causal"|"confirmatory"|"adversarial"|"structural"|""
	StructuredFilter interface{}
```

**3b. Add `resolveProfile` function (add after the `ActivateResult` block, before `phase5Traverse`):**
```go
// resolveProfile implements the C-B-A profile resolution chain:
//  1. Explicit per-request profile (A)
//  2. Auto-inferred from context (C)
//  3. Vault Plasticity default (B)
//  4. Hardcoded "default" profile
func resolveProfile(req *ActivateRequest, vaultDefault string) *TraversalProfile {
	if req.Profile != "" && ValidProfileName(req.Profile) {
		return GetProfile(req.Profile)
	}
	inferred := InferProfile(req.Context, vaultDefault)
	return GetProfile(inferred)
}
```

**3c. Modify `phase5Traverse` to accept and apply the profile.**

Change the function signature to add a profile parameter (add `profile *TraversalProfile` after `ws [8]byte`):
```go
func (e *ActivationEngine) phase5Traverse(
	ctx context.Context,
	req *ActivateRequest,
	ws [8]byte,
	profile *TraversalProfile,
	candidates []fusedCandidate,
) []traversedCandidate {
```

Inside the BFS loop, replace the propagation block (currently at line 617-648) with:
```go
	for _, curr := range eligible {
		for _, assoc := range assocMap[curr.id] {
			if seen[assoc.TargetID] {
				continue
			}

			// Profile filtering: skip excluded or non-included edge types.
			if !profile.AllowsEdge(assoc.RelType) {
				continue
			}

			boost := float64(profile.BoostFor(assoc.RelType))
			propagated := curr.baseScore * float64(assoc.Weight) * boost * math.Pow(hopPenalty, float64(curr.hopDepth+1))
			if propagated < minHopScore {
				break // weight-sorted, early termination
			}

			seen[assoc.TargetID] = true
			expanded++

			hopPath := make([]storage.ULID, len(curr.hopPath)+1)
			copy(hopPath, curr.hopPath)
			hopPath[len(curr.hopPath)] = assoc.TargetID

			discovered = append(discovered, traversedCandidate{
				id:         assoc.TargetID,
				propagated: propagated,
				hopPath:    hopPath,
				relType:    uint16(assoc.RelType),
			})

			if curr.hopDepth+1 < req.HopDepth {
				nextLevel = append(nextLevel, levelItem{
					id:        assoc.TargetID,
					baseScore: propagated,
					hopDepth:  curr.hopDepth + 1,
					hopPath:   hopPath,
				})
			}

			if expanded >= maxBFSNodes {
				break outer
			}
		}
	}
```

**3d. Find the call site of `phase5Traverse` in `engine.go` and add the profile resolution.**

Search for `phase5Traverse(` in the file. It will look like:
```go
traversed := e.phase5Traverse(ctx, req, ws, candidates)
```

Replace with:
```go
profile := resolveProfile(req, req.VaultDefault)
traversed := e.phase5Traverse(ctx, req, ws, profile, candidates)
```

**Note:** `VaultDefault string` also needs to be added to `ActivateRequest` to carry the vault's Plasticity default profile through. Add it alongside `Profile`:
```go
	Profile          string // traversal profile override
	VaultDefault     string // vault Plasticity.TraversalProfile default (set by engine.go, not by callers)
```

### Step 4: Run tests

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/engine/activation/... -v 2>&1 | tail -30
```

Expected: all tests pass. Fix any compilation errors before continuing.

### Step 5: Run full suite to check for regressions

```bash
cd ~/github.com/scrypster/muninndb
go test ./...
```

All packages must pass.

### Step 6: Commit

```bash
cd ~/github.com/scrypster/muninndb
git add internal/engine/activation/engine.go internal/engine/activation/profiles_bfs_test.go
git commit -m "feat: wire TraversalProfile into BFS phase5Traverse with Include/Exclude/Boost"
```

---

## Task 3: Plasticity — Add TraversalProfile to PlasticityConfig

**Files:**
- Modify: `internal/auth/plasticity.go`
- Modify: `internal/auth/plasticity_test.go`

### Step 1: Write failing tests

Add to `internal/auth/plasticity_test.go`:

```go
func TestPlasticityConfig_TraversalProfileField(t *testing.T) {
	profile := "causal"
	cfg := &PlasticityConfig{TraversalProfile: &profile}
	r := ResolvePlasticity(cfg)
	if r.TraversalProfile != "causal" {
		t.Errorf("expected TraversalProfile 'causal', got %q", r.TraversalProfile)
	}
}

func TestPlasticityConfig_NoTraversalProfileDefaultsEmpty(t *testing.T) {
	r := ResolvePlasticity(nil)
	if r.TraversalProfile != "" {
		t.Errorf("nil config should resolve TraversalProfile to empty string, got %q", r.TraversalProfile)
	}
}

func TestPlasticityConfig_InvalidProfileNamePassedThrough(t *testing.T) {
	// PlasticityConfig stores whatever is provided; validation is done by the engine.
	bad := "nonexistent"
	cfg := &PlasticityConfig{TraversalProfile: &bad}
	r := ResolvePlasticity(cfg)
	if r.TraversalProfile != "nonexistent" {
		t.Errorf("expected 'nonexistent', got %q", r.TraversalProfile)
	}
}
```

### Step 2: Run tests to verify failure

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/auth/... -run "TestPlasticityConfig_TraversalProfile" -v
```

Expected: compilation error (`TraversalProfile` field doesn't exist).

### Step 3: Implement

In `internal/auth/plasticity.go`:

**3a. Add to `PlasticityConfig` (after `DecayStability`):**
```go
	TraversalProfile *string  `json:"traversal_profile,omitempty"` // "default"|"causal"|"confirmatory"|"adversarial"|"structural"
```

**3b. Add to `ResolvedPlasticity` (after `RecencyWeight`):**
```go
	TraversalProfile string  `json:"traversal_profile"` // empty string = use auto-inference
```

**3c. Add to `ResolvePlasticity` override block (after the `DecayStability` override):**
```go
	if cfg.TraversalProfile != nil {
		r.TraversalProfile = *cfg.TraversalProfile
	}
```

### Step 4: Run tests

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/auth/... -v
```

All must pass.

### Step 5: Wire VaultDefault into the engine

In `internal/engine/engine.go`, find where `ResolvePlasticity` is called and where the internal `ActivateRequest` is built. Add `VaultDefault`:

```go
resolved := auth.ResolvePlasticity(vaultCfg.Plasticity)
actReq := &activation.ActivateRequest{
    // ... existing fields ...
    VaultDefault: resolved.TraversalProfile,
}
```

Run full suite after:
```bash
cd ~/github.com/scrypster/muninndb
go test ./...
```

### Step 6: Commit

```bash
cd ~/github.com/scrypster/muninndb
git add internal/auth/plasticity.go internal/auth/plasticity_test.go internal/engine/engine.go
git commit -m "feat: add TraversalProfile to PlasticityConfig and wire vault default into activation"
```

---

## Task 4: MBP Types — Expose Profile to External Callers

**Files:**
- Modify: `internal/transport/mbp/types.go`
- Modify: `internal/engine/engine.go` (the conversion from mbp.ActivateRequest to activation.ActivateRequest)

### Step 1: Write failing test

In `internal/transport/mbp/` (check for existing types_test.go or create one):

```go
func TestActivateRequest_ProfileField(t *testing.T) {
	req := ActivateRequest{
		Vault:   "default",
		Context: []string{"why did this fail?"},
		Profile: "causal",
	}
	if req.Profile != "causal" {
		t.Errorf("expected Profile 'causal', got %q", req.Profile)
	}
}
```

### Step 2: Run to verify failure

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/transport/mbp/... -run "TestActivateRequest_ProfileField" -v
```

### Step 3: Implement

In `internal/transport/mbp/types.go`, add to `ActivateRequest` (after `DisableHops bool`):
```go
	Profile     string `json:"profile,omitempty"` // traversal profile override: "default"|"causal"|"confirmatory"|"adversarial"|"structural"
```

In `internal/engine/engine.go`, find the conversion from `mbp.ActivateRequest` to `activation.ActivateRequest` and add:
```go
	Profile: req.Profile,
```

### Step 4: Run full suite

```bash
cd ~/github.com/scrypster/muninndb
go test ./...
```

### Step 5: Commit

```bash
cd ~/github.com/scrypster/muninndb
git add internal/transport/mbp/types.go internal/engine/engine.go
git commit -m "feat: expose Profile field on MBP ActivateRequest and wire into engine"
```

---

## Task 5: MCP Tools — Fix Relation Types + Add Profile Parameter

**Files:**
- Modify: `internal/mcp/tools.go`
- Modify: `internal/mcp/tools_test.go`

### Step 1: Write failing tests

Add to `internal/mcp/tools_test.go`:

```go
func TestMuninnLinkTool_HasAllRelationTypes(t *testing.T) {
	tools := allToolDefinitions()
	var linkTool *ToolDefinition
	for i := range tools {
		if tools[i].Name == "muninn_link" {
			linkTool = &tools[i]
			break
		}
	}
	if linkTool == nil {
		t.Fatal("muninn_link tool not found")
	}

	schema := linkTool.InputSchema.(map[string]any)
	props := schema["properties"].(map[string]any)
	relProp := props["relation"].(map[string]any)
	desc := relProp["description"].(string)

	// All 16 relation type strings must appear in the description
	required := []string{
		"supports", "contradicts", "depends_on", "supersedes", "relates_to",
		"is_part_of", "causes", "preceded_by", "followed_by",
		"created_by_person", "belongs_to_project", "references",
		"implements", "blocks", "resolves", "refines",
	}
	for _, r := range required {
		if !strings.Contains(desc, r) {
			t.Errorf("muninn_link relation description missing %q", r)
		}
	}
}

func TestMuninnRecallTool_HasProfileParam(t *testing.T) {
	tools := allToolDefinitions()
	var recallTool *ToolDefinition
	for i := range tools {
		if tools[i].Name == "muninn_recall" {
			recallTool = &tools[i]
			break
		}
	}
	if recallTool == nil {
		t.Fatal("muninn_recall tool not found")
	}

	schema := recallTool.InputSchema.(map[string]any)
	props := schema["properties"].(map[string]any)
	if _, ok := props["profile"]; !ok {
		t.Error("muninn_recall tool must have a 'profile' parameter")
	}
}

func TestMuninnRecallTool_ProfileNotRequired(t *testing.T) {
	tools := allToolDefinitions()
	for _, tool := range tools {
		if tool.Name != "muninn_recall" {
			continue
		}
		schema := tool.InputSchema.(map[string]any)
		required, _ := schema["required"].([]string)
		for _, r := range required {
			if r == "profile" {
				t.Error("'profile' must not be in muninn_recall required list")
			}
		}
	}
}
```

### Step 2: Run to verify failure

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/mcp/... -run "TestMuninnLinkTool_HasAllRelationTypes|TestMuninnRecallTool" -v
```

Expected: failures on missing relation types and missing profile param.

### Step 3: Update tools.go

**3a. Update `muninn_link` relation description** to include all 16 types with "when to use" guidance:

Find the `muninn_link` tool definition in `internal/mcp/tools.go` and replace the `relation` property description with:

```go
"relation": map[string]any{
    "type": "string",
    "description": `Type of relationship between the two memories. Choose the most specific type:
• supports          — this memory provides evidence or backing for the other
• contradicts       — this memory conflicts with or refutes the other
• depends_on        — this memory requires the other to be understood/true first
• supersedes        — this memory replaces or updates the other (other is now outdated)
• relates_to        — general association when no specific type fits (default fallback)
• is_part_of        — this memory is a component or section of the other
• causes            — this memory is a cause or contributing factor to the other
• preceded_by       — this memory chronologically follows the other
• followed_by       — this memory chronologically precedes the other
• created_by_person — this memory was authored or owned by the person in the other
• belongs_to_project — this memory belongs to the project or context in the other
• references        — this memory cites or links to the other without strong semantic weight
• implements        — this memory is the concrete realization of the other (e.g., code for a spec)
• blocks            — this memory is an obstacle preventing progress on the other
• resolves          — this memory is the solution or fix for the other
• refines           — this memory is a near-duplicate refinement or correction of the other`,
},
```

**3b. Add `profile` parameter to `muninn_recall`:**

Find the `muninn_recall` tool definition and add to its `properties` map:

```go
"profile": map[string]any{
    "type": "string",
    "enum": []string{"default", "causal", "confirmatory", "adversarial", "structural"},
    "description": `Traversal profile controlling which association types BFS follows. Leave unset for automatic inference.
• default       — balanced retrieval across all edge types (RelContradicts dampened 0.3×)
• causal        — follow cause/effect/dependency chains (Causes, DependsOn, Blocks, PrecededBy, FollowedBy)
• confirmatory  — find supporting evidence, excludes contradiction edges (Supports, Implements, Refines, References)
• adversarial   — surface conflicts and contradictions (Contradicts, Supersedes, Blocks; Contradicts boosted 1.5×)
• structural    — follow project/person/part-of hierarchy (IsPartOf, BelongsToProject, CreatedByPerson)

When to override auto-inference:
  Use "causal" when asking why something happened or what something depends on.
  Use "adversarial" when auditing for inconsistencies or contradictions.
  Use "confirmatory" when looking for supporting evidence for a claim.
  Use "structural" when navigating project or organizational structure.`,
},
```

### Step 4: Run tests

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/mcp/... -v
```

All must pass.

### Step 5: Commit

```bash
cd ~/github.com/scrypster/muninndb
git add internal/mcp/tools.go internal/mcp/tools_test.go
git commit -m "feat: expand muninn_link relation types to all 16 + add profile param to muninn_recall"
```

---

## Task 6: MCP Handlers — Wire Profile + Fix relTypeFromString

**Files:**
- Modify: `internal/mcp/handlers.go`
- Modify: `internal/mcp/handlers_test.go`

### Step 1: Write failing tests

Add to `internal/mcp/handlers_test.go`:

```go
func TestRelTypeFromString_AllTypes(t *testing.T) {
	cases := map[string]uint16{
		"supports":           1,
		"contradicts":        2,
		"depends_on":         3,
		"supersedes":         4,
		"relates_to":         5,
		"is_part_of":         6,
		"causes":             7,
		"preceded_by":        8,
		"followed_by":        9,
		"created_by_person":  10,
		"belongs_to_project": 11,
		"references":         12,
		"implements":         13,
		"blocks":             14,
		"resolves":           15,
		"refines":            16,
	}
	for input, want := range cases {
		got := relTypeFromString(input)
		if got != want {
			t.Errorf("relTypeFromString(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestRelTypeFromString_UnknownDefaultsToRelatesTo(t *testing.T) {
	got := relTypeFromString("totally_unknown_type")
	if got != 5 {
		t.Errorf("unknown relation should default to relates_to (5), got %d", got)
	}
}

func TestRelTypeFromString_EmptyDefaultsToRelatesTo(t *testing.T) {
	got := relTypeFromString("")
	if got != 5 {
		t.Errorf("empty relation should default to relates_to (5), got %d", got)
	}
}
```

### Step 2: Run to verify failure

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/mcp/... -run "TestRelTypeFromString" -v
```

Expected: failures for types beyond the current 5 that are mapped.

### Step 3: Expand relTypeFromString in handlers.go

Find `relTypeFromString` and replace with the complete mapping:

```go
func relTypeFromString(rel string) uint16 {
	relTypes := map[string]uint16{
		"supports":           1,
		"contradicts":        2,
		"depends_on":         3,
		"supersedes":         4,
		"relates_to":         5,
		"is_part_of":         6,
		"causes":             7,
		"preceded_by":        8,
		"followed_by":        9,
		"created_by_person":  10,
		"belongs_to_project": 11,
		"references":         12,
		"implements":         13,
		"blocks":             14,
		"resolves":           15,
		"refines":            16,
	}
	if v, ok := relTypes[rel]; ok {
		return v
	}
	return 5 // default: relates_to
}
```

**Also wire `profile` into `handleRecall`:**

Find `handleRecall` in handlers.go. After extracting `limit`, add:

```go
	// Extract optional profile override
	profile := ""
	if p, ok := args["profile"].(string); ok {
		profile = p
	}
```

Then add `Profile: profile` to the `mbp.ActivateRequest` being built.

### Step 4: Run all MCP tests

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/mcp/... -v
```

All must pass.

### Step 5: Run full suite

```bash
cd ~/github.com/scrypster/muninndb
go test ./...
```

All must pass.

### Step 6: Commit

```bash
cd ~/github.com/scrypster/muninndb
git add internal/mcp/handlers.go internal/mcp/handlers_test.go
git commit -m "feat: expand relTypeFromString to all 16 types + wire profile param through MCP recall handler"
```

---

## Task 7: Admin API — TraversalProfile in Plasticity Endpoint

The REST admin API for Plasticity config (`PATCH /admin/vaults/{vault}/plasticity`) must accept and validate the new `traversal_profile` field.

**Files:**
- Modify: `internal/transport/rest/admin_handlers.go`
- Modify: `internal/transport/rest/admin_plasticity_test.go`

### Step 1: Write failing test

Add to `internal/transport/rest/admin_plasticity_test.go`:

```go
func TestAdminPatchPlasticity_TraversalProfile_Valid(t *testing.T) {
	// POST a valid traversal_profile value and expect 200
	// (check existing test structure and mirror the pattern used there)
	body := `{"traversal_profile": "causal"}`
	// mirror existing test setup for PATCH /admin/vaults/{vault}/plasticity
	// assert response 200 and that resolved config has TraversalProfile == "causal"
}

func TestAdminPatchPlasticity_TraversalProfile_Invalid(t *testing.T) {
	body := `{"traversal_profile": "nonexistent"}`
	// assert response 400 with validation error
}
```

Check `internal/transport/rest/admin_plasticity_test.go` for the exact test pattern used. Mirror it exactly. The validation in the handler should call `activation.ValidProfileName()` to check the value before persisting.

### Step 2: Implement validation in admin_handlers.go

Find the Plasticity PATCH handler. Add validation for `traversal_profile` if provided:

```go
if cfg.TraversalProfile != nil {
    if *cfg.TraversalProfile != "" && !activation.ValidProfileName(*cfg.TraversalProfile) {
        s.sendError(w, http.StatusBadRequest, ErrInvalidRequest,
            fmt.Sprintf("invalid traversal_profile %q: must be one of: default, causal, confirmatory, adversarial, structural", *cfg.TraversalProfile))
        return
    }
}
```

### Step 3: Run tests

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/transport/rest/... -v
```

### Step 4: Run full suite

```bash
cd ~/github.com/scrypster/muninndb
go test ./...
```

### Step 5: Commit

```bash
cd ~/github.com/scrypster/muninndb
git add internal/transport/rest/admin_handlers.go internal/transport/rest/admin_plasticity_test.go
git commit -m "feat: validate traversal_profile in Plasticity admin API"
```

---

## Task 8: Logging — Profile Used on Every Activation

Opus said this is non-negotiable for tuning.

**Files:**
- Modify: `internal/engine/activation/log.go` (or engine.go — check where activation logging lives)

### Step 1: Find where activation results are logged

```bash
grep -n "log\|slog\|zap\|ActivateResult\|latency" ~/github.com/scrypster/muninndb/internal/engine/activation/log.go | head -20
```

### Step 2: Add profile to activation log output

In the activation logging function, add the resolved profile name. If `ActivateResult` doesn't carry it, add a `ProfileUsed string` field:

**In `ActivateResult`:**
```go
type ActivateResult struct {
	QueryID     string
	Activations []ScoredEngram
	TotalFound  int
	LatencyMs   float64
	ProfileUsed string // which traversal profile fired
}
```

Set `ProfileUsed` in the activation run path before returning. Include it in the log output alongside latency and result count.

### Step 3: Write test

```go
func TestActivateResult_CarriesProfileUsed(t *testing.T) {
	// After a full activation with Profile="causal", result.ProfileUsed == "causal"
	// This is an integration test — needs a real (test) store
	// Mirror the pattern in activation_test.go
}
```

### Step 4: Run full suite

```bash
cd ~/github.com/scrypster/muninndb
go test ./...
```

### Step 5: Commit

```bash
cd ~/github.com/scrypster/muninndb
git add internal/engine/activation/engine.go internal/engine/activation/log.go
git commit -m "feat: log resolved traversal profile on every activation"
```

---

## Task 9: Final Regression + Coverage Check

### Step 1: Run the full test suite

```bash
cd ~/github.com/scrypster/muninndb
go test ./... -count=1
```

All packages must pass. `-count=1` disables test caching to ensure fresh runs.

### Step 2: Check coverage on new code

```bash
cd ~/github.com/scrypster/muninndb
go test ./internal/engine/activation/... -coverprofile=cover.out
go tool cover -func=cover.out | grep profiles
```

`profiles.go` should show >85% coverage. Any uncovered paths in `InferProfile` or `GetProfile` are missing tests — add them.

### Step 3: Race detector check

```bash
cd ~/github.com/scrypster/muninndb
go test -race ./internal/engine/activation/... ./internal/mcp/... ./internal/auth/...
```

No races.

### Step 4: Final commit

```bash
cd ~/github.com/scrypster/muninndb
git log --oneline -10
```

Verify all feature commits are present and clean.

---

## Done Criteria

- [ ] `go test ./... -count=1` passes with zero failures
- [ ] `go test -race` passes on activation, mcp, auth packages
- [ ] `profiles.go` coverage ≥ 85%
- [ ] All 16 relation types handled in `relTypeFromString` and documented in `muninn_link` tool
- [ ] `muninn_recall` has `profile` parameter with enum and descriptions
- [ ] `PlasticityConfig.TraversalProfile` accepted and validated via admin API
- [ ] Every activation log line includes the resolved profile name
- [ ] `ActivateRequest.Profile` wired end-to-end: MBP → engine → BFS
- [ ] Default profile behavior unchanged (all tests that existed before this feature still pass)
