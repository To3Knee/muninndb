package embed

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
	ort "github.com/yalue/onnxruntime_go"
)

const (
	localModelDim    = 384   // all-MiniLM-L6-v2 output dimension
	localMaxTokens   = 256   // model max sequence length
	localMaxBatch    = 1     // sequential: one text at a time (safe for embedded ORT)
	ortSentinelFile  = ".ort_extracted" // marker written after successful extraction
)

// ortInitOnce guards the global ORT environment — there can only be one.
var (
	ortInitOnce sync.Once
	ortInitErr  error
)

// LocalProvider implements Provider using the bundled all-MiniLM-L6-v2 ONNX model.
// No external process or network connection is required; all assets are embedded
// in the binary and extracted to DataDir on first Init.
type LocalProvider struct {
	// mu protects session and tokenizer after Init.
	mu sync.Mutex

	session   *ort.AdvancedSession
	tok       *tokenizer.Tokenizer
	dataDir   string

	// pre-allocated tensors reused across EmbedBatch calls (avoids GC pressure)
	inputIDs      *ort.Tensor[int64]
	attentionMask *ort.Tensor[int64]
	tokenTypeIDs  *ort.Tensor[int64]
	outputTensor  *ort.Tensor[float32]
}

func (p *LocalProvider) Name() string { return "local" }

func (p *LocalProvider) MaxBatchSize() int { return localMaxBatch }

// Init extracts embedded assets to DataDir and initializes the ORT session.
func (p *LocalProvider) Init(ctx context.Context, cfg ProviderHTTPConfig) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dataDir := cfg.DataDir
	if dataDir == "" {
		// Fallback: use a directory next to the binary.
		dataDir = "muninndb-data"
	}

	modelDir := filepath.Join(dataDir, "models", "miniLM")
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		return 0, fmt.Errorf("local provider: cannot create model dir %s: %w", modelDir, err)
	}

	// Extract embedded assets if not already present (checked via SHA256 sentinel).
	if err := ensureExtracted(ctx, modelDir); err != nil {
		return 0, fmt.Errorf("local provider: asset extraction failed: %w", err)
	}

	// Initialize ORT global environment (once per process).
	ortLibPath := filepath.Join(modelDir, nativeLibFilename)
	ortInitOnce.Do(func() {
		ort.SetSharedLibraryPath(ortLibPath)
		ortInitErr = ort.InitializeEnvironment()
	})
	if ortInitErr != nil {
		return 0, fmt.Errorf("local provider: ORT environment init: %w", ortInitErr)
	}

	// Load tokenizer.
	tokPath := filepath.Join(modelDir, "tokenizer.json")
	tok, err := pretrained.FromFile(tokPath)
	if err != nil {
		return 0, fmt.Errorf("local provider: load tokenizer: %w", err)
	}
	p.tok = tok

	// Pre-allocate tensors for one sequence of localMaxTokens tokens.
	shape := ort.NewShape(1, int64(localMaxTokens))

	inputIDs, err := ort.NewEmptyTensor[int64](shape)
	if err != nil {
		return 0, fmt.Errorf("local provider: alloc input_ids: %w", err)
	}
	attentionMask, err := ort.NewEmptyTensor[int64](shape)
	if err != nil {
		inputIDs.Destroy()
		return 0, fmt.Errorf("local provider: alloc attention_mask: %w", err)
	}
	tokenTypeIDs, err := ort.NewEmptyTensor[int64](shape)
	if err != nil {
		inputIDs.Destroy()
		attentionMask.Destroy()
		return 0, fmt.Errorf("local provider: alloc token_type_ids: %w", err)
	}
	outputShape := ort.NewShape(1, int64(localMaxTokens), int64(localModelDim))
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		inputIDs.Destroy()
		attentionMask.Destroy()
		tokenTypeIDs.Destroy()
		return 0, fmt.Errorf("local provider: alloc output: %w", err)
	}

	p.inputIDs = inputIDs
	p.attentionMask = attentionMask
	p.tokenTypeIDs = tokenTypeIDs
	p.outputTensor = outputTensor
	p.dataDir = dataDir

	// Create the ORT session.
	modelPath := filepath.Join(modelDir, "model_int8.onnx")
	opts, err := ort.NewSessionOptions()
	if err != nil {
		return 0, fmt.Errorf("local provider: ORT session options: %w", err)
	}
	defer opts.Destroy()

	// Prefer speed over accuracy since this is INT8 quantized already.
	opts.SetIntraOpNumThreads(1) //nolint:errcheck

	session, err := ort.NewAdvancedSession(
		modelPath,
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"last_hidden_state"},
		[]ort.Value{inputIDs, attentionMask, tokenTypeIDs},
		[]ort.Value{outputTensor},
		opts,
	)
	if err != nil {
		return 0, fmt.Errorf("local provider: create ORT session: %w", err)
	}
	p.session = session

	slog.Info("local embed provider initialized",
		"model", "all-MiniLM-L6-v2",
		"dimension", localModelDim,
		"model_dir", modelDir,
	)

	return localModelDim, nil
}

// EmbedBatch encodes a single text and returns a 384-dim embedding.
// MaxBatchSize() == 1 so BatchEmbedder always calls with len(texts)==1.
func (p *LocalProvider) EmbedBatch(ctx context.Context, texts []string) ([]float32, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.session == nil {
		return nil, fmt.Errorf("local provider not initialized")
	}

	result := make([]float32, 0, len(texts)*localModelDim)

	for i, text := range texts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		vec, err := p.embedOne(text)
		if err != nil {
			return nil, fmt.Errorf("local provider: embed text[%d]: %w", i, err)
		}
		result = append(result, vec...)
	}

	return result, nil
}

