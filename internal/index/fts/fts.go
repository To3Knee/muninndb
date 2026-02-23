package fts

import (
	"context"
	"encoding/binary"
	"math"
	"strings"
	"sync"
	"unicode"

	"github.com/cockroachdb/pebble"
	"github.com/scrypster/muninndb/internal/storage/keys"
)

const (
	k1 = 1.2
	b  = 0.75

	FieldConcept    = uint8(0x01)
	FieldTags       = uint8(0x02)
	FieldContent    = uint8(0x03)
	FieldCreatedBy  = uint8(0x04)

	fieldWeightConcept   = 3.0
	fieldWeightTags      = 2.0
	fieldWeightContent   = 1.0
	fieldWeightCreatedBy = 0.5

	ContentCompressThreshold = 512
)

// stop words — common English words that add no search value
var stopWords = map[string]bool{
	"the": true, "is": true, "a": true, "an": true, "and": true, "or": true,
	"but": true, "in": true, "on": true, "at": true, "to": true, "for": true,
	"of": true, "with": true, "by": true, "from": true, "up": true, "about": true,
	"into": true, "through": true, "this": true, "that": true, "these": true,
	"those": true, "it": true, "its": true, "be": true, "was": true, "were": true,
	"are": true, "been": true, "have": true, "has": true, "had": true, "do": true,
	"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
	"may": true, "might": true, "can": true, "as": true, "if": true, "then": true,
}

// ScoredID is a scored search result.
type ScoredID struct {
	ID    [16]byte
	Score float64
}

// PostingValue is the 7-byte per-posting entry value.
type PostingValue struct {
	TF     float32
	Field  uint8
	DocLen uint16
}

// Index is the FTS inverted index backed by Pebble.
type Index struct {
	db       *pebble.DB
	mu       sync.RWMutex
	// In-memory IDF cache: term → idf
	idfCache map[string]float64
}

func New(db *pebble.DB) *Index {
	return &Index{
		db:       db,
		idfCache: make(map[string]float64, 1024),
	}
}

// InvalidateIDFCache clears the in-memory IDF cache, forcing fresh recalculation
// on the next search. Call this after a vault clear to prevent stale IDF values
// from influencing BM25 scoring.
func (idx *Index) InvalidateIDFCache() {
	idx.mu.Lock()
	idx.idfCache = make(map[string]float64)
	idx.mu.Unlock()
}

// Tokenize lowercases, removes non-letter chars, filters stop words.
func Tokenize(text string) []string {
	text = strings.ToLower(text)
	var b strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	tokens := strings.Fields(b.String())
	result := tokens[:0]
	for _, t := range tokens {
		if len(t) < 2 {
			continue
		}
		if stopWords[t] {
			continue
		}
		if len([]rune(t)) > 64 {
			t = string([]rune(t)[:64])
		}
		result = append(result, t)
	}
	return result
}

// Trigrams extracts 3-character windows from a term.
func Trigrams(term string) [][3]byte {
	if len(term) < 3 {
		return nil
	}
	var result [][3]byte
	for i := 0; i+2 < len(term); i++ {
		result = append(result, [3]byte{term[i], term[i+1], term[i+2]})
	}
	return result
}

// encodePosting encodes a PostingValue into 7 bytes.
func encodePosting(pv PostingValue) []byte {
	buf := make([]byte, 7)
	binary.BigEndian.PutUint32(buf[0:4], math.Float32bits(pv.TF))
	buf[4] = pv.Field
	binary.BigEndian.PutUint16(buf[5:7], pv.DocLen)
	return buf
}

// decodePosting decodes 7 bytes into a PostingValue.
func decodePosting(buf []byte) PostingValue {
	if len(buf) < 7 {
		return PostingValue{}
	}
	return PostingValue{
		TF:     math.Float32frombits(binary.BigEndian.Uint32(buf[0:4])),
		Field:  buf[4],
		DocLen: binary.BigEndian.Uint16(buf[5:7]),
	}
}

// fieldWeight returns the scoring weight for a field.
func fieldWeight(field uint8) float64 {
	switch field {
	case FieldConcept:
		return fieldWeightConcept
	case FieldTags:
		return fieldWeightTags
	case FieldContent:
		return fieldWeightContent
	case FieldCreatedBy:
		return fieldWeightCreatedBy
	default:
		return 1.0
	}
}

