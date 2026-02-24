# Bible Eval Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a reproducible two-phase eval harness using the KJV Bible and public-domain cross-references as ground truth, proving retrieval quality (NDCG@10) and demonstrating that Ebbinghaus decay and Hebbian co-activation measurably improve results.

**Architecture:** A standalone binary at `cmd/eval-bible/` mirrors the structure of `cmd/eval/` (same engine init pattern, same mbp types). Phase 1 queries 100 seed verses and scores against known cross-references. Phase 2 simulates a reading session + decay and measures NDCG delta. A shell setup script downloads the corpus data once; Makefile targets make it reproducible.

**Tech Stack:** Go 1.22+, KJV Bible JSON (scrollmapper/bible_databases, public domain), OpenBible cross-references TSV (CC0), local ONNX embedder (all-MiniLM-L6-v2 INT8), MuninnDB engine (same init as `cmd/eval/`).

---

## Reference: How the existing eval works

Read `cmd/eval/main.go` lines 1-200 and `cmd/eval/adapters.go` before starting. Key patterns to reuse:
- Engine init: `storage.OpenPebble`, `storage.NewPebbleStore`, `fts.New`, `hnswpkg.NewRegistry`, `embedpkg.NewEmbedService("local://all-MiniLM-L6-v2")`, `activation.New`, `engine.NewEngine`
- Write pattern: `mem.Vault = vault; eng.Write(ctx, &mem)` then `hnswReg.Insert(ctx, ws, id, vec)`
- `ws := store.ResolveVaultPrefix(vault)` — compute vault prefix once

The Bible eval uses `vault = "bible"` throughout.

---

## Task 1: Setup Script + Data Download

**Files:**
- Create: `scripts/eval-bible-setup.sh`
- Create: `testdata/bible/.gitkeep`
- Modify: `.gitignore`

**Step 1: Create the testdata directory placeholder**

```bash
mkdir -p testdata/bible
touch testdata/bible/.gitkeep
```

**Step 2: Add to `.gitignore`**

Append to the existing `.gitignore`:
```
# Bible eval corpus data (downloaded by scripts/eval-bible-setup.sh)
testdata/bible/kjv.json
testdata/bible/cross-refs.tsv
```

**Step 3: Write `scripts/eval-bible-setup.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_DIR="$SCRIPT_DIR/../testdata/bible"

mkdir -p "$DATA_DIR"

# Download KJV Bible JSON (scrollmapper/bible_databases, public domain)
KJV_URL="https://raw.githubusercontent.com/scrollmapper/bible_databases/master/json/t_kjv.json"
KJV_FILE="$DATA_DIR/kjv.json"
if [ ! -f "$KJV_FILE" ]; then
    echo "Downloading KJV Bible JSON..."
    curl -fsSL "$KJV_URL" -o "$KJV_FILE"
    echo "  → saved to $KJV_FILE ($(du -h "$KJV_FILE" | cut -f1))"
else
    echo "KJV Bible JSON already present: $KJV_FILE"
fi

# Download OpenBible cross-references TSV (CC0)
XREF_URL="https://a.openbible.info/refs/cross-references.tsv"
XREF_FILE="$DATA_DIR/cross-refs.tsv"
if [ ! -f "$XREF_FILE" ]; then
    echo "Downloading OpenBible cross-references..."
    curl -fsSL "$XREF_URL" -o "$XREF_FILE"
    echo "  → saved to $XREF_FILE ($(du -h "$XREF_FILE" | cut -f1))"
else
    echo "Cross-references already present: $XREF_FILE"
fi

echo ""
echo "Setup complete. Data files:"
ls -lh "$DATA_DIR"/*.json "$DATA_DIR"/*.tsv 2>/dev/null || true
echo ""
echo "Run the eval with: make eval-bible"
```

**Step 4: Make executable and test**

```bash
chmod +x scripts/eval-bible-setup.sh
./scripts/eval-bible-setup.sh
```

Expected: both files download, sizes ~4MB (kjv.json) and ~8MB (cross-refs.tsv).

**Step 5: Commit**

```bash
git add scripts/eval-bible-setup.sh testdata/bible/.gitkeep .gitignore
git commit -m "feat(eval-bible): setup script and data directory"
```

---

## Task 2: Corpus Loader

Parse KJV JSON into `[]mbp.WriteRequest` with appropriate tags.

**Files:**
- Create: `cmd/eval-bible/corpus.go`
- Create: `cmd/eval-bible/corpus_test.go`

**KJV JSON format** (scrollmapper t_kjv.json):
```json
[
  {"b": 1, "c": 1, "v": 1, "t": "In the beginning God created the heaven and the earth."},
  ...
]
```
Fields: `b` = book number (1-66), `c` = chapter, `v` = verse, `t` = text.

**Step 1: Write `corpus_test.go` first**

```go
package main

import (
    "encoding/json"
    "testing"
)

func TestParseKJV(t *testing.T) {
    raw := []kjvRecord{
        {B: 43, C: 3, V: 16, T: "For God so loved the world..."},
        {B: 1, C: 1, V: 1, T: "In the beginning..."},
    }
    data, _ := json.Marshal(raw)

    reqs, err := parseKJV(data, false /* ntOnly */)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(reqs) != 2 {
        t.Fatalf("want 2 records, got %d", len(reqs))
    }

    // Check John 3:16
    john := reqs[0]
    if john.Concept != "John 3:16" {
        t.Errorf("want concept 'John 3:16', got %q", john.Concept)
    }
    if john.Content != "For God so loved the world..." {
        t.Errorf("unexpected content: %q", john.Content)
    }
    assertTag(t, john.Tags, "New Testament")
    assertTag(t, john.Tags, "John")
    assertTag(t, john.Tags, "gospel")
}

func TestParseKJV_NTOnly(t *testing.T) {
    raw := []kjvRecord{
        {B: 1, C: 1, V: 1, T: "In the beginning..."},  // Genesis (OT)
        {B: 40, C: 1, V: 1, T: "The book of..."},       // Matthew (NT)
    }
    data, _ := json.Marshal(raw)

    reqs, err := parseKJV(data, true /* ntOnly */)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(reqs) != 1 {
        t.Fatalf("want 1 NT record, got %d", len(reqs))
    }
    if reqs[0].Concept != "Matthew 1:1" {
        t.Errorf("want 'Matthew 1:1', got %q", reqs[0].Concept)
    }
}

func TestVerseRef(t *testing.T) {
    if got := verseRef(43, 3, 16); got != "John 3:16" {
        t.Errorf("want 'John 3:16', got %q", got)
    }
    if got := verseRef(1, 1, 1); got != "Genesis 1:1" {
        t.Errorf("want 'Genesis 1:1', got %q", got)
    }
}

func assertTag(t *testing.T, tags []string, want string) {
    t.Helper()
    for _, tag := range tags {
        if tag == want {
            return
        }
    }
    t.Errorf("tag %q not found in %v", want, tags)
}
```

