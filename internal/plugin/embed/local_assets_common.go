package embed

import _ "embed"

// embeddedModel is the all-MiniLM-L6-v2 INT8 ONNX model (~23 MB).
// It is embedded at build time from the assets directory.
// Run `make fetch-assets` before building to populate this file.
//
//go:embed assets/model_int8.onnx
var embeddedModel []byte

// embeddedTokenizer is the HuggingFace tokenizer.json for all-MiniLM-L6-v2.
//
//go:embed assets/tokenizer.json
var embeddedTokenizer []byte

// LocalAvailable reports whether the bundled ONNX model and tokenizer were
// embedded at build time (i.e. make fetch-assets was run before building).
func LocalAvailable() bool {
	return len(embeddedModel) > 0 && len(embeddedTokenizer) > 0
}
