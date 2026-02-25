package scores

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Entry holds a single high score record.
type Entry struct {
	Value int    `json:"value"`
	Date  string `json:"date"`
}

// GameScores stores high scores for all games.
type GameScores struct {
	Yahtzee       *Entry            `json:"yahtzee,omitempty"`
	Blackjack     *Entry            `json:"blackjack,omitempty"`
	Wordle        *Entry            `json:"wordle,omitempty"`
	Minesweeper   map[string]*Entry `json:"minesweeper,omitempty"`
	Sudoku        map[string]*Entry `json:"sudoku,omitempty"`
	TwoFortyEight *Entry            `json:"2048,omitempty"`
}

// Store manages high score persistence.
type Store struct {
	path   string
	Scores GameScores
}

// Load reads the high scores file. Returns an empty store if the file
// doesn't exist.
func Load() (*Store, error) {
	return LoadFrom("")
}

// LoadFrom reads scores from a specific path. If path is empty, uses
// the default location (~/.cli-play/scores.json).
func LoadFrom(path string) (*Store, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return &Store{Scores: GameScores{}}, err
		}
		path = filepath.Join(home, ".cli-play", "scores.json")
	}

	s := &Store{path: path, Scores: GameScores{}}

	data, err := os.ReadFile(path) //nolint:gosec // G304: path is from UserHomeDir or test-controlled input
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}

	if err := json.Unmarshal(data, &s.Scores); err != nil {
		return s, err
	}
	return s, nil
}

// Save writes the high scores to disk.
func (s *Store) Save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.Scores, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

// Update records a score if it beats the current high score. Returns
// true if a new high score was set.
func (s *Store) Update(game string, value int, lowerIsBetter bool) bool {
	today := time.Now().Format("2006-01-02")
	entry := &Entry{Value: value, Date: today}

	current := s.Get(game)
	if current != nil {
		if lowerIsBetter && value >= current.Value {
			return false
		}
		if !lowerIsBetter && value <= current.Value {
			return false
		}
	}

	switch game {
	case "yahtzee":
		s.Scores.Yahtzee = entry
	case "blackjack":
		s.Scores.Blackjack = entry
	case "wordle":
		s.Scores.Wordle = entry
	case "2048":
		s.Scores.TwoFortyEight = entry
	}
	return true
}

// UpdateDifficulty records a score for a difficulty-based game. Returns
// true if a new high score was set.
func (s *Store) UpdateDifficulty(game, difficulty string, value int, lowerIsBetter bool) bool {
	today := time.Now().Format("2006-01-02")
	entry := &Entry{Value: value, Date: today}

	current := s.GetDifficulty(game, difficulty)
	if current != nil {
		if lowerIsBetter && value >= current.Value {
			return false
		}
		if !lowerIsBetter && value <= current.Value {
			return false
		}
	}

	switch game {
	case "minesweeper":
		if s.Scores.Minesweeper == nil {
			s.Scores.Minesweeper = make(map[string]*Entry)
		}
		s.Scores.Minesweeper[difficulty] = entry
	case "sudoku":
		if s.Scores.Sudoku == nil {
			s.Scores.Sudoku = make(map[string]*Entry)
		}
		s.Scores.Sudoku[difficulty] = entry
	}
	return true
}

// Get returns the high score entry for a game, or nil if none exists.
func (s *Store) Get(game string) *Entry {
	switch game {
	case "yahtzee":
		return s.Scores.Yahtzee
	case "blackjack":
		return s.Scores.Blackjack
	case "wordle":
		return s.Scores.Wordle
	case "2048":
		return s.Scores.TwoFortyEight
	}
	return nil
}

// GetDifficulty returns the high score entry for a difficulty-based
// game, or nil if none exists.
func (s *Store) GetDifficulty(game, difficulty string) *Entry {
	switch game {
	case "minesweeper":
		if s.Scores.Minesweeper == nil {
			return nil
		}
		return s.Scores.Minesweeper[difficulty]
	case "sudoku":
		if s.Scores.Sudoku == nil {
			return nil
		}
		return s.Scores.Sudoku[difficulty]
	}
	return nil
}