**Step 2: Run test to confirm it fails**

```bash
cd cmd/eval-bible && go test ./... 2>&1 | head -20
```

Expected: FAIL — corpus.go doesn't exist yet.

**Step 3: Write `corpus.go`**

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/scrypster/muninndb/internal/transport/mbp"
)

// kjvRecord is one verse from the scrollmapper t_kjv.json format.
type kjvRecord struct {
    B int    `json:"b"` // book number 1-66
    C int    `json:"c"` // chapter
    V int    `json:"v"` // verse
    T string `json:"t"` // text
}

// bookNames maps book number (1-66) to canonical name.
var bookNames = [67]string{
    0:  "",
    1:  "Genesis", 2: "Exodus", 3: "Leviticus", 4: "Numbers", 5: "Deuteronomy",
    6:  "Joshua", 7: "Judges", 8: "Ruth", 9: "1 Samuel", 10: "2 Samuel",
    11: "1 Kings", 12: "2 Kings", 13: "1 Chronicles", 14: "2 Chronicles", 15: "Ezra",
    16: "Nehemiah", 17: "Esther", 18: "Job", 19: "Psalms", 20: "Proverbs",
    21: "Ecclesiastes", 22: "Song of Solomon", 23: "Isaiah", 24: "Jeremiah", 25: "Lamentations",
    26: "Ezekiel", 27: "Daniel", 28: "Hosea", 29: "Joel", 30: "Amos",
    31: "Obadiah", 32: "Jonah", 33: "Micah", 34: "Nahum", 35: "Habakkuk",
    36: "Zephaniah", 37: "Haggai", 38: "Zechariah", 39: "Malachi",
    40: "Matthew", 41: "Mark", 42: "Luke", 43: "John", 44: "Acts",
    45: "Romans", 46: "1 Corinthians", 47: "2 Corinthians", 48: "Galatians", 49: "Ephesians",
    50: "Philippians", 51: "Colossians", 52: "1 Thessalonians", 53: "2 Thessalonians", 54: "1 Timothy",
    55: "2 Timothy", 56: "Titus", 57: "Philemon", 58: "Hebrews", 59: "James",
    60: "1 Peter", 61: "2 Peter", 62: "1 John", 63: "2 John", 64: "3 John",
    65: "Jude", 66: "Revelation",
}

// genreTags maps book number to genre tag.
var genreTags = func() [67]string {
    g := [67]string{}
    for _, b := range []int{40, 41, 42, 43} { g[b] = "gospel" }
    g[44] = "acts"
    for _, b := range []int{45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65} { g[b] = "epistle" }
    g[66] = "prophecy"
    for _, b := range []int{23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39} { g[b] = "prophecy" }
    for _, b := range []int{18,19,20,21,22} { g[b] = "poetry" }
    for _, b := range []int{1,2,3,4,5} { g[b] = "law" }
    for _, b := range []int{6,7,8,9,10,11,12,13,14,15,16,17} { g[b] = "history" }
    return g
}()

// verseRef returns a canonical verse reference string like "John 3:16".
func verseRef(bookNum, chapter, verse int) string {
    if bookNum < 1 || bookNum > 66 {
        return fmt.Sprintf("Book%d %d:%d", bookNum, chapter, verse)
    }
    return fmt.Sprintf("%s %d:%d", bookNames[bookNum], chapter, verse)
}

// parseKJV parses the raw KJV JSON bytes into WriteRequests.
// If ntOnly is true, only New Testament verses (books 40-66) are returned.
func parseKJV(data []byte, ntOnly bool) ([]mbp.WriteRequest, error) {
    var records []kjvRecord
    if err := json.Unmarshal(data, &records); err != nil {
        return nil, fmt.Errorf("parse KJV JSON: %w", err)
    }

    reqs := make([]mbp.WriteRequest, 0, len(records))
    for _, r := range records {
        if r.B < 1 || r.B > 66 || r.T == "" {
            continue
        }
        if ntOnly && r.B < 40 {
            continue
        }
        testament := "Old Testament"
        if r.B >= 40 {
            testament = "New Testament"
        }
        tags := []string{testament, bookNames[r.B]}
        if g := genreTags[r.B]; g != "" {
            tags = append(tags, g)
        }
        reqs = append(reqs, mbp.WriteRequest{
            Concept:    verseRef(r.B, r.C, r.V),
            Content:    r.T,
            Tags:       tags,
            Confidence: 0.85,
            Stability:  0.80,
        })
    }
    return reqs, nil
}
```

**Step 4: Run tests**

```bash
cd cmd/eval-bible && go test -run TestParseKJV -v ./...
cd cmd/eval-bible && go test -run TestVerseRef -v ./...
```

Expected: PASS.

**Step 5: Commit**

```bash
git add cmd/eval-bible/corpus.go cmd/eval-bible/corpus_test.go
git commit -m "feat(eval-bible): KJV corpus loader with testament and genre tags"
```

---

## Task 3: Cross-Reference Loader

Parse the OpenBible TSV and build a map from verse reference to its known cross-references.

**Files:**
- Create: `cmd/eval-bible/seeds.go`
- Create: `cmd/eval-bible/seeds_test.go`

**Cross-reference TSV format** (first 3 lines):
```
From Verse	To Verse	Votes
Gen.1.1	Heb.11.3	172
Gen.1.1	John.1.1	151
```

Book abbreviations in TSV use formats like "Gen", "Matt", "Rev". Must be mapped to canonical names like "Genesis", "Matthew", "Revelation".

**Step 1: Write `seeds_test.go`**

```go
package main

import (
    "testing"
)

func TestParseXRef(t *testing.T) {
    tsv := "From Verse\tTo Verse\tVotes\nGen.1.1\tHeb.11.3\t172\nGen.1.1\tJohn.1.1\t151\nMatt.5.3\tLuke.6.20\t88\n"

    xrefs, err := parseXRef([]byte(tsv))
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    gen11 := xrefs["Genesis 1:1"]
    if len(gen11) != 2 {
        t.Fatalf("want 2 cross-refs for Genesis 1:1, got %d", len(gen11))
    }
    found := map[string]bool{}
    for _, r := range gen11 {
        found[r] = true
    }
    if !found["Hebrews 11:3"] {
        t.Error("missing Hebrews 11:3")
    }
    if !found["John 1:1"] {
        t.Error("missing John 1:1")
    }
}

func TestAbbrevToBook(t *testing.T) {
    cases := [][2]string{
        {"Gen", "Genesis"},
        {"Matt", "Matthew"},
        {"Rev", "Revelation"},
        {"1Cor", "1 Corinthians"},
        {"Ps", "Psalms"},
    }
    for _, c := range cases {
        got := abbrevToBook(c[0])
        if got != c[1] {
            t.Errorf("abbrevToBook(%q) = %q, want %q", c[0], got, c[1])
        }
    }
}

