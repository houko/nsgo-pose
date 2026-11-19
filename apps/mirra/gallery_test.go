package main

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"
)

func ids(items []Item) []string {
	var out []string
	for _, it := range items {
		out = append(out, it.ID)
	}
	return out
}

func TestFetchCatalogs(t *testing.T) {
	_, sites := newServer(t, &fakeAPI{})
	ctx := context.Background()

	poses, err := FetchCatalog(ctx, http.DefaultClient, sites, KindPose)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ids(poses.Items), []string{"full_std_P001_front", "upper_P002_side", "full_sexy_P000_front"}; !slices.Equal(got, want) {
		t.Errorf("pose ids = %v, want %v (everyday poses first; two-person and explicit poses left out)", got, want)
	}
	if want := []string{"正面", "横"}; !slices.Equal(poses.Groups, want) {
		t.Errorf("pose groups = %v, want the angles %v", poses.Groups, want)
	}
	if got := ids(poses.InGroup("正面")); !slices.Equal(got, []string{"full_std_P001_front", "full_sexy_P000_front"}) {
		t.Errorf("front poses = %v", got)
	}
	first := poses.Items[0]
	if first.Full != sites.Pose+"/img/full_std_P001_front.png" || first.Thumb != sites.Pose+"/thumbs/full_std_P001_front.jpg" || first.Title != "片手を腰に当てて立つ" {
		t.Errorf("first pose = %+v", first)
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

	faces, err := FetchCatalog(ctx, http.DefaultClient, sites, KindFace)
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(faces.Items); !slices.Equal(got, []string{"f002", "f001"}) {
		t.Errorf("faces = %v, want the face section only", got)
	}
	if want := []string{"ショート", "ロング"}; !slices.Equal(faces.Groups, want) {
		t.Errorf("face groups = %v, want hair lengths shortest first %v", faces.Groups, want)
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
