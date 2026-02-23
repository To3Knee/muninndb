package keys

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"testing"
)

func newTestID() [16]byte {
	var id [16]byte
	_, _ = rand.Read(id[:])
	return id
}

func TestVaultPrefix(t *testing.T) {
	vault1 := "vault1"
	vault2 := "vault2"

	prefix1 := VaultPrefix(vault1)
	prefix1Again := VaultPrefix(vault1)
	prefix2 := VaultPrefix(vault2)

	// Same vault should produce same prefix
	if prefix1 != prefix1Again {
		t.Error("Same vault produced different prefixes")
	}

	// Different vaults should produce different prefixes
	if prefix1 == prefix2 {
		t.Error("Different vaults produced same prefix")
	}
}

func TestEngramKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	id := newTestID()

	key := EngramKey(ws, id)

	// Should be 1 + 8 + 16 = 25 bytes
	if len(key) != 25 {
		t.Errorf("EngramKey length wrong: %d != 25", len(key))
	}

	// First byte should be 0x01
	if key[0] != 0x01 {
		t.Error("EngramKey prefix wrong")
	}

	// Next 8 bytes should be workspace prefix
	if [8]byte(key[1:9]) != ws {
		t.Error("EngramKey workspace prefix wrong")
	}

	// Next 16 bytes should be ULID
	if [16]byte(key[9:25]) != id {
		t.Error("EngramKey ULID wrong")
	}
}

func TestMetaKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	id := newTestID()

	key := MetaKey(ws, id)

	if len(key) != 25 {
		t.Errorf("MetaKey length wrong: %d != 25", len(key))
	}
	if key[0] != 0x02 {
		t.Error("MetaKey prefix wrong")
	}
}

func TestAssocFwdKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	src := newTestID()
	dst := newTestID()
	weight := float32(0.75)

	key := AssocFwdKey(ws, src, weight, dst)

	// Should be 1 + 8 + 16 + 4 + 16 = 45 bytes
	if len(key) != 45 {
		t.Errorf("AssocFwdKey length wrong: %d != 45", len(key))
	}

	if key[0] != 0x03 {
		t.Error("AssocFwdKey prefix wrong")
	}
}

func TestAssocRevKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	src := newTestID()
	dst := newTestID()
	weight := float32(0.75)

	key := AssocRevKey(ws, dst, weight, src)

	if len(key) != 45 {
		t.Errorf("AssocRevKey length wrong: %d != 45", len(key))
	}

	if key[0] != 0x04 {
		t.Error("AssocRevKey prefix wrong")
	}
}

func TestWeightComplement(t *testing.T) {
	// Higher weight should produce lower complement (for descending sort)
	wc1 := WeightComplement(0.9)
	wc2 := WeightComplement(0.1)

	// Convert to uint32 for comparison
	w1 := binary.BigEndian.Uint32(wc1[:])
	w2 := binary.BigEndian.Uint32(wc2[:])

	if w1 >= w2 {
		t.Error("WeightComplement: higher weight should have lower complement")
	}
}

func TestFTSPostingKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	term := "search"
	id := newTestID()

	key := FTSPostingKey(ws, term, id)

	// Should be 1 + 8 + len(term) + 1 + 16
	expectedLen := 1 + 8 + len(term) + 1 + 16
	if len(key) != expectedLen {
		t.Errorf("FTSPostingKey length wrong: %d != %d", len(key), expectedLen)
	}

	if key[0] != 0x05 {
		t.Error("FTSPostingKey prefix wrong")
	}

	// Check separator byte at position 1+8+len(term)
	if key[1+8+len(term)] != 0x00 {
		t.Error("FTSPostingKey separator byte wrong")
	}
}

func TestTrigramKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	trigram := [3]byte{'a', 'b', 'c'}
	id := newTestID()

	key := TrigramKey(ws, trigram, id)

	// Should be 1 + 8 + 3 + 16 = 28 bytes
	if len(key) != 28 {
		t.Errorf("TrigramKey length wrong: %d != 28", len(key))
	}

	if key[0] != 0x06 {
		t.Error("TrigramKey prefix wrong")
	}
}

func TestHNSWNodeKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	id := newTestID()
	layer := uint8(3)

	key := HNSWNodeKey(ws, id, layer)

	// Should be 1 + 8 + 16 + 1 = 26 bytes
	if len(key) != 26 {
		t.Errorf("HNSWNodeKey length wrong: %d != 26", len(key))
	}

	if key[0] != 0x07 {
		t.Error("HNSWNodeKey prefix wrong")
	}
	if key[25] != layer {
		t.Error("HNSWNodeKey layer byte wrong")
	}
}

func TestStateIndexKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	state := uint8(0x01) // StateActive
	id := newTestID()

	key := StateIndexKey(ws, state, id)

	// Should be 1 + 8 + 1 + 16 = 26 bytes
	if len(key) != 26 {
		t.Errorf("StateIndexKey length wrong: %d != 26", len(key))
	}

	if key[0] != 0x0B {
		t.Error("StateIndexKey prefix wrong")
	}
	if key[9] != state {
		t.Error("StateIndexKey state byte wrong")
	}
}

func TestTagIndexKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	tagHash := uint32(12345)
	id := newTestID()

	key := TagIndexKey(ws, tagHash, id)

	// Should be 1 + 8 + 4 + 16 = 29 bytes
	if len(key) != 29 {
		t.Errorf("TagIndexKey length wrong: %d != 29", len(key))
	}

	if key[0] != 0x0C {
		t.Error("TagIndexKey prefix wrong")
	}
}

func TestCreatorIndexKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	creatorHash := uint32(54321)
	id := newTestID()

	key := CreatorIndexKey(ws, creatorHash, id)

	// Should be 1 + 8 + 4 + 16 = 29 bytes
	if len(key) != 29 {
		t.Errorf("CreatorIndexKey length wrong: %d != 29", len(key))
	}

	if key[0] != 0x0D {
		t.Error("CreatorIndexKey prefix wrong")
	}
}

func TestVaultMetaKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}

	key := VaultMetaKey(ws)

	// Should be 1 + 8 = 9 bytes
	if len(key) != 9 {
		t.Errorf("VaultMetaKey length wrong: %d != 9", len(key))
	}

	if key[0] != 0x0E {
		t.Error("VaultMetaKey prefix wrong")
	}
}

func TestHash(t *testing.T) {
	h1 := Hash("test")
	h2 := Hash("test")
	h3 := Hash("different")

	if h1 != h2 {
		t.Error("Hash not consistent")
	}

	if h1 == h3 {
		t.Error("Hash collision for different strings (unlikely but possible)")
	}
}

func TestRelevanceBucketKey(t *testing.T) {
	ws := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	id := newTestID()

	tests := []struct {
		name         string
		relevance    float32
		expectedBucket uint8
	}{
		{"relevance 1.0", 1.0, 0},
		{"relevance 0.0", 0.0, 9},
		{"relevance 0.5", 0.5, 4},
		{"relevance 0.35", 0.35, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := RelevanceBucketKey(ws, tt.relevance, id)

			// Check total length: 1 + 8 + 1 + 16 = 26 bytes
			if len(key) != 26 {
				t.Errorf("RelevanceBucketKey length wrong: %d != 26", len(key))
			}

			// Check prefix byte is 0x10
			if key[0] != 0x10 {
				t.Errorf("RelevanceBucketKey prefix wrong: 0x%02x != 0x10", key[0])
			}

			// Check workspace bytes at [1:9]
			if [8]byte(key[1:9]) != ws {
				t.Error("RelevanceBucketKey workspace prefix wrong")
			}

			// Check bucket byte at [9]
			if key[9] != tt.expectedBucket {
				t.Errorf("RelevanceBucketKey bucket byte wrong: %d != %d", key[9], tt.expectedBucket)
			}

			// Check id bytes at [10:26]
			if [16]byte(key[10:26]) != id {
				t.Error("RelevanceBucketKey id wrong")
			}
		})
	}

	// Test sort order: higher relevance should sort before lower relevance
	t.Run("sort order", func(t *testing.T) {
		keyHigh := RelevanceBucketKey(ws, 0.9, id)
		keyLow := RelevanceBucketKey(ws, 0.1, id)

		// keyHigh should be less than keyLow (higher relevance sorts first)
		if bytes.Compare(keyHigh, keyLow) >= 0 {
			t.Error("Higher relevance key should sort before lower relevance key")
		}
	})

	// Test boundary: value > 1.0 clamps to bucket 0
	t.Run("relevance > 1.0", func(t *testing.T) {
		key := RelevanceBucketKey(ws, 1.5, id)
		if key[9] != 0 {
			t.Errorf("relevance > 1.0 should clamp to bucket 0, got %d", key[9])
		}
	})

	// Test boundary: negative value clamps to bucket 9 (lowest relevance)
	t.Run("relevance < 0.0", func(t *testing.T) {
		key := RelevanceBucketKey(ws, -0.5, id)
		if key[9] != 9 {
			t.Errorf("relevance < 0.0 should clamp to bucket 9, got %d", key[9])
		}
	})
}