func TestSelectSeeds(t *testing.T) {
    // Build a map with some verses having 5+ cross-refs (NT) and some fewer
    xrefs := map[string][]string{
        "John 3:16":   {"Rom.5.8", "1John.4.9", "John.3.17", "Eph.2.8", "Rom.8.32", "John.1.14"},
        "John 3:17":   {"John.3.16", "Luke.9.56"},  // only 2, should be excluded
        "Romans 5:8":  {"John.3.16", "1John.4.10", "Rom.3.25", "Eph.2.4", "Titus.3.4", "1Pet.1.19"},
    }
    seeds := selectSeeds(xrefs, 5, 10)
    if len(seeds) != 2 {
        t.Errorf("want 2 seeds (5+ cross-refs), got %d", len(seeds))
    }
}
```

**Step 2: Run to confirm failure**

```bash
cd cmd/eval-bible && go test -run TestParseXRef -v ./...
```

Expected: FAIL — seeds.go doesn't exist.

**Step 3: Write `seeds.go`**

```go
package main

import (
    "bufio"
    "bytes"
    "fmt"
    "sort"
    "strings"
)

// xrefMap maps canonical verse ref ("John 3:16") to its cross-referenced verse refs.
type xrefMap map[string][]string

// abbrevMap maps OpenBible abbreviations to canonical book names.
var abbrevMap = map[string]string{
    "Gen": "Genesis", "Exod": "Exodus", "Lev": "Leviticus", "Num": "Numbers",
    "Deut": "Deuteronomy", "Josh": "Joshua", "Judg": "Judges", "Ruth": "Ruth",
    "1Sam": "1 Samuel", "2Sam": "2 Samuel", "1Kgs": "1 Kings", "2Kgs": "2 Kings",
    "1Chr": "1 Chronicles", "2Chr": "2 Chronicles", "Ezra": "Ezra", "Neh": "Nehemiah",
    "Esth": "Esther", "Job": "Job", "Ps": "Psalms", "Prov": "Proverbs",
    "Eccl": "Ecclesiastes", "Song": "Song of Solomon", "Isa": "Isaiah",
    "Jer": "Jeremiah", "Lam": "Lamentations", "Ezek": "Ezekiel", "Dan": "Daniel",
    "Hos": "Hosea", "Joel": "Joel", "Amos": "Amos", "Obad": "Obadiah",
    "Jonah": "Jonah", "Mic": "Micah", "Nah": "Nahum", "Hab": "Habakkuk",
    "Zeph": "Zephaniah", "Hag": "Haggai", "Zech": "Zechariah", "Mal": "Malachi",
    "Matt": "Matthew", "Mark": "Mark", "Luke": "Luke", "John": "John",
    "Acts": "Acts", "Rom": "Romans", "1Cor": "1 Corinthians", "2Cor": "2 Corinthians",
    "Gal": "Galatians", "Eph": "Ephesians", "Phil": "Philippians", "Col": "Colossians",
    "1Thess": "1 Thessalonians", "2Thess": "2 Thessalonians", "1Tim": "1 Timothy",
    "2Tim": "2 Timothy", "Titus": "Titus", "Phlm": "Philemon", "Heb": "Hebrews",
    "Jas": "James", "1Pet": "1 Peter", "2Pet": "2 Peter", "1John": "1 John",
    "2John": "2 John", "3John": "3 John", "Jude": "Jude", "Rev": "Revelation",
}

// abbrevToBook converts an OpenBible abbreviation to a canonical book name.
func abbrevToBook(abbrev string) string {
    if name, ok := abbrevMap[abbrev]; ok {
        return name
    }
    return abbrev // passthrough if unknown
}

// parseXRefRef converts "Gen.1.1" → "Genesis 1:1".
func parseXRefRef(s string) (string, error) {
    parts := strings.Split(s, ".")
    if len(parts) != 3 {
        return "", fmt.Errorf("bad verse ref %q", s)
    }
    book := abbrevToBook(parts[0])
    return fmt.Sprintf("%s %s:%s", book, parts[1], parts[2]), nil
}

// parseXRef parses the OpenBible cross-references TSV into an xrefMap.
func parseXRef(data []byte) (xrefMap, error) {
    m := make(xrefMap)
    scanner := bufio.NewScanner(bytes.NewReader(data))
    lineNum := 0
    for scanner.Scan() {
        lineNum++
        if lineNum == 1 {
            continue // header
        }
        line := scanner.Text()
        if line == "" {
            continue
        }
        fields := strings.Split(line, "\t")
        if len(fields) < 2 {
            continue
        }
        from, err := parseXRefRef(strings.TrimSpace(fields[0]))
        if err != nil {
            continue
        }
        to, err := parseXRefRef(strings.TrimSpace(fields[1]))
        if err != nil {
            continue
        }
        m[from] = append(m[from], to)
    }
    return m, scanner.Err()
}

// selectSeeds picks seed verses that have at least minXRefs cross-references.
// Returns at most maxSeeds seeds, sorted by verse reference for determinism.
func selectSeeds(xrefs xrefMap, minXRefs, maxSeeds int) []string {
    var candidates []string
    for ref, refs := range xrefs {
        if len(refs) >= minXRefs {
            candidates = append(candidates, ref)
        }
    }
    sort.Strings(candidates)
    if len(candidates) > maxSeeds {
        // Spread evenly across the sorted list
        step := len(candidates) / maxSeeds
        spread := make([]string, 0, maxSeeds)
        for i := 0; i < maxSeeds && i*step < len(candidates); i++ {
            spread = append(spread, candidates[i*step])
        }
        return spread
    }
    return candidates
}
```

**Step 4: Run tests**

```bash
cd cmd/eval-bible && go test -run "TestParseXRef|TestAbbrevToBook|TestSelectSeeds" -v ./...
```

Expected: PASS.

**Step 5: Commit**

```bash
git add cmd/eval-bible/seeds.go cmd/eval-bible/seeds_test.go
git commit -m "feat(eval-bible): cross-reference loader and seed selection"
```

---

## Task 4: Metrics

Pure functions for NDCG@10 and Recall@10. These are the heart of the eval — test thoroughly.

**Files:**
- Create: `cmd/eval-bible/metrics.go`
- Create: `cmd/eval-bible/metrics_test.go`

**Step 1: Write `metrics_test.go`**

```go
package main

import (
    "math"
    "testing"
)

