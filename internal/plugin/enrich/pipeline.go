package enrich

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/scrypster/muninndb/internal/plugin"
	"github.com/scrypster/muninndb/internal/storage"
)

// EnrichmentPipeline orchestrates the 4 LLM calls per engram.
type EnrichmentPipeline struct {
	provider LLMProvider
	prompts  *Prompts
	limiter  *TokenBucketLimiter
}

// NewPipeline creates a new enrichment pipeline.
func NewPipeline(provider LLMProvider, limiter *TokenBucketLimiter) *EnrichmentPipeline {
	return &EnrichmentPipeline{
		provider: provider,
		prompts:  DefaultPrompts(),
		limiter:  limiter,
	}
}

// Run executes the enrichment pipeline for one engram.
// Returns nil, nil if all calls fail (graceful degradation).
// Returns error only if the entire pipeline is completely unavailable.
func (p *EnrichmentPipeline) Run(ctx context.Context, eng *storage.Engram) (result *plugin.EnrichmentResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("enrich pipeline panic: %v", r)
			slog.Error("enrich: panic recovered", "panic", r)
		}
	}()

	result = &plugin.EnrichmentResult{}

	// Call 1: Entity extraction
	entities, err := p.extractEntities(ctx, eng)
	if err != nil {
		slog.Warn("enrich: entity extraction failed", "id", eng.ID.String(), "err", err)
		entities = nil
	}
	result.Entities = entities

	// Call 2: Relationship extraction (only if we have entities)
	if len(entities) > 0 {
		rels, err := p.extractRelationships(ctx, eng, entities)
		if err != nil {
			slog.Warn("enrich: relationship extraction failed", "id", eng.ID.String(), "err", err)
			rels = nil
		}
		result.Relationships = rels
	}

	// Call 3: Classification (independent)
	memType, category, subcategory, tags, err := p.classify(ctx, eng)
	if err != nil {
		slog.Warn("enrich: classification failed", "id", eng.ID.String(), "err", err)
	} else {
		result.MemoryType = memType
		if category != "" && subcategory != "" {
			result.Classification = category + "/" + subcategory
		}
		// Tags are currently not stored in EnrichmentResult; could be added if needed
		_ = tags
	}

	// Call 4: Summarization (independent)
	summary, keyPoints, err := p.summarize(ctx, eng)
	if err != nil {
		slog.Warn("enrich: summarization failed", "id", eng.ID.String(), "err", err)
	} else {
		result.Summary = summary
		result.KeyPoints = keyPoints
	}

	// If ALL four calls failed, return error so retry can be attempted
	if result.Summary == "" && len(result.Entities) == 0 &&
		result.MemoryType == "" && result.Classification == "" {
		return nil, fmt.Errorf("enrich: all pipeline stages failed for engram %s", eng.ID.String())
	}

	return result, nil
}

// extractEntities executes Call 1: entity extraction.
func (p *EnrichmentPipeline) extractEntities(ctx context.Context, eng *storage.Engram) ([]plugin.ExtractedEntity, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	userMsg := fmt.Sprintf("Concept: %s\n\nContent: %s", eng.Concept, eng.Content)
	resp, err := p.provider.Complete(ctx, p.prompts.EntitiesSystem, userMsg)
	if err != nil {
		return nil, err
	}

	return ParseEntityResponse(resp)
}

// extractRelationships executes Call 2: relationship extraction.
func (p *EnrichmentPipeline) extractRelationships(ctx context.Context, eng *storage.Engram, entities []plugin.ExtractedEntity) ([]plugin.ExtractedRelation, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Build entities JSON for the prompt
	entitiesJSON := "["
	for i, e := range entities {
		if i > 0 {
			entitiesJSON += ", "
		}
		entitiesJSON += fmt.Sprintf(`{"name": %q, "type": %q, "confidence": %.2f}`, e.Name, e.Type, e.Confidence)
	}
	entitiesJSON += "]"

	userMsg := fmt.Sprintf("Entities: %s\n\nConcept: %s\n\nContent: %s",
		entitiesJSON, eng.Concept, eng.Content)
	resp, err := p.provider.Complete(ctx, p.prompts.RelationshipsSystem, userMsg)
	if err != nil {
		return nil, err
	}

	return ParseRelationshipResponse(resp)
}

// classify executes Call 3: classification.
func (p *EnrichmentPipeline) classify(ctx context.Context, eng *storage.Engram) (memType, category, subcategory string, tags []string, err error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return "", "", "", nil, err
	}

	userMsg := fmt.Sprintf("Concept: %s\n\nContent: %s", eng.Concept, eng.Content)
	resp, err := p.provider.Complete(ctx, p.prompts.ClassifySystem, userMsg)
	if err != nil {
		return "", "", "", nil, err
	}

	return ParseClassificationResponse(resp)
}

// summarize executes Call 4: summarization.
func (p *EnrichmentPipeline) summarize(ctx context.Context, eng *storage.Engram) (summary string, keyPoints []string, err error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return "", nil, err
	}

	userMsg := fmt.Sprintf("Concept: %s\n\nContent: %s", eng.Concept, eng.Content)
	resp, err := p.provider.Complete(ctx, p.prompts.SummarizeSystem, userMsg)
	if err != nil {
		return "", nil, err
	}

	return ParseSummarizeResponse(resp)
}
