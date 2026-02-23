package auth

import "testing"

func TestResolvePlasticity_NilUsesDefault(t *testing.T) {
	r := ResolvePlasticity(nil)
	if r.HopDepth != 2 {
		t.Errorf("want HopDepth=2, got %d", r.HopDepth)
	}
	if !r.HebbianEnabled {
		t.Error("want HebbianEnabled=true")
	}
	if !r.DecayEnabled {
		t.Error("want DecayEnabled=true")
	}
}

func TestResolvePlasticity_ScratchpadPreset(t *testing.T) {
	r := ResolvePlasticity(&PlasticityConfig{Preset: "scratchpad"})
	if r.HopDepth != 0 {
		t.Errorf("scratchpad HopDepth want 0, got %d", r.HopDepth)
	}
	if r.HebbianEnabled {
		t.Error("scratchpad: want HebbianEnabled=false")
	}
	if r.DecayStability != 7 {
		t.Errorf("scratchpad DecayStability want 7, got %f", r.DecayStability)
	}
}

func TestResolvePlasticity_ReferencePreset(t *testing.T) {
	r := ResolvePlasticity(&PlasticityConfig{Preset: "reference"})
	if r.DecayEnabled {
		t.Error("reference: want DecayEnabled=false")
	}
	if r.DecayFloor != 1.0 {
		t.Errorf("reference DecayFloor want 1.0, got %f", r.DecayFloor)
	}
}

func TestResolvePlasticity_KnowledgeGraphPreset(t *testing.T) {
	r := ResolvePlasticity(&PlasticityConfig{Preset: "knowledge-graph"})
	if r.HopDepth != 4 {
		t.Errorf("knowledge-graph HopDepth want 4, got %d", r.HopDepth)
	}
}

func TestResolvePlasticity_PointerOverride(t *testing.T) {
	hd := 5
	r := ResolvePlasticity(&PlasticityConfig{
		Preset:   "default",
		HopDepth: &hd,
	})
	if r.HopDepth != 5 {
		t.Errorf("override HopDepth want 5, got %d", r.HopDepth)
	}
	if !r.HebbianEnabled {
		t.Error("want HebbianEnabled=true (from default)")
	}
}

func TestResolvePlasticity_BoolOverride(t *testing.T) {
	f := false
	r := ResolvePlasticity(&PlasticityConfig{
		Preset:         "default",
		HebbianEnabled: &f,
	})
	if r.HebbianEnabled {
		t.Error("explicit false override should set HebbianEnabled=false")
	}
}

func TestResolvePlasticity_InvalidPresetFallsToDefault(t *testing.T) {
	r := ResolvePlasticity(&PlasticityConfig{Preset: "bogus"})
	if r.HopDepth != 2 {
		t.Errorf("invalid preset should fall to default, want HopDepth=2, got %d", r.HopDepth)
	}
}

func TestValidPlasticityPreset(t *testing.T) {
	if !ValidPlasticityPreset("default")         { t.Error("default should be valid") }
	if !ValidPlasticityPreset("reference")       { t.Error("reference should be valid") }
	if !ValidPlasticityPreset("scratchpad")      { t.Error("scratchpad should be valid") }
	if !ValidPlasticityPreset("knowledge-graph") { t.Error("knowledge-graph should be valid") }
	if ValidPlasticityPreset("bogus")            { t.Error("bogus should not be valid") }
}

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

func TestPlasticityConfig_TraversalProfilePresetOverride(t *testing.T) {
	// Setting TraversalProfile as override works regardless of preset
	profile := "adversarial"
	cfg := &PlasticityConfig{
		Preset:           "knowledge-graph",
		TraversalProfile: &profile,
	}
	r := ResolvePlasticity(cfg)
	if r.TraversalProfile != "adversarial" {
		t.Errorf("expected 'adversarial', got %q", r.TraversalProfile)
	}
}

func TestPlasticityConfig_NilTraversalProfileIsEmpty(t *testing.T) {
	// When PlasticityConfig exists but TraversalProfile is not set, it should be empty (use inference)
	cfg := &PlasticityConfig{Preset: "default"}
	r := ResolvePlasticity(cfg)
	if r.TraversalProfile != "" {
		t.Errorf("unset TraversalProfile should resolve to empty string, got %q", r.TraversalProfile)
	}
}