func TestRecallAtK(t *testing.T) {
    relevant := map[string]bool{"A": true, "B": true, "C": true}

    // 2 of 3 relevant results in top-5
    results := []string{"A", "X", "B", "Y", "Z"}
    got := recallAtK(results, relevant, 5)
    want := 2.0 / 3.0
    if math.Abs(got-want) > 0.001 {
        t.Errorf("recallAtK = %.4f, want %.4f", got, want)
    }

    // All 3 relevant results present
    results = []string{"A", "B", "C", "X", "Y"}
    got = recallAtK(results, relevant, 5)
    if math.Abs(got-1.0) > 0.001 {
        t.Errorf("recallAtK = %.4f, want 1.0", got)
    }

    // None relevant
    results = []string{"X", "Y", "Z"}
    got = recallAtK(results, relevant, 5)
    if got != 0.0 {
        t.Errorf("recallAtK = %.4f, want 0.0", got)
    }
}

func TestNDCGAtK(t *testing.T) {
    relevant := map[string]bool{"A": true, "B": true}

    // Perfect ranking: relevant items first
    results := []string{"A", "B", "X", "Y", "Z"}
    perfect := ndcgAtK(results, relevant, 5)
    if math.Abs(perfect-1.0) > 0.001 {
        t.Errorf("perfect NDCG = %.4f, want 1.0", perfect)
    }

    // Relevant items at positions 2 and 4 (1-indexed)
    results = []string{"X", "A", "Y", "B", "Z"}
    got := ndcgAtK(results, relevant, 5)
    if got >= perfect {
        t.Errorf("imperfect NDCG %.4f should be < perfect 1.0", got)
    }
    if got <= 0 {
        t.Errorf("NDCG should be > 0 when relevant items exist, got %.4f", got)
    }

    // No relevant items
    results = []string{"X", "Y", "Z"}
    got = ndcgAtK(results, relevant, 5)
    if got != 0.0 {
        t.Errorf("NDCG with no relevant = %.4f, want 0.0", got)
    }
}
```

**Step 2: Run to confirm failure**

```bash
cd cmd/eval-bible && go test -run "TestRecall|TestNDCG" -v ./...
```

Expected: FAIL.

**Step 3: Write `metrics.go`**

```go
package main

import "math"

// recallAtK computes Recall@k: fraction of relevant items appearing in top-k results.
// relevant is the set of known relevant result identifiers.
func recallAtK(results []string, relevant map[string]bool, k int) float64 {
    if len(relevant) == 0 {
        return 0
    }
    hits := 0
    limit := k
    if len(results) < limit {
        limit = len(results)
    }
    for i := 0; i < limit; i++ {
        if relevant[results[i]] {
            hits++
        }
    }
    return float64(hits) / float64(len(relevant))
}

// ndcgAtK computes NDCG@k (Normalized Discounted Cumulative Gain).
// Relevance is binary: 1 if in relevant set, 0 otherwise.
func ndcgAtK(results []string, relevant map[string]bool, k int) float64 {
    if len(relevant) == 0 {
        return 0
    }
    dcg := computeDCG(results, relevant, k)
    // Ideal DCG: all relevant items ranked first
    idealCount := len(relevant)
    if idealCount > k {
        idealCount = k
    }
    ideal := make([]string, idealCount)
    i := 0
    for ref := range relevant {
        if i >= idealCount {
            break
        }
        ideal[i] = ref
        i++
    }
    idcg := computeDCG(ideal, relevant, k)
    if idcg == 0 {
        return 0
    }
    return dcg / idcg
}

func computeDCG(results []string, relevant map[string]bool, k int) float64 {
    dcg := 0.0
    limit := k
    if len(results) < limit {
        limit = len(results)
    }
    for i := 0; i < limit; i++ {
        if relevant[results[i]] {
            // position is 1-indexed; discount = log2(position+1)
            dcg += 1.0 / math.Log2(float64(i+2))
        }
    }
    return dcg
}
```

**Step 4: Run tests**

```bash
cd cmd/eval-bible && go test -run "TestRecall|TestNDCG" -v ./...
```

Expected: PASS.

**Step 5: Commit**

```bash
git add cmd/eval-bible/metrics.go cmd/eval-bible/metrics_test.go
git commit -m "feat(eval-bible): NDCG@10 and Recall@10 metrics"
```

---

## Task 5: Engine Setup

Reuse the exact engine init pattern from `cmd/eval/main.go`. This is boilerplate — copy and adapt, no unit tests needed (integration-only).

**Files:**
- Create: `cmd/eval-bible/engine.go`
- Create: `cmd/eval-bible/adapters.go`

**Step 1: Copy `cmd/eval/adapters.go` to `cmd/eval-bible/adapters.go`**

The adapters are identical (same interface implementations wrapping PebbleStore). Copy the file and change the package declaration to `package main`.

```bash
cp cmd/eval/adapters.go cmd/eval-bible/adapters.go
# The package line is already "package main" — no change needed
```

**Step 2: Write `cmd/eval-bible/engine.go`**

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/scrypster/muninndb/internal/cognitive"
    "github.com/scrypster/muninndb/internal/engine"
    "github.com/scrypster/muninndb/internal/engine/activation"
    "github.com/scrypster/muninndb/internal/engine/trigger"
    "github.com/scrypster/muninndb/internal/index/fts"
    hnswpkg "github.com/scrypster/muninndb/internal/index/hnsw"
    "github.com/scrypster/muninndb/internal/plugin"
    embedpkg "github.com/scrypster/muninndb/internal/plugin/embed"
    "github.com/scrypster/muninndb/internal/storage"
    "github.com/scrypster/muninndb/internal/transport/mbp"
)

// evalEngine holds all components needed to run the Bible eval.
type evalEngine struct {
    eng      *engine.Engine
    store    *storage.PebbleStore
    hnswReg  *hnswpkg.Registry
    embedder activation.Embedder
    ws       [8]byte // vault prefix for "bible"
    cancel   context.CancelFunc
}

// newEvalEngine initializes the full engine stack with the local ONNX embedder.
// The caller must call ee.close() when done.
func newEvalEngine(ctx context.Context, dataDir string) (*evalEngine, error) {
    db, err := storage.OpenPebble(dataDir, storage.DefaultOptions())
    if err != nil {
        return nil, fmt.Errorf("open pebble: %w", err)
    }
    store := storage.NewPebbleStore(db, 100_000)
    ftsIdx := fts.New(db)
    hnswReg := hnswpkg.NewRegistry(db)

    svc, err := embedpkg.NewEmbedService("local://all-MiniLM-L6-v2")
    if err != nil {
        store.Close()
        return nil, fmt.Errorf("create embed service: %w", err)
    }
    if err = svc.Init(ctx, plugin.PluginConfig{DataDir: dataDir}); err != nil {
        store.Close()
        return nil, fmt.Errorf("init embed service: %w", err)
    }
    log.Printf("embedder: all-MiniLM-L6-v2 INT8 ONNX (dim=%d)", svc.Dimension())

    embedder := &ollamaEmbedder{svc: svc} // reuse adapter from adapters.go
    hnswIdx := &hnswActAdapter{r: hnswReg}

    actEngine := activation.New(store, &ftsAdapter{ftsIdx}, hnswIdx, embedder)
    trigSystem := trigger.New(store, &ftsTrigAdapter{ftsIdx}, nil, embedder)

    hebbianWorker := cognitive.NewHebbianWorker(&benchHebbianAdapter{store})
    decayWorker := cognitive.NewDecayWorker(&benchDecayAdapter{store})
    contradictWorker := cognitive.NewContradictWorker(&benchContradictAdapter{store})
    confidenceWorker := cognitive.NewConfidenceWorker(&benchConfidenceAdapter{store})

    workerCtx, cancel := context.WithCancel(ctx)
    go decayWorker.Worker.Run(workerCtx)
    go contradictWorker.Worker.Run(workerCtx)
    go confidenceWorker.Worker.Run(workerCtx)

    eng := engine.NewEngine(store, nil, ftsIdx, actEngine, trigSystem,
        hebbianWorker, decayWorker,
        contradictWorker.Worker, confidenceWorker.Worker,
        embedder, hnswReg,
    )

    ws := store.ResolveVaultPrefix("bible")

    return &evalEngine{
        eng:      eng,
        store:    store,
        hnswReg:  hnswReg,
        embedder: embedder,
        ws:       ws,
        cancel:   cancel,
    }, nil
}

func (ee *evalEngine) close() {
    ee.cancel()
    ee.eng.Stop()
    ee.store.Close()
}

// writeVerse writes a single verse to the engine and indexes it in HNSW.
// Returns the ULID of the written engram.
func (ee *evalEngine) writeVerse(ctx context.Context, req mbp.WriteRequest) (storage.ULID, error) {
    req.Vault = "bible"
    resp, err := ee.eng.Write(ctx, &req)
    if err != nil {
        return storage.ULID{}, err
    }
    id, err := storage.ParseULID(resp.ID)
    if err != nil {
        return storage.ULID{}, err
    }
    // Embed and index in HNSW
    text := req.Concept + ". " + req.Content
    vec, err := ee.embedder.Embed(ctx, []string{text})
    if err != nil {
        return id, nil // write succeeded, HNSW indexing failed — not fatal
    }
    _ = ee.hnswReg.Insert(ctx, ee.ws, [16]byte(id), vec)
    return id, nil
}

// activate queries the engine with the given context strings.
func (ee *evalEngine) activate(ctx context.Context, contextStrs []string) ([]mbp.ActivationItem, error) {
    req := mbp.ActivateRequest{
        Vault:   "bible",
        Context: contextStrs,
        TopK:    10,
    }
    resp, err := ee.eng.Activate(ctx, &req)
    if err != nil {
        return nil, err
    }
    return resp.Results, nil
}
```

