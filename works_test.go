package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorksNewestFirst(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "works")
	if ws, err := listWorks(dir); err != nil || len(ws) != 0 {
		t.Fatalf("a missing directory: %v, %v", ws, err)
	}
	t0 := time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC)
	older, err := saveWork(dir, tinyPNG(t), t0)
	if err != nil {
		t.Fatal(err)
	}
	newer, err := saveWork(dir, []byte("\xff\xd8\xff\xe0jpeg"), t0.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(older.Path) != ".png" || filepath.Ext(newer.Path) != ".jpg" {
		t.Errorf("extensions = %s, %s", older.Path, newer.Path)
	}
	if err := os.WriteFile(filepath.Join(dir, ".DS_Store"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ws, err := listWorks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 2 || ws[0].Path != newer.Path || ws[1].Path != older.Path || !ws[1].Created.Equal(t0) {
		t.Errorf("works = %+v", ws)
	}
}
