package funscript

import (
	"strings"
	"testing"
)

func TestParseFunscript(t *testing.T) {
	jsonContent := `{
		"actions": [
			{"at": 0, "pos": 50},
			{"at": 1000, "pos": 100},
			{"at": 2000, "pos": 0}
		],
		"metadata": {
			"bookmarks": [
				{"name": "Start", "time": 0.0}
			]
		}
	}`

	r := strings.NewReader(jsonContent)
	script, multi, err := ParseFunscript(r, "test_script", "test_path")
	if err != nil {
		t.Fatalf("Failed to parse funscript: %v", err)
	}

	if script == nil {
		t.Fatal("Main script is nil")
	}

	if len(multi) != 0 {
		t.Errorf("Expected 0 multi-axis scripts, got %d", len(multi))
	}

	if len(script.Keyframes) != 3 {
		t.Fatalf("Expected 3 keyframes, got %d", len(script.Keyframes))
	}

	// Test conversions (at is in ms in JSON, stored in seconds in struct)
	if script.Keyframes[0].At != 0 || script.Keyframes[0].Pos != 0.5 {
		t.Errorf("First keyframe mismatch: %+v", script.Keyframes[0])
	}
	if script.Keyframes[1].At != 1.0 || script.Keyframes[1].Pos != 1.0 {
		t.Errorf("Second keyframe mismatch: %+v", script.Keyframes[1])
	}
	if script.Keyframes[2].At != 2.0 || script.Keyframes[2].Pos != 0.0 {
		t.Errorf("Third keyframe mismatch: %+v", script.Keyframes[2])
	}
}

func TestEvaluateInterpolation(t *testing.T) {
	script := &Script{
		Keyframes: []Keyframe{
			{At: 0.0, Pos: 0.0},
			{At: 1.0, Pos: 1.0},
			{At: 2.0, Pos: 0.0},
		},
	}

	tests := []struct {
		time     float64
		interp   string
		expected float64
	}{
		{0.0, "linear", 0.0},
		{0.5, "linear", 0.5},
		{1.0, "linear", 1.0},
		{1.5, "linear", 0.5},
		{2.0, "linear", 0.0},

		// Step
		{0.0, "step", 0.0},
		{0.5, "step", 0.0},
		{1.0, "step", 1.0},
		{1.9, "step", 1.0},
		{2.0, "step", 0.0},
	}

	for _, tc := range tests {
		val := script.Evaluate(tc.time, tc.interp)
		if val != tc.expected {
			t.Errorf("For time %f (%s) expected %f, got %f", tc.time, tc.interp, tc.expected, val)
		}
	}
}