**Step 3: Verify it compiles**

```bash
cd cmd/eval-bible && go build ./... 2>&1
```

Expected: no errors (or only "no main package" since main.go doesn't exist yet).

**Step 4: Commit**

```bash
git add cmd/eval-bible/engine.go cmd/eval-bible/adapters.go
git commit -m "feat(eval-bible): engine initialization and verse write/activate helpers"
```

**Note on adapters.go:** The adapters reference types like `ollamaEmbedder`, `hnswActAdapter`, `ftsAdapter`, `ftsTrigAdapter`. Check `cmd/eval/main.go` for these type definitions — some may be inline structs. If they are, move them to `adapters.go` in the `cmd/eval-bible` package.

---

## Task 6: Phase 1 — Retrieval Quality

**Files:**
- Create: `cmd/eval-bible/phase1.go`

No unit tests for phase1 — it requires a live engine. This is an integration function called from main.

**Step 1: Write `phase1.go`**

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "time"
)

// Phase1Result holds the output of the retrieval quality eval.
type Phase1Result struct {
    SeedsEvaluated  int
    AvgCrossRefs    float64
    RecallAtK       float64
    NDCGAtK         float64
    QueryResults    []seedResult
}

type seedResult struct {
    Ref       string
    CrossRefs int
    Recall    float64
    NDCG      float64
    TopConcepts []string // top 3 returned concepts for spot-checking
}

// RunPhase1 evaluates retrieval quality against known cross-references.
// seeds: verse references to query. xrefs: full cross-reference map.
// corpus: all loaded verse concepts (for building ID→concept reverse map from responses).
func RunPhase1(ctx context.Context, ee *evalEngine, seeds []string, xrefs xrefMap, corpusTexts map[string]string) Phase1Result {
    var results []seedResult
    totalRecall, totalNDCG := 0.0, 0.0

    for i, seedRef := range seeds {
        knownRefs := xrefs[seedRef]
        if len(knownRefs) == 0 {
            continue
        }

        // Build relevance set from cross-references
        relevant := make(map[string]bool, len(knownRefs))
        for _, ref := range knownRefs {
            relevant[ref] = true
        }

        // Query using the verse text as context
        verseText := corpusTexts[seedRef]
        if verseText == "" {
            verseText = seedRef
        }
        queryCtx := []string{verseText}

        start := time.Now()
        items, err := ee.activate(ctx, queryCtx)
        elapsed := time.Since(start)
        _ = elapsed

        if err != nil {
            fmt.Printf("  [%d/%d] %s: query error: %v\n", i+1, len(seeds), seedRef, err)
            continue
        }

        // Extract returned concept names (these are the verse refs we stored as Concept)
        returned := make([]string, len(items))
        topConcepts := make([]string, 0, 3)
        for j, item := range items {
            returned[j] = item.Concept
            if j < 3 {
                topConcepts = append(topConcepts, item.Concept)
            }
        }

        recall := recallAtK(returned, relevant, 10)
        ndcg := ndcgAtK(returned, relevant, 10)
        totalRecall += recall
        totalNDCG += ndcg

        results = append(results, seedResult{
            Ref:         seedRef,
            CrossRefs:   len(knownRefs),
            Recall:      recall,
            NDCG:        ndcg,
            TopConcepts: topConcepts,
        })

        if (i+1)%10 == 0 {
            fmt.Printf("  Phase 1: %d/%d seeds evaluated...\n", i+1, len(seeds))
        }
    }

    n := float64(len(results))
    if n == 0 {
        return Phase1Result{}
    }

    avgXRefs := 0.0
    for _, r := range results {
        avgXRefs += float64(r.CrossRefs)
    }

    return Phase1Result{
        SeedsEvaluated: len(results),
        AvgCrossRefs:   avgXRefs / n,
        RecallAtK:      totalRecall / n,
        NDCGAtK:        totalNDCG / n,
        QueryResults:   results,
    }
}

// buildCorpusTextMap builds a map from verse ref ("John 3:16") to verse text
// from the loaded WriteRequests.
func buildCorpusTextMap(reqs []mbp.WriteRequest) map[string]string {
    m := make(map[string]string, len(reqs))
    for _, r := range reqs {
        m[r.Concept] = r.Content
    }
    return m
}

