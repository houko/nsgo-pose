package main

import (
	"bytes"
	"context"
	"errors"
	"image/png"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	if err := (Request{Face: &Item{}, Costume: &Item{}}).Validate(); !errors.Is(err, ErrNoPose) {
		t.Errorf("without a pose: %v", err)
	}
	if err := (Request{Pose: &Item{}}).Validate(); err != nil {
		t.Errorf("a pose alone: %v", err)
	}
}

func TestPromptNumbersImagesInOrder(t *testing.T) {
	p := Request{Pose: &Item{}, Face: &Item{}, Note: "  golden hour  "}.Prompt()
	if !hasAll(p, "Image 1 is a faceless gray mannequin", "Image 2 is a portrait", "plain everyday clothes", "Also: golden hour") || strings.Contains(p, "Image 3") {
		t.Errorf("pose + face prompt = %q", p)
	}
	p = Request{Pose: &Item{}, Costume: &Item{}, Face: &Item{}}.Prompt()
	if !hasAll(p, "Image 2 shows an outfit", "Image 3 is a portrait") || strings.Contains(p, "everyday clothes") || strings.Contains(p, "natural adult face") {
		t.Errorf("pose + outfit + face prompt = %q", p)
	}
}

func TestTwoPersonPosesAskForTwoPeople(t *testing.T) {
	pair := &Item{People: 2}
	p := Request{Pose: pair, Costume: &Item{}, Face: &Item{}, Note: "laughing"}.Prompt()
	if !hasAll(p, "two faceless gray mannequins", "two adult people", "Image 2 shows an outfit", "dress the main person", "Image 3 is a portrait", "give the main person that face", "Also: laughing", "these two adults") || strings.Contains(p, "single adult person") || strings.Contains(p, "one adult person") {
		t.Errorf("two-person prompt = %q", p)
	}
	p = Request{Pose: pair}.Prompt()
	if !hasAll(p, "Dress both", "Give both natural, different adult faces") || strings.Contains(p, "Image 2") {
		t.Errorf("two-person pose alone = %q", p)
	}
	if p := (Request{Pose: &Item{People: 1}}).Prompt(); !strings.Contains(p, "single adult person") {
		t.Errorf("one-person pose = %q", p)
	}
}

