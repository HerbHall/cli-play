package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AnimationSpeed controls how fast animations play.
type AnimationSpeed string

const (
	SpeedSlow   AnimationSpeed = "slow"
	SpeedNormal AnimationSpeed = "normal"
	SpeedFast   AnimationSpeed = "fast"
	SpeedOff    AnimationSpeed = "off"
)

// Theme selects the color scheme.
type Theme string

const (
	ThemeMatrix Theme = "matrix"
	ThemeAmber  Theme = "amber"
	ThemeBlue   Theme = "blue"
	ThemeRed    Theme = "red"
)

// Config stores user preferences persisted to disk.
type Config struct {
	AnimationSpeed     AnimationSpeed `json:"animation_speed"`
	Theme              Theme          `json:"theme"`
	MinesweeperDefault string         `json:"minesweeper_default"`
	SudokuDefault      string         `json:"sudoku_default"`
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		AnimationSpeed:     SpeedNormal,
		Theme:              ThemeMatrix,
		MinesweeperDefault: "beginner",
		SudokuDefault:      "easy",
	}
}

// Store manages settings persistence.
type Store struct {
	path   string
	Config Config
}

// Load reads settings from the default location.
func Load() (*Store, error) {
	return LoadFrom("")
}

// LoadFrom reads settings from a specific path. If path is empty, uses
// ~/.cli-play/settings.json.
func LoadFrom(path string) (*Store, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			c := DefaultConfig()
			return &Store{Config: c}, err
		}
		path = filepath.Join(home, ".cli-play", "settings.json")
	}

	s := &Store{path: path, Config: DefaultConfig()}

	data, err := os.ReadFile(path) //nolint:gosec // G304: path is from UserHomeDir or test-controlled input
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}

	if err := json.Unmarshal(data, &s.Config); err != nil {
		return s, err
	}
	s.normalize()
	return s, nil
}

// Save writes the settings to disk.
func (s *Store) Save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.Config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

// normalize ensures all config values are valid, falling back to defaults.
func (s *Store) normalize() {
	switch s.Config.AnimationSpeed {
	case SpeedSlow, SpeedNormal, SpeedFast, SpeedOff:
	default:
		s.Config.AnimationSpeed = SpeedNormal
	}
	switch s.Config.Theme {
	case ThemeMatrix, ThemeAmber, ThemeBlue, ThemeRed:
	default:
		s.Config.Theme = ThemeMatrix
	}
	switch s.Config.MinesweeperDefault {
	case "beginner", "intermediate", "expert":
	default:
		s.Config.MinesweeperDefault = "beginner"
	}
	switch s.Config.SudokuDefault {
	case "easy", "medium", "hard":
	default:
		s.Config.SudokuDefault = "easy"
	}
}

// BlinkInterval returns the splash blink duration based on animation speed.
func (c Config) BlinkInterval() int {
	switch c.AnimationSpeed {
	case SpeedSlow:
		return 800
	case SpeedNormal:
		return 500
	case SpeedFast:
		return 250
	case SpeedOff:
		return 0
	}
	return 500
}

// TransitionTickMs returns the transition frame interval in milliseconds.
func (c Config) TransitionTickMs() int {
	switch c.AnimationSpeed {
	case SpeedSlow:
		return 50
	case SpeedNormal:
		return 33
	case SpeedFast:
		return 16
	case SpeedOff:
		return 0
	}
	return 33
}

// SpawnRate returns the rain column spawn probability per frame.
func (c Config) SpawnRate() float64 {
	switch c.AnimationSpeed {
	case SpeedSlow:
		return 0.08
	case SpeedNormal:
		return 0.15
	case SpeedFast:
		return 0.25
	case SpeedOff:
		return 0.0
	}
	return 0.15
}