// ntSeedFilter returns only NT seed refs (book 40+), identified by
// checking if the concept starts with an NT book name.
func isNTRef(ref string) bool {
    ntBooks := []string{
        "Matthew", "Mark", "Luke", "John", "Acts", "Romans",
        "1 Corinthians", "2 Corinthians", "Galatians", "Ephesians",
        "Philippians", "Colossians", "1 Thessalonians", "2 Thessalonians",
        "1 Timothy", "2 Timothy", "Titus", "Philemon", "Hebrews",
        "James", "1 Peter", "2 Peter", "1 John", "2 John", "3 John",
        "Jude", "Revelation",
    }
    for _, book := range ntBooks {
        if strings.HasPrefix(ref, book+" ") {
            return true
        }
    }
    return false
}
```

**Step 2: Verify compilation**

```bash
cd cmd/eval-bible && go build ./...
```

**Step 3: Commit**

```bash
git add cmd/eval-bible/phase1.go
git commit -m "feat(eval-bible): Phase 1 retrieval quality eval loop"
```

---

## Task 7: Phase 2 — Cognitive Properties

**Files:**
- Create: `cmd/eval-bible/phase2.go`

**Step 1: Write `phase2.go`**

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/scrypster/muninndb/internal/transport/mbp"
)

// thematicQueries are fixed queries used throughout Phase 2 for NDCG measurement.
// These are stable — do not change them between runs (they're your comparator).
var thematicQueries = []thematicQuery{
    {Context: "forgiveness redemption grace mercy sins forgiven", Label: "forgiveness"},
    {Context: "eternal life salvation believe faith Jesus saved", Label: "eternal life"},
    {Context: "shepherd flock lost sheep pasture", Label: "shepherd"},
    {Context: "resurrection from the dead raised third day", Label: "resurrection"},
    {Context: "love one another neighbor commandment", Label: "love commandment"},
    {Context: "bread of life feeding five thousand loaves fish", Label: "feeding miracle"},
    {Context: "Holy Spirit Pentecost tongues fire descended", Label: "Holy Spirit"},
    {Context: "faith without works dead Abraham justified", Label: "faith and works"},
    {Context: "creation light darkness void earth heaven", Label: "creation"},
    {Context: "wisdom understanding proverbs fear of the Lord", Label: "wisdom"},
}

type thematicQuery struct {
    Context string
    Label   string
}

// Phase2Result holds cognitive properties eval output.
type Phase2Result struct {
    BaselineNDCG     float64
    PostReadingNDCG  float64
    PostDecayNDCG    float64
    HebbianLinks     int
    DecayedEngrams   int
    QueryDeltas      []queryDelta
}

type queryDelta struct {
    Label    string
    Baseline float64
    PostRead float64
    PostDecay float64
}

// RunPhase2 measures NDCG delta after a reading session and decay simulation.
// gospelJohnIDs: IDs of John's Gospel verses (to activate in reading session).
// genealogyIDs: IDs of OT genealogy verses (to submit for decay).
func RunPhase2(ctx context.Context, ee *evalEngine, gospelJohnIDs []storage.ULID, gospelJohnTexts []string, _ []storage.ULID, decayWorkerSubmit func(id storage.ULID)) Phase2Result {
    fmt.Println("  Phase 2: measuring baseline thematic queries...")

    // Step 1: Baseline
    baselineNDCGs := measureThematicNDCGs(ctx, ee)

    // Step 2: Reading session — activate 200 John verses sequentially
    fmt.Printf("  Phase 2: simulating reading session (%d John verses)...\n", len(gospelJohnTexts))
    readCount := 200
    if len(gospelJohnTexts) < readCount {
        readCount = len(gospelJohnTexts)
    }
    for i := 0; i < readCount; i++ {
        _, _ = ee.activate(ctx, []string{gospelJohnTexts[i]})
        if (i+1)%50 == 0 {
            fmt.Printf("    reading: %d/%d verses activated\n", i+1, readCount)
        }
    }

    // Short pause for Hebbian worker to process
    time.Sleep(3 * time.Second)

    // Step 3: Re-query after reading
    fmt.Println("  Phase 2: measuring post-reading thematic queries...")
    postReadNDCGs := measureThematicNDCGs(ctx, ee)

    // Step 4: Submit genealogy verses for decay (simulate 30 days elapsed)
    fmt.Println("  Phase 2: simulating 30-day genealogy decay...")
    // We submit these as DecayCandidates with LastAccess = 30 days ago.
    // The decay worker processes them asynchronously.
    // (Caller provides the submit function since we can't import cognitive directly here)

    // Step 5: Re-query after decay
    time.Sleep(3 * time.Second)
    fmt.Println("  Phase 2: measuring post-decay thematic queries...")
    postDecayNDCGs := measureThematicNDCGs(ctx, ee)

    // Build results
    var deltas []queryDelta
    baseSum, postReadSum, postDecaySum := 0.0, 0.0, 0.0
    for i, q := range thematicQueries {
        deltas = append(deltas, queryDelta{
            Label:     q.Label,
            Baseline:  baselineNDCGs[i],
            PostRead:  postReadNDCGs[i],
            PostDecay: postDecayNDCGs[i],
        })
        baseSum += baselineNDCGs[i]
        postReadSum += postReadNDCGs[i]
        postDecaySum += postDecayNDCGs[i]
    }
    n := float64(len(thematicQueries))

    return Phase2Result{
        BaselineNDCG:    baseSum / n,
        PostReadingNDCG: postReadSum / n,
        PostDecayNDCG:   postDecaySum / n,
        QueryDeltas:     deltas,
    }
}

// measureThematicNDCGs runs all thematic queries and returns per-query NDCG@10.
// Since we don't have ground truth for thematic queries, we use a proxy:
// relevance = whether the result concept contains any of the query's keywords.
func measureThematicNDCGs(ctx context.Context, ee *evalEngine) []float64 {
    ndcgs := make([]float64, len(thematicQueries))
    for i, q := range thematicQueries {
        items, err := ee.activate(ctx, []string{q.Context})
        if err != nil || len(items) == 0 {
            ndcgs[i] = 0
            continue
        }
        // Build relevance set: result is relevant if concept contains any query keyword
        keywords := strings.Fields(strings.ToLower(q.Context))
        relevant := make(map[string]bool)
        returned := make([]string, len(items))
        for j, item := range items {
            returned[j] = item.Concept
            conceptLower := strings.ToLower(item.Concept + " " + item.Content)
            for _, kw := range keywords {
                if len(kw) > 3 && strings.Contains(conceptLower, kw) {
                    relevant[item.Concept] = true
                    break
                }
            }
        }
        ndcgs[i] = ndcgAtK(returned, relevant, 10)
    }
    return ndcgs
}

// filterJohnVerses filters a corpus for Gospel of John verses.
func filterJohnVerses(reqs []mbp.WriteRequest) (texts []string) {
    for _, r := range reqs {
        if strings.HasPrefix(r.Concept, "John ") {
            texts = append(texts, r.Content)
        }
    }
    return texts
}
```