func TestGenerateSendsReferencesInOrder(t *testing.T) {
	api := &fakeAPI{}
	srv, sites := newServer(t, api)
	ctx := context.Background()
	poses, _ := FetchCatalog(ctx, http.DefaultClient, sites, KindPose)
	costumes, _ := FetchCatalog(ctx, http.DefaultClient, sites, KindCostume)
	faces, _ := FetchCatalog(ctx, http.DefaultClient, sites, KindFace)
	winter := costumes.InGroup("冬", false)[0]

	cfg := Config{BaseURL: srv.URL + "/v1/", APIKey: " sk-test ", Size: "1024x1024"}
	got, err := Generate(ctx, http.DefaultClient, cfg, Request{Pose: &poses.Items[0], Costume: &winter, Face: &faces.Items[1], Count: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !bytes.Equal(got[0], tinyPNG(t)) || !bytes.Equal(got[1], tinyPNG(t)) {
		t.Errorf("Generate returned %d pictures, want the API's 2 decoded", len(got))
	}
	req := api.last()
	if req.Auth != "Bearer sk-test" || req.Model != defaultModel || req.Size != "1024x1024" || req.N != "2" {
		t.Errorf("request = %+v", req)
	}
	if want := []string{"full_std_P001_front.png", "WIN-0001.png", "f001.png"}; !slices.Equal(req.Images, want) {
		t.Errorf("images = %v, want pose, outfit, face: %v", req.Images, want)
	}
	// Only the three chosen images are downloaded, one each.
	if want := []string{"/pose/img/full_std_P001_front.png", "/costume/png/WIN-0001.png", "/face/img/full/f001.png"}; !slices.Equal(req.Downloads, want) {
		t.Errorf("downloads = %v, want %v", req.Downloads, want)
	}
}

func TestGenerateBoundsCount(t *testing.T) {
	api := &fakeAPI{}
	srv, _ := newServer(t, api)
	pose := &Item{Full: srv.URL + "/pose/img/a.png"}
	for _, tc := range []struct {
		count int
		want  string
	}{{0, "1"}, {maxCount + 5, "4"}} {
		if _, err := Generate(context.Background(), http.DefaultClient, Config{BaseURL: srv.URL + "/v1", APIKey: "k"}, Request{Pose: pose, Count: tc.count}); err != nil {
			t.Fatal(err)
		}
		if got := api.last().N; got != tc.want {
			t.Errorf("Count %d sent n = %s, want %s", tc.count, got, tc.want)
		}
	}
}

func TestGenerateErrors(t *testing.T) {
	api := &fakeAPI{status: http.StatusBadRequest, errBody: `{"error":{"message":"Your request was rejected by the safety system."}}`}
	srv, _ := newServer(t, api)
	pose := &Item{Full: srv.URL + "/pose/img/a.png"}
	_, err := Generate(context.Background(), http.DefaultClient, Config{BaseURL: srv.URL + "/v1", APIKey: "k"}, Request{Pose: pose})
	if err == nil || !strings.Contains(err.Error(), "safety system") {
		t.Errorf("err = %v, want the API's message", err)
	}
	if _, err := Generate(context.Background(), http.DefaultClient, Config{}, Request{Pose: pose}); !errors.Is(err, ErrNoAPIKey) {
		t.Errorf("without a key: err = %v", err)
	}
	if _, err := Generate(context.Background(), http.DefaultClient, Config{APIKey: "k"}, Request{}); !errors.Is(err, ErrNoPose) {
		t.Errorf("without a pose: err = %v", err)
	}
}

func TestGenerateLocalWithoutKey(t *testing.T) {
	api := &fakeAPI{}
	srv, _ := newServer(t, api)
	pose := &Item{Full: srv.URL + "/pose/img/a.png"}
	if _, err := Generate(context.Background(), http.DefaultClient, Config{BaseURL: srv.URL + "/v1", Local: true}, Request{Pose: pose}); err != nil {
		t.Fatal(err)
	}
	if got := api.last().Auth; got != "" {
		t.Errorf("a keyless local request sent Authorization %q", got)
	}
}

func TestGenerateReadsUploadedClothes(t *testing.T) {
	api := &fakeAPI{}
	srv, _ := newServer(t, api)
	up, err := saveUpload(filePictures{}, t.TempDir(), tinyPNG(t), time.Unix(0, 42))
	if err != nil {
		t.Fatal(err)
	}
	pose := &Item{Full: srv.URL + "/pose/img/a.png"}
	if _, err := Generate(context.Background(), http.DefaultClient, Config{BaseURL: srv.URL + "/v1", APIKey: "k"},
		Request{Pose: pose, Costume: &Item{Kind: KindCostume, ID: "MY-01", Group: uploadGroup, Full: up}}); err != nil {
		t.Fatal(err)
	}
	if got := api.last().Images; !slices.Equal(got, []string{"a.png", "42.png"}) {
		t.Errorf("images = %v", got)
	}
}

func TestCheckConnection(t *testing.T) {
	srv, _ := newServer(t, &fakeAPI{})
	if err := CheckConnection(context.Background(), http.DefaultClient, Config{BaseURL: srv.URL + "/v1/", APIKey: "k"}); err != nil {
		t.Errorf("a good key: %v", err)
	}
	err := CheckConnection(context.Background(), http.DefaultClient, Config{BaseURL: srv.URL + "/v1", APIKey: "bad"})
	if err == nil || !strings.Contains(err.Error(), "Incorrect API key") {
		t.Errorf("a bad key: %v", err)
	}
	if err := CheckConnection(context.Background(), http.DefaultClient, Config{BaseURL: srv.URL + "/nowhere"}); err == nil {
		t.Error("a wrong address passed")
	}
}

func TestSheetSendsTheAnglesAsOneImage(t *testing.T) {
	api := &fakeAPI{}
	srv, sites := newServer(t, api)
	ctx := context.Background()
	poses, _ := FetchCatalog(ctx, http.DefaultClient, sites, KindPose)
	faces, _ := FetchCatalog(ctx, http.DefaultClient, sites, KindFace)
	views := poses.Angles(poses.Items[0])
	r := Request{Pose: &views[0], Views: []*Item{&views[0], &views[1]}, Face: &faces.Items[1]}
	if _, err := Generate(ctx, http.DefaultClient, Config{BaseURL: srv.URL + "/v1", APIKey: "k", Size: "1024x1536"}, r); err != nil {
		t.Fatal(err)
	}
	req := api.last()
	if want := []string{"angles.png", "f001.png"}; !slices.Equal(req.Images, want) {
		t.Errorf("images = %v, want %v", req.Images, want)
	}
	if req.Size != "1536x1024" {
		t.Errorf("size = %q, want the portrait size on its side", req.Size)
	}
	if !hasAll(req.Prompt, "from 2 camera angles in 2 panels", "Image 2 is a portrait", "2 evenly spaced panels") || strings.Contains(req.Prompt, "Image 3") {
		t.Errorf("prompt = %q", req.Prompt)
	}
	if want := []string{"/pose/img/full_std_P001_front.png", "/pose/img/full_std_P001_side.png", "/face/img/full/f001.png"}; !slices.Equal(req.Downloads, want) {
		t.Errorf("downloads = %v, want %v", req.Downloads, want)
	}

	if p := (Request{Pose: &views[0], Views: []*Item{&views[0], &views[1]}, Costume: &Item{}}).Prompt(); !hasAll(p, "Image 2 shows an outfit", "natural adult face") {
		t.Errorf("sheet with an outfit = %q", p)
	}
	pair := &Item{People: 2}
	if p := (Request{Pose: pair, Views: []*Item{pair, pair}}).Prompt(); strings.Contains(p, "panels") {
		t.Errorf("a two-person pose asked for a sheet: %q", p)
	}
	for in, want := range map[string]string{"1024x1536": "1536x1024", "1024*1536": "1536x1024", "1536x1024": "1536x1024", "1024x1024": "1024x1024", "2K": "2K"} {
		if got := landscape(in); got != want {
			t.Errorf("landscape(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSideBySide(t *testing.T) {
	pic := part{name: "a.png", data: tinyPNG(t)}
	out, err := sideBySide([]part{pic, pic})
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out.data))
	if err != nil {
		t.Fatal(err)
	}
	// Two 2×3 images scaled to 768 high are 512 wide each, with a 24-pixel gap around them.
	if b := img.Bounds(); b.Dx() != 24+512+24+512+24 || b.Dy() != 768+48 {
		t.Errorf("sheet is %v", b)
	}
	if _, err := sideBySide([]part{{name: "bad", data: []byte("not an image")}}); err == nil {
		t.Error("an unreadable reference made a sheet")
	}
}
