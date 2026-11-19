package main

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestFetchCatalogs(t *testing.T) {
	_, sites := newServer(t, &fakeAPI{})
	ctx := context.Background()

	poses, err := FetchCatalog(ctx, http.DefaultClient, sites, KindPose)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, it := range poses.Items {
		ids = append(ids, it.ID)
	}
	if want := []string{"full_std_P001_front", "upper_P002_side", "full_sexy_P000_front"}; !slices.Equal(ids, want) {
		t.Errorf("pose ids = %v, want %v (everyday poses first; two-person and explicit poses left out)", ids, want)
	}
	if want := []string{"全身1人・標準系", "上半身1人", "全身1人・セクシー系"}; !slices.Equal(poses.Groups, want) {
		t.Errorf("pose groups = %v, want %v", poses.Groups, want)
	}
	first := poses.Items[0]
	if first.Full != sites.Pose+"/img/full_std_P001_front.png" || first.Thumb != sites.Pose+"/thumbs/full_std_P001_front.jpg" {
		t.Errorf("pose urls = %q, %q", first.Thumb, first.Full)
	}
	if first.Title != "片手を腰に当てて立つ・正面" {
		t.Errorf("pose title = %q", first.Title)
	}

	costumes, err := FetchCatalog(ctx, http.DefaultClient, sites, KindCostume)
	if err != nil {
		t.Fatal(err)
	}
	if len(costumes.Items) != 2 || slices.Contains(costumes.Groups, "ランジェリー") {
		t.Errorf("costumes = %+v, want lingerie left out", costumes)
	}
	if got := costumes.InGroup("冬"); len(got) != 1 || got[0].ID != "WIN-0001" || got[0].Full != sites.Costume+"/png/WIN-0001.png" {
		t.Errorf("winter costumes = %+v", got)
	}
	if got := costumes.InGroup(""); len(got) != 2 {
		t.Errorf("all costumes = %d items", len(got))
	}

	faces, err := FetchCatalog(ctx, http.DefaultClient, sites, KindFace)
	if err != nil {
		t.Fatal(err)
	}
	if len(faces.Items) != 2 || faces.Items[1].Title != "ロング・ゆる巻き" || faces.Items[0].Full != sites.Face+"/img/full/f001.png" {
		t.Errorf("faces = %+v", faces.Items)
	}
}

func TestParseRejectsChangedFormat(t *testing.T) {
	for name, parse := range map[string]func([]byte, string) (Catalog, error){
		"pose": parsePoses, "costume": parseCostumes, "face": parseFaces,
	} {
		if _, err := parse([]byte("<html>redesigned</html>"), "https://x"); err == nil || !strings.Contains(err.Error(), "format changed") {
			t.Errorf("%s: err = %v, want a format error", name, err)
		}
	}
	if _, err := parsePoses([]byte(`const D=[{"id":"a","系統":"2人絡み","原寸":"a.png"}];`), "https://x"); err == nil {
		t.Error("a page with only excluded poses parsed as a usable catalog")
	}
}
