package scores

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "scores.json")
	return &Store{path: path, Scores: GameScores{}}
}

func TestLoadMissingFile(t *testing.T) {
	s, err := LoadFrom(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if s.Get("yahtzee") != nil {
		t.Error("expected nil for missing game")
	}
}

func TestSaveAndLoad(t *testing.T) {
	s := tempStore(t)
	s.Update("yahtzee", 287, false)
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	s2, err := LoadFrom(s.path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	e := s2.Get("yahtzee")
	if e == nil || e.Value != 287 {
		t.Errorf("got %v, want 287", e)
	}
}

func TestUpdateHigherIsBetter(t *testing.T) {
	s := tempStore(t)

	if !s.Update("yahtzee", 200, false) {
		t.Error("first score should always be a high score")
	}
	if s.Update("yahtzee", 150, false) {
		t.Error("lower score should not beat higher")
	}
	if s.Update("yahtzee", 200, false) {
		t.Error("equal score should not beat current")
	}
	if !s.Update("yahtzee", 300, false) {
		t.Error("higher score should beat current")
	}
	if s.Get("yahtzee").Value != 300 {
		t.Errorf("got %d, want 300", s.Get("yahtzee").Value)
	}
}

func TestUpdateLowerIsBetter(t *testing.T) {
	s := tempStore(t)

	if !s.Update("wordle", 4, true) {
		t.Error("first score should always be a high score")
	}
	if s.Update("wordle", 5, true) {
		t.Error("higher score should not beat lower")
	}
	if s.Update("wordle", 4, true) {
		t.Error("equal score should not beat current")
	}
	if !s.Update("wordle", 2, true) {
		t.Error("lower score should beat current")
	}
	if s.Get("wordle").Value != 2 {
		t.Errorf("got %d, want 2", s.Get("wordle").Value)
	}
}

func TestUpdateDifficulty(t *testing.T) {
	s := tempStore(t)

	if !s.UpdateDifficulty("minesweeper", "beginner", 42, true) {
		t.Error("first score should be high score")
	}
	if !s.UpdateDifficulty("minesweeper", "intermediate", 120, true) {
		t.Error("different difficulty should be independent")
	}
	if s.UpdateDifficulty("minesweeper", "beginner", 50, true) {
		t.Error("slower time should not beat faster")
	}
	if !s.UpdateDifficulty("minesweeper", "beginner", 30, true) {
		t.Error("faster time should beat slower")
	}

	e := s.GetDifficulty("minesweeper", "beginner")
	if e == nil || e.Value != 30 {
		t.Errorf("got %v, want 30", e)
	}
	e2 := s.GetDifficulty("minesweeper", "intermediate")
	if e2 == nil || e2.Value != 120 {
		t.Errorf("got %v, want 120", e2)
	}
}

func TestSaveCreatesDirRecursively(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "c")
	s := &Store{path: filepath.Join(dir, "scores.json"), Scores: GameScores{}}
	s.Update("2048", 5000, false)
	if err := s.Save(); err != nil {
		t.Fatalf("Save with nested dir: %v", err)
	}
	if _, err := os.Stat(s.path); err != nil {
		t.Errorf("file not created: %v", err)
	}
}
