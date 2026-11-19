package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/nsgo-dev/nsgo/core"
)

// pictures keeps the app's pictures and their sidecars: works, results not saved yet and the user's clothes. Paths are absolute.
type pictures interface {
	Write(path string, data []byte) error
	Read(path string) ([]byte, error)
	// Remove deletes a path; one already gone is fine.
	Remove(path string) error
	Rename(from, to string) error
	// List names the files directly in dir; a dir never written holds none.
	List(dir string) ([]string, error)
	// Source shows a picture in an image view.
	Source(path string) core.ImageSource
	// Lasting says whether what is kept outlives the app: false in a browser, where it lasts as long as the page.
	Lasting() bool
}

// defaultPictures is files where the platform has a file system to write, and memory in a browser, where a WebAssembly build has none.
func defaultPictures() pictures {
	if runtime.GOOS == "js" {
		return newMemPictures()
	}
	return filePictures{}
}

type filePictures struct{}

func (filePictures) Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (filePictures) Read(path string) ([]byte, error) { return os.ReadFile(path) }

func (filePictures) Remove(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (f filePictures) Rename(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	if err := os.Rename(from, to); err == nil {
		return nil
	}
	// A path on another volume, such as a cache, is copied instead.
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	if err := os.WriteFile(to, data, 0o644); err != nil {
		return err
	}
	return f.Remove(from)
}

func (filePictures) List(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func (filePictures) Source(path string) core.ImageSource { return core.Src(path) }

func (filePictures) Lasting() bool { return true }

// memPictures keeps pictures in memory, for as long as the app runs.
type memPictures struct {
	mu    sync.Mutex
	files map[string][]byte
}

func newMemPictures() *memPictures { return &memPictures{files: map[string][]byte{}} }

func (m *memPictures) Write(path string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[filepath.Clean(path)] = append([]byte(nil), data...)
	return nil
}

func (m *memPictures) Read(path string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.files[filepath.Clean(path)]
	if !ok {
		return nil, &os.PathError{Op: "read", Path: path, Err: os.ErrNotExist}
	}
	return data, nil
}

func (m *memPictures) Remove(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, filepath.Clean(path))
	return nil
}

func (m *memPictures) Rename(from, to string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.files[filepath.Clean(from)]
	if !ok {
		return &os.PathError{Op: "rename", Path: from, Err: os.ErrNotExist}
	}
	delete(m.files, filepath.Clean(from))
	m.files[filepath.Clean(to)] = data
	return nil
}

func (m *memPictures) List(dir string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prefix := filepath.Clean(dir) + string(filepath.Separator)
	var names []string
	for p := range m.files {
		if rest, ok := strings.CutPrefix(p, prefix); ok && !strings.ContainsRune(rest, filepath.Separator) {
			names = append(names, rest)
		}
	}
	sort.Strings(names)
	return names, nil
}

func (m *memPictures) Source(path string) core.ImageSource {
	data, err := m.Read(path)
	if err != nil {
		return core.ImageSource{}
	}
	return core.ImageSource{Data: data}
}

func (m *memPictures) Lasting() bool { return false }