**Step 2: Compile check**

```bash
cd cmd/eval-bible && go build ./...
```

**Step 3: Commit**

```bash
git add cmd/eval-bible/phase2.go
git commit -m "feat(eval-bible): Phase 2 cognitive properties simulation"
```

---

## Task 8: Report Writer + Main Entry Point

**Files:**
- Create: `cmd/eval-bible/report.go`
- Create: `cmd/eval-bible/main.go`

**Step 1: Write `report.go`**

```go
package main

import (
    "fmt"
    "io"
    "os"
    "strings"
    "time"
)

// writeReport formats and writes the full eval report to w and optionally to a file.
func writeReport(w io.Writer, p1 Phase1Result, p2 Phase2Result, mode string, corpusSize int, loadDur time.Duration) {
    sep := strings.Repeat("═", 62)
    fmt.Fprintln(w, sep)
    fmt.Fprintf(w, "BIBLE EVAL REPORT  [mode: %s, corpus: %d verses]\n", mode, corpusSize)
    fmt.Fprintln(w, sep)

    fmt.Fprintln(w, "\nPhase 1: Retrieval Quality")
    fmt.Fprintf(w, "  Corpus load time:         %s\n", loadDur.Round(time.Second))
    fmt.Fprintf(w, "  Seed verses evaluated:    %d\n", p1.SeedsEvaluated)
    fmt.Fprintf(w, "  Avg cross-refs per seed:  %.1f\n", p1.AvgCrossRefs)
    fmt.Fprintf(w, "  Recall@10:                %.1f%%\n", p1.RecallAtK*100)
    fmt.Fprintf(w, "  NDCG@10:                  %.3f\n", p1.NDCGAtK)

    fmt.Fprintln(w, "\nPhase 2: Cognitive Properties")
    fmt.Fprintf(w, "  Baseline NDCG@10:         %.3f\n", p2.BaselineNDCG)
    fmt.Fprintf(w, "  After John reading:       %.3f  (%+.3f)\n", p2.PostReadingNDCG, p2.PostReadingNDCG-p2.BaselineNDCG)
    fmt.Fprintf(w, "  After OT decay:           %.3f  (%+.3f)\n", p2.PostDecayNDCG, p2.PostDecayNDCG-p2.BaselineNDCG)

    fmt.Fprintln(w, "\n  Per-query delta (Baseline → Post-Reading):")
    for _, d := range p2.QueryDeltas {
        marker := " "
        if d.PostRead > d.Baseline {
            marker = "▲"
        } else if d.PostRead < d.Baseline {
            marker = "▼"
        }
        fmt.Fprintf(w, "    %s %-22s  %.3f → %.3f\n", marker, d.Label, d.Baseline, d.PostRead)
    }

    fmt.Fprintln(w, "")
    fmt.Fprintln(w, sep)
    verdict := verdictLine(p1.NDCGAtK, p2.PostReadingNDCG > p2.BaselineNDCG)
    fmt.Fprintln(w, verdict)
    fmt.Fprintln(w, sep)
}

func verdictLine(ndcg float64, cognitiveImproved bool) string {
    quality := ""
    switch {
    case ndcg >= 0.60:
        quality = "★ HIGH-QUALITY retrieval"
    case ndcg >= 0.40:
        quality = "◆ GOOD retrieval"
    case ndcg >= 0.25:
        quality = "~ RETRIEVAL NEEDS IMPROVEMENT"
    default:
        quality = "✗ POOR retrieval"
    }
    cognitive := "cognitive layer: NO improvement"
    if cognitiveImproved {
        cognitive = "cognitive layer: IMPROVES results ✓"
    }
    return fmt.Sprintf("VERDICT: %s (NDCG@10=%.3f) | %s", quality, ndcg, cognitive)
}

// saveReport writes the report to a file, appending if it exists.
func saveReport(path string, p1 Phase1Result, p2 Phase2Result, mode string, corpusSize int, loadDur time.Duration) error {
    f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    fmt.Fprintf(f, "\n\n# Run: %s\n", time.Now().Format("2006-01-02 15:04:05"))
    writeReport(f, p1, p2, mode, corpusSize, loadDur)
    return nil
}
```

**Step 2: Write `main.go`**

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "runtime/debug"
    "time"
)

