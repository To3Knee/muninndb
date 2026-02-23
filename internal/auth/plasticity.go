package auth

// PlasticityConfig is the per-vault cognitive pipeline configuration.
// A nil PlasticityConfig means "use defaults" (equivalent to Preset: "default").
// Non-nil pointer fields override the chosen preset value.
type PlasticityConfig struct {
	Version int    `json:"version,omitempty"` // schema version, currently 1
	Preset  string `json:"preset,omitempty"`  // "default" | "reference" | "scratchpad" | "knowledge-graph"

	// Optional overrides (nil = use preset value)
	HebbianEnabled *bool    `json:"hebbian_enabled,omitempty"`
	DecayEnabled   *bool    `json:"decay_enabled,omitempty"`
	HopDepth       *int     `json:"hop_depth,omitempty"`       // BFS hops 0–8
	SemanticWeight *float32 `json:"semantic_weight,omitempty"` // 0–1
	FTSWeight      *float32 `json:"fts_weight,omitempty"`      // 0–1
	DecayFloor     *float32 `json:"decay_floor,omitempty"`     // 0–1
	DecayStability  *float32 `json:"decay_stability,omitempty"`  // days
	TraversalProfile *string  `json:"traversal_profile,omitempty"` // "default"|"causal"|"confirmatory"|"adversarial"|"structural"; empty = use auto-inference
}

// ResolvedPlasticity is the fully-merged configuration after applying preset defaults
// and any field-level overrides from PlasticityConfig.
// Weight fields (SemanticWeight, FTSWeight, HebbianWeight, DecayWeight, RecencyWeight)
// are independent multipliers and are not required to sum to 1.0.
// The engine may normalize or use them as-is depending on the activation context.
type ResolvedPlasticity struct {
	HebbianEnabled bool    `json:"hebbian_enabled"`
	DecayEnabled   bool    `json:"decay_enabled"`
	HopDepth       int     `json:"hop_depth"`
	SemanticWeight float32 `json:"semantic_weight"`
	FTSWeight      float32 `json:"fts_weight"`
	DecayFloor     float32 `json:"decay_floor"`
	DecayStability float32 `json:"decay_stability"` // days
	HebbianWeight  float32 `json:"hebbian_weight"`
	DecayWeight    float32 `json:"decay_weight"`
	RecencyWeight    float32 `json:"recency_weight"`
	TraversalProfile string  `json:"traversal_profile"` // empty string = use auto-inference
}

type plasticityPreset struct {
	HebbianEnabled bool
	DecayEnabled   bool
	HopDepth       int
	SemanticWeight float32
	FTSWeight      float32
	DecayFloor     float32
	DecayStability float32
	HebbianWeight  float32
	DecayWeight    float32
	RecencyWeight  float32
}

var plasticityPresets = map[string]plasticityPreset{
	"default": {
		HebbianEnabled: true,
		DecayEnabled:   true,
		HopDepth:       2,
		SemanticWeight: 0.6,
		FTSWeight:      0.3,
		DecayFloor:     0.05,
		DecayStability: 30,
		HebbianWeight:  0.5,
		DecayWeight:    0.4,
		RecencyWeight:  0.3,
	},
	"reference": {
		HebbianEnabled: true,
		DecayEnabled:   false,
		HopDepth:       3,
		SemanticWeight: 0.7,
		FTSWeight:      0.5,
		DecayFloor:     1.0,
		DecayStability: 365,
		HebbianWeight:  0.6,
		DecayWeight:    0.0,
		RecencyWeight:  0.1,
	},
	"scratchpad": {
		HebbianEnabled: false,
		DecayEnabled:   true,
		HopDepth:       0,
		SemanticWeight: 0.5,
		FTSWeight:      0.4,
		DecayFloor:     0.01,
		DecayStability: 7,
		HebbianWeight:  0.0,
		DecayWeight:    0.8,
		RecencyWeight:  0.5,
	},
	"knowledge-graph": {
		HebbianEnabled: true,
		DecayEnabled:   true,
		HopDepth:       4,
		SemanticWeight: 0.5,
		FTSWeight:      0.2,
		DecayFloor:     0.1,
		DecayStability: 60,
		HebbianWeight:  0.8,
		DecayWeight:    0.2,
		RecencyWeight:  0.2,
	},
}

// ResolvePlasticity merges cfg (which may be nil) atop its chosen preset,
// returning a fully-populated ResolvedPlasticity.
func ResolvePlasticity(cfg *PlasticityConfig) ResolvedPlasticity {
	presetName := "default"
	if cfg != nil && cfg.Preset != "" {
		presetName = cfg.Preset
	}
	p, ok := plasticityPresets[presetName]
	if !ok {
		p = plasticityPresets["default"]
	}

	r := ResolvedPlasticity{
		HebbianEnabled: p.HebbianEnabled,
		DecayEnabled:   p.DecayEnabled,
		HopDepth:       p.HopDepth,
		SemanticWeight: p.SemanticWeight,
		FTSWeight:      p.FTSWeight,
		DecayFloor:     p.DecayFloor,
		DecayStability: p.DecayStability,
		HebbianWeight:  p.HebbianWeight,
		DecayWeight:    p.DecayWeight,
		RecencyWeight:  p.RecencyWeight,
	}

	if cfg == nil {
		return r
	}

	// Apply pointer-field overrides
	if cfg.HebbianEnabled != nil {
		r.HebbianEnabled = *cfg.HebbianEnabled
	}
	if cfg.DecayEnabled != nil {
		r.DecayEnabled = *cfg.DecayEnabled
	}
	if cfg.HopDepth != nil {
		r.HopDepth = *cfg.HopDepth
		if r.HopDepth < 0 {
			r.HopDepth = 0
		}
		if r.HopDepth > 8 {
			r.HopDepth = 8
		}
	}
	if cfg.SemanticWeight != nil {
		r.SemanticWeight = *cfg.SemanticWeight
		if r.SemanticWeight < 0 {
			r.SemanticWeight = 0
		}
		if r.SemanticWeight > 1 {
			r.SemanticWeight = 1
		}
	}
	if cfg.FTSWeight != nil {
		r.FTSWeight = *cfg.FTSWeight
		if r.FTSWeight < 0 {
			r.FTSWeight = 0
		}
		if r.FTSWeight > 1 {
			r.FTSWeight = 1
		}
	}
	if cfg.DecayFloor != nil {
		r.DecayFloor = *cfg.DecayFloor
		if r.DecayFloor < 0 {
			r.DecayFloor = 0
		}
		if r.DecayFloor > 1 {
			r.DecayFloor = 1
		}
	}
	if cfg.DecayStability != nil {
		stability := *cfg.DecayStability
		if stability > 0 {
			r.DecayStability = stability
		}
	}
	if cfg.TraversalProfile != nil {
		r.TraversalProfile = *cfg.TraversalProfile
	}

	return r
}

// ValidPlasticityPreset returns true if s is a known preset name.
func ValidPlasticityPreset(s string) bool {
	_, ok := plasticityPresets[s]
	return ok
}