// IndexEngram writes FTS posting list entries for an engram.
// ws is the 8-byte workspace prefix. id is the ULID.
func (idx *Index) IndexEngram(ws [8]byte, id [16]byte, concept, createdBy, content string, tags []string) error {
	batch := idx.db.NewBatch()

	// Collect all (term, field, docLen) tuples
	type entry struct {
		term  string
		field uint8
		tf    float32
	}
	termCounts := make(map[string]map[uint8]int)
	addTerms := func(text string, field uint8) {
		tokens := Tokenize(text)
		for _, t := range tokens {
			if termCounts[t] == nil {
				termCounts[t] = make(map[uint8]int)
			}
			termCounts[t][field]++
		}
	}

	addTerms(concept, FieldConcept)
	addTerms(createdBy, FieldCreatedBy)
	addTerms(content, FieldContent)
	for _, tag := range tags {
		addTerms(tag, FieldTags)
	}

	// Total doc len for BM25 normalization — must include all indexed fields.
	allTokens := Tokenize(concept + " " + content + " " + createdBy + " " + strings.Join(tags, " "))
	docLen := uint16(len(allTokens))

	for term, fieldCounts := range termCounts {
		for field, count := range fieldCounts {
			pv := PostingValue{
				TF:     float32(count),
				Field:  field,
				DocLen: docLen,
			}
			key := keys.FTSPostingKey(ws, term, id)
			val := encodePosting(pv)
			batch.Set(key, val, nil)
		}

		// Write trigrams
		for _, tri := range Trigrams(term) {
			tkey := keys.TrigramKey(ws, tri, id)
			batch.Set(tkey, nil, nil)
		}
	}

	if err := batch.Commit(pebble.NoSync); err != nil {
		return err
	}

	// Update per-term document frequency (df) for each unique term.
	// Hold mu across the full read-modify-write to prevent lost updates
	// under concurrent indexing calls.
	for term := range termCounts {
		tkey := keys.TermStatsKey(ws, term)
		idx.mu.Lock()
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

		// Invalidate IDF cache for this term so it's recalculated
		delete(idx.idfCache, term)
		idx.mu.Unlock()
	}

	// Update global stats (TotalEngrams, AvgDocLen)
	return idx.UpdateStats(ws, int(docLen))
}

// Search performs a BM25 search for the given query string.
func (idx *Index) Search(ctx context.Context, ws [8]byte, query string, topK int) ([]ScoredID, error) {
	tokens := Tokenize(query)
	if len(tokens) == 0 {
		return nil, nil
	}

	// Read global stats
	stats := idx.readStats(ws)
	N := float64(stats.TotalEngrams)
	avgdl := float64(stats.AvgDocLen)
	if avgdl <= 0 {
		avgdl = 1
	}

	// Guard against zero avgdl before the BM25 loop to prevent division by zero
	// in the b*dl/avgdl term, even if readStats returns a zero value.
	if avgdl <= 0 {
		avgdl = 1.0
	}

	// Per-engram accumulated scores
	scores := make(map[[16]byte]float64)

	for _, term := range tokens {
		idf := idx.getIDF(ws, term, N)
		if idf <= 0 {
			continue
		}

		// Prefix scan for this term
		lowerBound := keys.FTSPostingKey(ws, term, [16]byte{})
		// Upper bound: increment the separator byte after the term prefix.
		// Allocate one byte longer than lowerBound so sepPos is always in bounds
		// even when the term extends to the last position of lowerBound.
		upperBound := make([]byte, len(lowerBound)+1)
		copy(upperBound, lowerBound)
		// Increment the separator byte position
		sepPos := 1 + 8 + len(term)
		upperBound[sepPos] = 0x01

		iter, err := idx.db.NewIter(&pebble.IterOptions{
			LowerBound: lowerBound,
			UpperBound: upperBound,
		})
		if err != nil {
			continue
		}

		for iter.First(); iter.Valid(); iter.Next() {
			key := iter.Key()
			if len(key) < 1+8+len(term)+1+16 {
				continue
			}
			var engramID [16]byte
			copy(engramID[:], key[1+8+len(term)+1:])

			val := iter.Value()
			pv := decodePosting(val)

			tf := float64(pv.TF)
			dl := float64(pv.DocLen)
			if dl < 1 {
				dl = avgdl
			}

			// BM25 formula
			tfNorm := tf * (k1 + 1) / (tf + k1*(1-b+b*dl/avgdl))
			bm25 := idf * tfNorm * fieldWeight(pv.Field)

			// Guard against NaN/Inf scores that corrupt sorting
			if math.IsNaN(bm25) || math.IsInf(bm25, 0) {
				continue
			}

			scores[engramID] += bm25
		}
		iter.Close()
	}

	// Sort by score descending
	results := make([]ScoredID, 0, len(scores))
	for id, score := range scores {
		results = append(results, ScoredID{ID: id, Score: score})
	}
	sortScoredIDs(results)

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}
	return results, nil
}