// embedOne tokenizes one text, runs ORT inference, and returns a mean-pooled
// L2-normalized 384-dim vector.
func (p *LocalProvider) embedOne(text string) ([]float32, error) {
	// Tokenize with special tokens ([CLS], [SEP]).
	enc, err := p.tok.EncodeSingle(text, true)
	if err != nil {
		return nil, fmt.Errorf("tokenize: %w", err)
	}

	ids := enc.GetIds()
	mask := enc.GetAttentionMask()
	typeIDs := enc.GetTypeIds()

	// Truncate to max sequence length.
	seqLen := len(ids)
	if seqLen > localMaxTokens {
		ids = ids[:localMaxTokens]
		mask = mask[:localMaxTokens]
		typeIDs = typeIDs[:localMaxTokens]
		seqLen = localMaxTokens
	}

	// Fill pre-allocated tensor buffers.
	// Buffers are length localMaxTokens; zero out then fill the active prefix.
	inputBuf := p.inputIDs.GetData()
	maskBuf := p.attentionMask.GetData()
	typeBuf := p.tokenTypeIDs.GetData()

	for i := range inputBuf {
		inputBuf[i] = 0
		maskBuf[i] = 0
		typeBuf[i] = 0
	}
	for i := 0; i < seqLen; i++ {
		inputBuf[i] = int64(ids[i])
		maskBuf[i] = int64(mask[i])
		typeBuf[i] = int64(typeIDs[i])
	}

	// Run inference.
	if err := p.session.Run(); err != nil {
		return nil, fmt.Errorf("ORT run: %w", err)
	}

	// last_hidden_state shape: [1, localMaxTokens, localModelDim]
	// Mean pool over the non-padding tokens (where attention_mask == 1).
	hidden := p.outputTensor.GetData()
	vec := meanPool(hidden, maskBuf, seqLen, localModelDim)
	l2Normalize(vec)

	return vec, nil
}

// meanPool computes the mean of token embeddings weighted by the attention mask.
// hidden: flat [1, seqLen, dim] slice; mask: [seqLen] with 0/1 values.
func meanPool(hidden []float32, mask []int64, seqLen, dim int) []float32 {
	result := make([]float32, dim)
	var count float32
	for t := 0; t < seqLen; t++ {
		if mask[t] == 0 {
			continue
		}
		count++
		base := t * dim
		for d := 0; d < dim; d++ {
			result[d] += hidden[base+d]
		}
	}
	if count > 0 {
		for d := range result {
			result[d] /= count
		}
	}
	return result
}

// l2Normalize divides v by its L2 norm in place.
func l2Normalize(v []float32) {
	var sum float32
	for _, x := range v {
		sum += x * x
	}
	if sum == 0 {
		return
	}
	inv := float32(1.0 / math.Sqrt(float64(sum)))
	for i := range v {
		v[i] *= inv
	}
}

// Close releases ORT session and pre-allocated tensors.
func (p *LocalProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.session != nil {
		_ = p.session.Destroy()
		p.session = nil
	}
	if p.inputIDs != nil {
		_ = p.inputIDs.Destroy()
		p.inputIDs = nil
	}
	if p.attentionMask != nil {
		_ = p.attentionMask.Destroy()
		p.attentionMask = nil
	}
	if p.tokenTypeIDs != nil {
		_ = p.tokenTypeIDs.Destroy()
		p.tokenTypeIDs = nil
	}
	if p.outputTensor != nil {
		_ = p.outputTensor.Destroy()
		p.outputTensor = nil
	}
	return nil
}

// ensureExtracted writes embedded assets to modelDir if not already present.
// Uses a SHA256 sentinel file to avoid redundant extraction.
func ensureExtracted(ctx context.Context, modelDir string) error {
	sentinelPath := filepath.Join(modelDir, ortSentinelFile)
	if _, err := os.Stat(sentinelPath); err == nil {
		// Already extracted.
		return nil
	}

	slog.Info("extracting bundled local embed assets", "dir", modelDir)

	files := map[string][]byte{
		"model_int8.onnx":      embeddedModel,
		"tokenizer.json":       embeddedTokenizer,
		nativeLibFilename:      embeddedNativeLib,
	}

	var sentinelHash string
	for name, data := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(data) == 0 {
			return fmt.Errorf("embedded asset %q is empty — run `make fetch-assets` and rebuild", name)
		}

		dest := filepath.Join(modelDir, name)
		if err := atomicWrite(dest, data); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}

		// Accumulate SHA256 for sentinel.
		h := sha256.Sum256(data)
		sentinelHash += hex.EncodeToString(h[:]) + "\n"
	}

	// Write sentinel only after all files succeed.
	if err := atomicWrite(sentinelPath, []byte(sentinelHash)); err != nil {
		return fmt.Errorf("write sentinel: %w", err)
	}

	// Make the native lib executable (required on unix).
	libPath := filepath.Join(modelDir, nativeLibFilename)
	if err := os.Chmod(libPath, 0o755); err != nil {
		return fmt.Errorf("chmod native lib: %w", err)
	}

	slog.Info("local embed assets extracted", "dir", modelDir)
	return nil
}

// atomicWrite writes data to dest via a temp file + rename to prevent corruption.
func atomicWrite(dest string, data []byte) error {
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	_, writeErr := tmp.Write(data)
	closeErr := tmp.Close()
	if writeErr != nil {
		os.Remove(tmpName)
		return writeErr
	}
	if closeErr != nil {
		os.Remove(tmpName)
		return closeErr
	}

	// Atomic replace.
	if err := os.Rename(tmpName, dest); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

// Convenience reader that works with either io.Reader or raw bytes.
func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
