package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	if c.AnimationSpeed != SpeedNormal {
		t.Errorf("AnimationSpeed = %q, want %q", c.AnimationSpeed, SpeedNormal)
	}
	if c.Theme != ThemeMatrix {
		t.Errorf("Theme = %q, want %q", c.Theme, ThemeMatrix)
	}
	if c.MinesweeperDefault != "beginner" {
		t.Errorf("MinesweeperDefault = %q, want %q", c.MinesweeperDefault, "beginner")
	}
	if c.SudokuDefault != "easy" {
		t.Errorf("SudokuDefault = %q, want %q", c.SudokuDefault, "easy")
	}
}

func TestLoadFromMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	s, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom missing file: %v", err)
	}
	if s.Config.Theme != ThemeMatrix {
		t.Errorf("Theme = %q, want default %q", s.Config.Theme, ThemeMatrix)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	s, _ := LoadFrom(path)
	s.Config.Theme = ThemeAmber
	s.Config.AnimationSpeed = SpeedFast
	s.Config.MinesweeperDefault = "expert"
	s.Config.SudokuDefault = "hard"

	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if loaded.Config.Theme != ThemeAmber {
		t.Errorf("Theme = %q, want %q", loaded.Config.Theme, ThemeAmber)
	}
	if loaded.Config.AnimationSpeed != SpeedFast {
		t.Errorf("AnimationSpeed = %q, want %q", loaded.Config.AnimationSpeed, SpeedFast)
	}
	if loaded.Config.MinesweeperDefault != "expert" {
		t.Errorf("MinesweeperDefault = %q, want %q", loaded.Config.MinesweeperDefault, "expert")
	}
	if loaded.Config.SudokuDefault != "hard" {
		t.Errorf("SudokuDefault = %q, want %q", loaded.Config.SudokuDefault, "hard")
	}
}

func TestNormalizeInvalidValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Write invalid JSON values.
	data := []byte(`{
		"animation_speed": "turbo",
		"theme": "neon",
		"minesweeper_default": "nightmare",
		"sudoku_default": "impossible"
	}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	s, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if s.Config.AnimationSpeed != SpeedNormal {
		t.Errorf("AnimationSpeed = %q, want default %q", s.Config.AnimationSpeed, SpeedNormal)
	}
	if s.Config.Theme != ThemeMatrix {
		t.Errorf("Theme = %q, want default %q", s.Config.Theme, ThemeMatrix)
	}
	if s.Config.MinesweeperDefault != "beginner" {
		t.Errorf("MinesweeperDefault = %q, want default %q", s.Config.MinesweeperDefault, "beginner")
	}
	if s.Config.SudokuDefault != "easy" {
		t.Errorf("SudokuDefault = %q, want default %q", s.Config.SudokuDefault, "easy")
	}
}

func TestBlinkInterval(t *testing.T) {
	tests := []struct {
		speed AnimationSpeed
		want  int
	}{
		{SpeedSlow, 800},
		{SpeedNormal, 500},
		{SpeedFast, 250},
		{SpeedOff, 0},
	}
	for _, tt := range tests {
		c := Config{AnimationSpeed: tt.speed}
		if got := c.BlinkInterval(); got != tt.want {
			t.Errorf("BlinkInterval(%q) = %d, want %d", tt.speed, got, tt.want)
		}
	}
}

func TestTransitionTickMs(t *testing.T) {
	tests := []struct {
		speed AnimationSpeed
		want  int
	}{
		{SpeedSlow, 50},
		{SpeedNormal, 33},
		{SpeedFast, 16},
		{SpeedOff, 0},
	}
	for _, tt := range tests {
		c := Config{AnimationSpeed: tt.speed}
		if got := c.TransitionTickMs(); got != tt.want {
			t.Errorf("TransitionTickMs(%q) = %d, want %d", tt.speed, got, tt.want)
		}
	}
}

func TestSpawnRate(t *testing.T) {
	tests := []struct {
		speed AnimationSpeed
		want  float64
	}{
		{SpeedSlow, 0.08},
		{SpeedNormal, 0.15},
		{SpeedFast, 0.25},
		{SpeedOff, 0.0},
	}
	for _, tt := range tests {
		c := Config{AnimationSpeed: tt.speed}
		if got := c.SpawnRate(); got != tt.want {
			t.Errorf("SpawnRate(%q) = %f, want %f", tt.speed, got, tt.want)
		}
	}
}
