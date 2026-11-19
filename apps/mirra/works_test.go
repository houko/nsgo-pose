package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestWorksAreNumberedAndKeepTheirLook(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "works")
	if ws, err := listWorks(dir); err != nil || len(ws) != 0 {
		t.Fatalf("a missing directory: %v, %v", ws, err)
	}
	t0 := time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC)
	pending := func(name string, data []byte) string {
		p, err := writePicture(filepath.Join(root, "results"), data, t0.Add(time.Duration(len(name))))
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	coat := &Item{Kind: KindCostume, ID: "WIN-0001", Group: "冬", Title: "camel wool coat"}
	first, err := saveWork(dir, pending("a", tinyPNG(t)), t0, workMeta{Costume: coat})
	if err != nil {
		t.Fatal(err)
	}
	second, err := saveWork(dir, pending("bb", []byte("\xff\xd8\xff\xe0jpeg")), t0.Add(time.Minute), workMeta{Favorite: true})
	if err != nil {
		t.Fatal(err)
	}
	if first.No != 1 || second.No != 2 || filepath.Ext(second.Path) != ".jpg" {
		t.Errorf("numbers = %d, %d; second = %s", first.No, second.No, second.Path)
	}
	if err := os.WriteFile(filepath.Join(dir, ".DS_Store"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ws, err := listWorks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 2 || ws[0].Path != second.Path || !ws[0].Favorite || ws[1].Costume == nil || *ws[1].Costume != *coat {
		t.Fatalf("works = %+v", ws)
	}
	if left, _ := os.ReadDir(filepath.Join(root, "results")); len(left) != 0 {
		t.Errorf("saving left %d pictures among the results", len(left))
	}

	ws[1].Favorite = true
	if err := writeMeta(ws[1]); err != nil {
		t.Fatal(err)
	}
	if err := deleteWork(ws[0]); err != nil {
		t.Fatal(err)
	}
	ws, _ = listWorks(dir)
	if len(ws) != 1 || ws[0].No != 1 || !ws[0].Favorite {
		t.Errorf("after starring the first and deleting the second: %+v", ws)
	}
	if _, err := os.Stat(metaPath(second.Path)); !os.IsNotExist(err) {
		t.Errorf("the deleted work's sidecar is still there: %v", err)
	}
}

func TestUploadsAreNumberedOldestFirst(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC)
	for _, at := range []time.Time{t0.Add(time.Hour), t0} {
		if _, err := saveUpload(dir, tinyPNG(t), at); err != nil {
			t.Fatal(err)
		}
	}
	ups, err := listUploads(dir)
	if err != nil || len(ups) != 2 {
		t.Fatalf("uploads = %v, %v", ups, err)
	}
	if ups[0].ID != "MY-01" || ups[1].ID != "MY-02" || ups[0].Group != uploadGroup || !filepath.IsAbs(ups[0].Full) {
		t.Errorf("uploads = %+v", ups)
	}
	if filepath.Base(ups[0].Full) != filepath.Base(writeName(t0)) {
		t.Errorf("oldest upload = %s", ups[0].Full)
	}
}

func writeName(at time.Time) string {
	return filepath.Join("x", strconv.FormatInt(at.UnixNano(), 10)+".png")
}
