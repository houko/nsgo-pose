package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	pose := &Item{ID: "p"}
	outfit := &Item{ID: "c"}
	for _, tc := range []struct {
		name string
		r    Request
		want error
	}{
		{"empty", Request{}, ErrNoBase},
		{"outfit without base", Request{Costume: outfit}, ErrNoBase},
		{"pose alone", Request{Base: Source{Item: pose}}, nil},
		{"photo alone", Request{Base: Source{Photo: []byte{1}}}, ErrNothingToDo},
		{"photo with face", Request{Base: Source{Photo: []byte{1}}, Face: &Item{ID: "f"}}, nil},
		{"photo with outfit", Request{Base: Source{Photo: []byte{1}}, Costume: outfit}, ErrPhotoOutfit},
		{"photo with outfit and face", Request{Base: Source{Photo: []byte{1}}, Costume: outfit, Face: &Item{ID: "f"}}, ErrPhotoOutfit},
	} {
		if got := tc.r.Validate(); got != tc.want {
			t.Errorf("%s: Validate() = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestPromptNumbersImagesInOrder(t *testing.T) {
	r := Request{Base: Source{Item: &Item{}}, Face: &Item{}, Note: "  golden hour  "}
	p := r.Prompt()
	if !hasAll(p, "Image 1 is a faceless gray mannequin", "Image 2 is a portrait", "plain everyday clothes", "Also: golden hour") {
		t.Errorf("pose + face prompt = %q", p)
	}
	r = Request{Base: Source{Photo: []byte{1}}, Face: &Item{}}
	p = r.Prompt()
	if !hasAll(p, "Image 1 is a photo of a person", "Image 2 is a portrait: replace the person's face") || strings.Contains(p, "Image 3") || strings.Contains(p, "outfit") {
		t.Errorf("photo + face prompt = %q", p)
	}
}

func TestGenerateSendsReferencesInOrder(t *testing.T) {
	api := &fakeAPI{}
	srv, sites := newServer(t, api)
	ctx := context.Background()
	poses, _ := FetchCatalog(ctx, http.DefaultClient, sites, KindPose)
	costumes, _ := FetchCatalog(ctx, http.DefaultClient, sites, KindCostume)
	faces, _ := FetchCatalog(ctx, http.DefaultClient, sites, KindFace)

	cfg := Config{BaseURL: srv.URL + "/v1/", APIKey: " sk-test ", Model: "", Size: "1024x1024"}
	got, err := Generate(ctx, http.DefaultClient, cfg, Request{Base: Source{Item: &poses.Items[0]}, Costume: &costumes.Items[1], Face: &faces.Items[0]})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, tinyPNG(t)) {
		t.Error("Generate did not return the API's decoded image")
	}
	req := api.last()
	if req.Auth != "Bearer sk-test" || req.Model != defaultModel || req.Size != "1024x1024" {
		t.Errorf("request = %+v", req)
	}
	if want := []string{"full_std_P001_front.png", "WIN-0001.png", "f001.png"}; !slices.Equal(req.Images, want) {
		t.Errorf("images = %v, want %v", req.Images, want)
	}
	// Only the three chosen images are downloaded, one each.
	if want := []string{"/pose/img/full_std_P001_front.png", "/costume/png/WIN-0001.png", "/face/img/full/f001.png"}; !slices.Equal(req.Downloads, want) {
		t.Errorf("downloads = %v, want %v", req.Downloads, want)
	}
}

func TestGenerateUsesPickedPhoto(t *testing.T) {
	api := &fakeAPI{}
	srv, sites := newServer(t, api)
	faces, _ := FetchCatalog(context.Background(), http.DefaultClient, sites, KindFace)
	_, err := Generate(context.Background(), http.DefaultClient, Config{BaseURL: srv.URL + "/v1", APIKey: "k"},
		Request{Base: Source{Photo: tinyPNG(t), Name: "IMG_0042.png", MIME: "image/png"}, Face: &faces.Items[0]})
	if err != nil {
		t.Fatal(err)
	}
	if got := api.last().Images; !slices.Equal(got, []string{"IMG_0042.png", "f001.png"}) {
		t.Errorf("images = %v", got)
	}
}

func TestGenerateErrors(t *testing.T) {
	api := &fakeAPI{status: http.StatusBadRequest, errBody: `{"error":{"message":"Your request was rejected by the safety system."}}`}
	srv, _ := newServer(t, api)
	pose := &Item{Full: srv.URL + "/pose/img/a.png"}
	_, err := Generate(context.Background(), http.DefaultClient, Config{BaseURL: srv.URL + "/v1", APIKey: "k"}, Request{Base: Source{Item: pose}})
	if err == nil || !strings.Contains(err.Error(), "safety system") {
		t.Errorf("err = %v, want the API's message", err)
	}
	if _, err := Generate(context.Background(), http.DefaultClient, Config{}, Request{Base: Source{Item: pose}}); !errors.Is(err, ErrNoAPIKey) {
		t.Errorf("without a key: err = %v", err)
	}
	if _, err := Generate(context.Background(), http.DefaultClient, Config{APIKey: "k"}, Request{}); !errors.Is(err, ErrNoBase) {
		t.Errorf("without a base: err = %v", err)
	}
}
