package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type OllamaProvider struct {
	client  *http.Client
	baseURL string
	model   string
}

type ollamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

func (p *OllamaProvider) Init(ctx context.Context, cfg ProviderHTTPConfig) (int, error) {
	p.baseURL = cfg.BaseURL
	p.model = cfg.Model

	// Create HTTP client with 30s timeout (local, may be loading model)
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	p.client = &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	// Probe connectivity with root GET
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, "GET", p.baseURL, nil)
	if err != nil {
		return 0, fmt.Errorf("cannot connect to Ollama at %s — is it running? (%w)", p.baseURL, err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cannot connect to Ollama at %s — is it running? (%w)", p.baseURL, err)
	}
	resp.Body.Close()

	// Embed probe text to detect dimension
	embedCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	body, _ := json.Marshal(ollamaEmbedRequest{
		Model:  p.model,
		Prompt: "dimension detection probe",
	})

	embedReq, err := http.NewRequestWithContext(embedCtx, "POST",
		p.baseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("cannot create embed request: %w", err)
	}
	embedReq.Header.Set("Content-Type", "application/json")

	embedResp, err := p.client.Do(embedReq)
	if err != nil {
		return 0, fmt.Errorf("cannot connect to Ollama at %s — is it running? (%w)", p.baseURL, err)
	}
	defer embedResp.Body.Close()

	if embedResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(embedResp.Body)
		return 0, fmt.Errorf("Ollama returned status %d: %s", embedResp.StatusCode, string(bodyBytes))
	}

	var ollamaResp ollamaEmbedResponse
	if err := json.NewDecoder(embedResp.Body).Decode(&ollamaResp); err != nil {
		return 0, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	if len(ollamaResp.Embedding) == 0 {
		return 0, fmt.Errorf("Ollama returned empty embedding")
	}

	dim := len(ollamaResp.Embedding)
	slog.Info("Ollama dimension probe successful", "dimension", dim)

	return dim, nil
}

func (p *OllamaProvider) EmbedBatch(ctx context.Context, texts []string) ([]float32, error) {
	// Ollama embeds one text at a time — loop and concatenate
	result := make([]float32, 0)

	for _, text := range texts {
		body, _ := json.Marshal(ollamaEmbedRequest{
			Model:  p.model,
			Prompt: text,
		})

		req, err := http.NewRequestWithContext(ctx, "POST",
			p.baseURL+"/api/embeddings", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("ollama embed: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ollama embed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		var ollamaResp ollamaEmbedResponse
		if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
			return nil, fmt.Errorf("ollama decode: %w", err)
		}

		for _, v := range ollamaResp.Embedding {
			result = append(result, float32(v))
		}
	}

	return result, nil
}

func (p *OllamaProvider) MaxBatchSize() int {
	// Ollama embeds one at a time
	return 1
}

func (p *OllamaProvider) Close() error {
	if p.client != nil {
		p.client.CloseIdleConnections()
	}
	return nil
}
