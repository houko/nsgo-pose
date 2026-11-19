package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// A work is a generated picture the user saved to the lookbook, named by the time it was made. Its sidecar, the same name with .json, keeps its number, whether it is a favorite and the references it was made from.
type work struct {
	Path    string
	Created time.Time
	workMeta
}

type workMeta struct {
	No       int   `json:"no"`
	Favorite bool  `json:"favorite,omitempty"`
	Face     *Item `json:"face,omitempty"`
	Costume  *Item `json:"costume,omitempty"`
	Pose     *Item `json:"pose,omitempty"`
}

func metaPath(picture string) string {
	return strings.TrimSuffix(picture, filepath.Ext(picture)) + ".json"
}

func writeMeta(w work) error {
	data, err := json.Marshal(w.workMeta)
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath(w.Path), data, 0o644)
}

// writePicture puts data in dir under the time it was made.
func writePicture(dir string, data []byte, at time.Time) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, fmt.Sprintf("%d%s", at.UnixNano(), imageExt(data)))
	return path, os.WriteFile(path, data, 0o644)
}

// saveWork moves a generated picture into the lookbook in dir as its next number.
func saveWork(dir, picture string, at time.Time, meta workMeta) (work, error) {
	existing, err := listWorks(dir)
	if err != nil {
		return work{}, err
	}
	for _, w := range existing {
		meta.No = max(meta.No, w.No)
	}
	meta.No++
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return work{}, err
	}
	path := filepath.Join(dir, fmt.Sprintf("%d%s", at.UnixNano(), filepath.Ext(picture)))
	if err := os.Rename(picture, path); err != nil {
		// A picture on another volume, such as a cache, is copied instead.
		data, readErr := os.ReadFile(picture)
		if readErr != nil {
			return work{}, errors.Join(err, readErr)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return work{}, err
		}
		_ = os.Remove(picture)
	}
	w := work{Path: path, Created: at, workMeta: meta}
	if err := writeMeta(w); err != nil {
		return work{}, err
	}
	return w, nil
}

// deleteWork removes a work's picture and sidecar; ones already gone are fine.
func deleteWork(w work) error {
	for _, p := range []string{w.Path, metaPath(w.Path)} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
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
		ext := filepath.Ext(name)
		stem := strings.TrimSuffix(name, ext)
		nanos, err := strconv.ParseInt(stem, 10, 64)
		if e.IsDir() || err != nil || ext == ".json" {
			continue
		}
		w := work{Path: filepath.Join(dir, name), Created: time.Unix(0, nanos)}
		if data, err := os.ReadFile(metaPath(w.Path)); err == nil {
			_ = json.Unmarshal(data, &w.workMeta)
		}
		out = append(out, w)
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

// workNo writes a work's number as the lookbook shows it, two digits at least.
func workNo(n int) string { return fmt.Sprintf("%02d", n) }