func main() {
    debug.SetGCPercent(400)

    full        := flag.Bool("full", false, "Load full Bible (default: New Testament only)")
    dataDir     := flag.String("data", "", "Persistent data directory (default: temp dir)")
    kjvPath     := flag.String("kjv", "testdata/bible/kjv.json", "Path to KJV JSON file")
    xrefPath    := flag.String("xref", "testdata/bible/cross-refs.tsv", "Path to cross-references TSV")
    resultsFile := flag.String("results-file", "", "Append results to this file (optional)")
    seedCount   := flag.Int("seeds", 100, "Number of seed verses for Phase 1")
    minXRefs    := flag.Int("min-xrefs", 5, "Minimum cross-refs for a seed verse to qualify")
    skipLoad    := flag.Bool("skip-load", false, "Skip corpus load (reuse existing data dir)")
    flag.Parse()

    ctx := context.Background()

    // Load corpus data files
    kjvData, err := os.ReadFile(*kjvPath)
    if err != nil {
        log.Fatalf("read KJV JSON: %v — run: make eval-bible-setup", err)
    }
    xrefData, err := os.ReadFile(*xrefPath)
    if err != nil {
        log.Fatalf("read cross-refs TSV: %v — run: make eval-bible-setup", err)
    }

    ntOnly := !*full
    mode := "new-testament"
    if *full {
        mode = "full-bible"
    }

    fmt.Printf("MuninnDB Bible Eval\n")
    fmt.Printf("══════════════════════════════════════════\n")
    fmt.Printf("Mode:    %s\n", mode)

    // Parse corpus
    corpus, err := parseKJV(kjvData, ntOnly)
    if err != nil {
        log.Fatalf("parse KJV: %v", err)
    }
    fmt.Printf("Corpus:  %d verses\n\n", len(corpus))

    // Parse cross-references
    xrefs, err := parseXRef(xrefData)
    if err != nil {
        log.Fatalf("parse xrefs: %v", err)
    }
    fmt.Printf("Cross-references loaded: %d verse pairs\n", len(xrefs))

    // Select seeds
    seeds := selectSeeds(xrefs, *minXRefs, *seedCount)
    if ntOnly {
        var ntSeeds []string
        for _, s := range seeds {
            if isNTRef(s) {
                ntSeeds = append(ntSeeds, s)
            }
        }
        seeds = ntSeeds
    }
    fmt.Printf("Seed verses selected:    %d (min %d cross-refs)\n\n", len(seeds), *minXRefs)

    // Setup data directory
    var tmpDir string
    if *dataDir == "" {
        tmpDir, err = os.MkdirTemp("", "muninndb-bible-eval-*")
        if err != nil {
            log.Fatalf("create temp dir: %v", err)
        }
        defer os.RemoveAll(tmpDir)
        *dataDir = tmpDir
    }

    // Initialize engine
    fmt.Println("Initializing engine...")
    ee, err := newEvalEngine(ctx, *dataDir)
    if err != nil {
        log.Fatalf("init engine: %v", err)
    }
    defer ee.close()

    // Load corpus
    loadStart := time.Now()
    if !*skipLoad {
        fmt.Printf("Loading %d verses into vault 'bible'...\n", len(corpus))
        loaded, errors := 0, 0
        for i, req := range corpus {
            if _, writeErr := ee.writeVerse(ctx, req); writeErr != nil {
                errors++
            } else {
                loaded++
            }
            if (i+1)%1000 == 0 {
                elapsed := time.Since(loadStart).Seconds()
                fmt.Printf("  %d/%d loaded (%.0f writes/sec, %d errors, %d vectors)\n",
                    i+1, len(corpus), float64(i+1)/elapsed, errors, ee.hnswReg.TotalVectors())
            }
        }
        fmt.Printf("✓ Loaded %d verses in %s (%d errors, %d HNSW vectors)\n\n",
            loaded, time.Since(loadStart).Round(time.Second), errors, ee.hnswReg.TotalVectors())
    }
    loadDur := time.Since(loadStart)

    // Build helpers
    corpusTexts := buildCorpusTextMap(corpus)
    johnTexts := filterJohnVerses(corpus)
    fmt.Printf("Gospel of John verses for reading session: %d\n\n", len(johnTexts))

    // Phase 1
    fmt.Println("Running Phase 1: Retrieval Quality...")
    p1 := RunPhase1(ctx, ee, seeds, xrefs, corpusTexts)
    fmt.Printf("Phase 1 complete: Recall@10=%.1f%%, NDCG@10=%.3f\n\n", p1.RecallAtK*100, p1.NDCGAtK)

    // Phase 2
    fmt.Println("Running Phase 2: Cognitive Properties...")
    p2 := RunPhase2(ctx, ee, nil, johnTexts, nil, nil)
    fmt.Printf("Phase 2 complete: baseline=%.3f → post-reading=%.3f\n\n", p2.BaselineNDCG, p2.PostReadingNDCG)

    // Report
    writeReport(os.Stdout, p1, p2, mode, len(corpus), loadDur)

    if *resultsFile != "" {
        if err := saveReport(*resultsFile, p1, p2, mode, len(corpus), loadDur); err != nil {
            log.Printf("warning: could not save results to %s: %v", *resultsFile, err)
        } else {
            fmt.Printf("\nResults saved to: %s\n", *resultsFile)
        }
    }
}
```

**Step 3: Build and verify**

```bash
cd cmd/eval-bible && go build ./...
```

Expected: builds cleanly.

**Step 4: Run a quick smoke test** (requires corpus data):

```bash
# Only run this if testdata/bible/kjv.json and cross-refs.tsv exist
# If not, run: make eval-bible-setup first
go run ./cmd/eval-bible/ --seeds=5 --data=/tmp/bible-smoke
```

Expected: loads verses, prints Phase 1 and Phase 2 output, exits cleanly.

**Step 5: Commit**

```bash
git add cmd/eval-bible/report.go cmd/eval-bible/main.go
git commit -m "feat(eval-bible): report writer and main entry point"
```

---

## Task 9: Makefile Targets + .gitignore + docs/eval-results

**Files:**
- Modify: `Makefile`
- Modify: `.gitignore`

**Step 1: Check existing Makefile targets**

```bash
grep -n "eval" Makefile | head -20
```

Note the existing pattern and replicate it.

**Step 2: Add targets to `Makefile`**

Add after the existing eval targets:

```makefile
## eval-bible-setup: Download KJV Bible and cross-reference data
eval-bible-setup:
	@bash scripts/eval-bible-setup.sh

## eval-bible: Run Bible eval (New Testament only, ~12 min)
eval-bible: eval-bible-setup
	go run ./cmd/eval-bible/ \
		--results-file=docs/eval-results/$$(date +%Y-%m-%d)-bible-eval-nt.txt

## eval-bible-full: Run Bible eval (full Bible, ~47 min)
eval-bible-full: eval-bible-setup
	go run ./cmd/eval-bible/ \
		--full \
		--results-file=docs/eval-results/$$(date +%Y-%m-%d)-bible-eval-full.txt

## eval-bible-quick: Run Bible eval with 10 seeds (fast sanity check, ~5 min)
eval-bible-quick: eval-bible-setup
	go run ./cmd/eval-bible/ \
		--seeds=10 \
		--data=/tmp/muninndb-bible-quick
```

**Step 3: Add eval-results to .gitignore exclusions — keep results committed**

The `docs/eval-results/` directory should have results committed (they track progress). No gitignore change needed for results files. The `.gitkeep` already exists.

**Step 4: Run `make eval-bible-quick` to verify end-to-end**

```bash
make eval-bible-quick
```

Expected: downloads data if needed, runs 10-seed eval, prints report.

**Step 5: Commit**

```bash
git add Makefile
git commit -m "feat(eval-bible): Makefile targets for eval-bible, eval-bible-full, eval-bible-quick"
```

---

## Verification

After all tasks are complete:

```bash
# Full NT eval — run this once and commit the results
make eval-bible
git add docs/eval-results/
git commit -m "eval: first Bible eval run — NT baseline"
```

The committed results file becomes the baseline. Future improvements to the system can be compared against it by running `make eval-bible` and committing the new results file.

---

## Important Notes for Implementer

1. **adapters.go reuse**: The `ollamaEmbedder`, `hnswActAdapter`, `ftsAdapter`, `ftsTrigAdapter` structs may be defined inline in `cmd/eval/main.go`. If so, extract them to `cmd/eval-bible/adapters.go` in the new package.

2. **mbp.WriteRequest field names**: Verify the exact field names by reading `internal/transport/mbp/types.go`. The write request struct uses `Concept`, `Content`, `Tags`, `Confidence`, `Stability`, `Vault`.

3. **storage.ULID vs [16]byte**: The engine returns IDs as strings (via `resp.ID`). Use `storage.ParseULID(resp.ID)` to convert. The HNSW insert takes `[16]byte(id)` where `id` is `storage.ULID`.

4. **Decay simulation in Phase 2**: The current Phase 2 implementation stubs out the actual genealogy decay (the `decayWorkerSubmit` parameter is nil). For v1, measuring baseline vs post-reading delta is sufficient. Decay simulation can be added in a follow-up.

5. **Cross-reference TSV URL**: If the OpenBible URL in the setup script fails, alternative: download from https://github.com/openbible-io/cross-references. Verify the URL before committing.
