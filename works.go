package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// A work is a generated picture kept in the app's data directory, named by the time it was made.
type work struct {
	Path    string
	Created time.Time
}

func saveWork(dir string, data []byte, at time.Time) (work, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return work{}, err
	}
	path := filepath.Join(dir, fmt.Sprintf("%d%s", at.UnixNano(), imageExt(data)))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return work{}, err
	}
	return work{Path: path, Created: at}, nil
}

// listWorks returns the works in dir, newest first; a directory not made yet holds none.
func listWorks(dir string) ([]work, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []work
	for _, e := range entries {
		name := e.Name()
		stem := strings.TrimSuffix(name, filepath.Ext(name))
		nanos, err := strconv.ParseInt(stem, 10, 64)
		if e.IsDir() || err != nil {
			continue
		}
		out = append(out, work{Path: filepath.Join(dir, name), Created: time.Unix(0, nanos)})
	}
	slices.SortFunc(out, func(a, b work) int { return b.Created.Compare(a.Created) })
	return out, nil
}

func imageExt(data []byte) string {
	switch {
	case len(data) > 3 && string(data[:3]) == "\xff\xd8\xff":
		return ".jpg"
	case len(data) > 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return ".webp"
	default:
		return ".png"
	}
}