// sortScoredIDs sorts in descending order by score.
func sortScoredIDs(s []ScoredID) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j].Score > s[j-1].Score; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// FTSStats holds global FTS statistics.
type FTSStats struct {
	TotalEngrams uint64
	AvgDocLen    float32
	VocabSize    uint64
}

// encodeStats encodes FTSStats to 20 bytes.
func encodeStats(st FTSStats) []byte {
	buf := make([]byte, 20)
	binary.BigEndian.PutUint64(buf[0:8], st.TotalEngrams)
	binary.BigEndian.PutUint32(buf[8:12], math.Float32bits(st.AvgDocLen))
	binary.BigEndian.PutUint64(buf[12:20], st.VocabSize)
	return buf
}

// decodeStats decodes 20 bytes into FTSStats.
func decodeStats(buf []byte) FTSStats {
	if len(buf) < 20 {
		return FTSStats{}
	}
	return FTSStats{
		TotalEngrams: binary.BigEndian.Uint64(buf[0:8]),
		AvgDocLen:    math.Float32frombits(binary.BigEndian.Uint32(buf[8:12])),
		VocabSize:    binary.BigEndian.Uint64(buf[12:20]),
	}
}

func (idx *Index) readStats(ws [8]byte) FTSStats {
	key := keys.FTSStatsKey(ws)
	val, closer, err := idx.db.Get(key)
	if err != nil {
		return FTSStats{TotalEngrams: 1, AvgDocLen: 100}
	}
	defer closer.Close()
	return decodeStats(val)
}

func (idx *Index) getIDF(ws [8]byte, term string, N float64) float64 {
	idx.mu.RLock()
	idf, ok := idx.idfCache[term]
	idx.mu.RUnlock()
	if ok {
		return idf
	}

	key := keys.TermStatsKey(ws, term)
	val, closer, err := idx.db.Get(key)
	if err != nil || len(val) < 8 {
		return 0
	}
	defer closer.Close()

	df := float64(binary.BigEndian.Uint32(val[0:4]))
	idf = math.Log((N-df+0.5)/(df+0.5) + 1)

	idx.mu.Lock()
	// Double-check: another goroutine may have populated the cache while we
	// held no lock (between RUnlock above and this Lock).
	if cached, ok := idx.idfCache[term]; ok {
		idx.mu.Unlock()
		return cached
	}
	idx.idfCache[term] = idf
	idx.mu.Unlock()
	return idf
}

// UpdateStats increments the engram count and recalculates avgdl.
// The read-modify-write on the Pebble stats key is protected by idx.mu to prevent
// concurrent IndexEngram calls from producing a lost-update race.
func (idx *Index) UpdateStats(ws [8]byte, docLen int) error {
	key := keys.FTSStatsKey(ws)

	idx.mu.Lock()
	defer idx.mu.Unlock()

	val, closer, err := idx.db.Get(key)
	var st FTSStats
	if err == nil {
		st = decodeStats(val)
		closer.Close()
	}

	// Rolling average of doc length
	oldTotal := float64(st.TotalEngrams) * float64(st.AvgDocLen)
	st.TotalEngrams++
	st.AvgDocLen = float32((oldTotal + float64(docLen)) / float64(st.TotalEngrams))

	return idx.db.Set(key, encodeStats(st), pebble.NoSync)
}
