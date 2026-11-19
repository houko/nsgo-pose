package main

import (
	"slices"
	"testing"
	"time"

	"github.com/nsgo-dev/nsgo/nsgotest"
)

// TestMain gives the platform the main thread under nsgo test -target.
func TestMain(m *testing.M) { nsgotest.Main(m) }

const wait = 5 * time.Second

// start runs the app against local galleries and a local image API, keeping its works in a temporary directory.
func start(t *testing.T, api *fakeAPI, key string, opts ...nsgotest.Option) (*nsgotest.App, string) {
	t.Helper()
	srv, sites := newServer(t, api)
	dir := t.TempDir()
	secrets := map[string]string{}
	if key != "" {
		secrets[apiKeyAccount] = key
	}
	opts = append([]nsgotest.Option{
		nsgotest.Storage(map[string]string{"baseURL": srv.URL + "/v1", "model": "test-model"}),
		nsgotest.Secrets(secrets),
	}, opts...)
	app := nsgotest.Start(t, App{Client: srv.Client(), Sites: sites, WorksDir: dir}.Root, opts...)
	return app, dir
}

// choose presses a gallery cell until it reports itself chosen. On the macOS target a press while a pushed page is still moving in is lost, and neither Settle nor the view tree says when the move ends, so there the press first waits out the transition: https://github.com/nsgo-dev/nsgo/issues/854
func choose(t *testing.T, app *nsgotest.App, selector string) {
	t.Helper()
	app.WaitFor(selector, wait)
	if nsgotest.Target() != "" {
		time.Sleep(time.Second)
	}
	app.Press(selector)
	app.WaitUntil(wait, func() bool { return app.Find(selector).Props.Selected })
}

// The cells' labels are the galleries' own words, the same in every language; a face without a hairstyle is labeled by its hair length and ID, so its selector matches the ID alone.
const (
	faceShort   = "label:/^[^,]+ f001$/"
	outfitCoat  = "label:camel wool coat"
	poseStand   = "label:片手を腰に当てて立つ"
	poseTurning = "label:振り返る"
)

func TestPhoneLookStepByStep(t *testing.T) {
	api := &fakeAPI{}
	app, dir := start(t, api, "sk-test")

	if app.Enabled("id:generate") {
		t.Error("Generate is enabled with nothing chosen")
	}
	app.Find("id:hint")

	app.Press("id:row-face")
	choose(t, app, faceShort)
	app.Press("id:next")
	choose(t, app, outfitCoat)
	app.Press("id:next")
	if app.Exists(poseTurning) {
		t.Error("the pose step shows a side view while the front chip is on")
	}
	choose(t, app, poseStand)
	app.Press("id:generate")
	app.WaitFor("id:work-image", wait)

	req := api.last()
	if req.Auth != "Bearer sk-test" || req.Model != "test-model" || req.N != "1" {
		t.Errorf("request = %+v", req)
	}
	if want := []string{"full_std_P001_front.png", "WIN-0001.png", "f001.png"}; !slices.Equal(req.Images, want) {
		t.Errorf("images = %v, want %v", req.Images, want)
	}
	ws, err := listWorks(dir)
	if err != nil || len(ws) != 1 {
		t.Fatalf("works = %v, %v", ws, err)
	}
}

func TestDoneKeepsTheLook(t *testing.T) {
	app, _ := start(t, &fakeAPI{}, "k")
	app.Press("id:row-pose")
	choose(t, app, poseStand)
	app.Press("id:done")
	app.WaitForGone("id:steps", wait)
	if !app.Enabled("id:generate") {
		t.Error("Generate is off with a pose chosen")
	}
	if app.Exists("id:hint") {
		t.Error("the hint still shows with a pose chosen")
	}
}

func TestAngleChipsFilterPoses(t *testing.T) {
	app, _ := start(t, &fakeAPI{}, "k")
	app.Press("id:row-pose")
	app.WaitFor(poseStand, wait)
	app.Press("label:/^(Side|侧面)$/")
	app.WaitFor(poseTurning, wait)
	if app.Exists(poseStand) {
		t.Error("a front pose shows under the side chip")
	}
}

func TestSeveralPictures(t *testing.T) {
	api := &fakeAPI{}
	app, dir := start(t, api, "k")
	app.Press("id:count-2")
	app.Press("id:row-pose")
	choose(t, app, poseStand)
	app.Press("id:generate")
	app.WaitFor("id:work-image", wait)
	if got := api.last().N; got != "2" {
		t.Errorf("n = %s, want 2", got)
	}
	if ws, _ := listWorks(dir); len(ws) != 2 {
		t.Errorf("works = %d, want 2", len(ws))
	}
}

func TestCancelGenerating(t *testing.T) {
	api := &fakeAPI{hang: true}
	app, dir := start(t, api, "k")
	app.Press("id:row-pose")
	choose(t, app, poseStand)
	app.Press("id:generate")
	app.WaitFor("id:generating", wait)
	app.Press("id:cancel")
	app.WaitForGone("id:generating", wait)
	if app.Exists("role:alert") {
		t.Error("cancelling showed an error")
	}
	if ws, _ := listWorks(dir); len(ws) != 0 {
		t.Errorf("a cancelled look left %d works", len(ws))
	}
	if !app.Enabled("id:generate") {
		t.Error("Generate stays off after cancelling")
	}
}

func TestGeneratingWithoutAKeyOpensSettings(t *testing.T) {
	api := &fakeAPI{}
	app, _ := start(t, api, "")
	app.Press("id:row-pose")
	choose(t, app, poseStand)
	app.Press("id:generate")
	app.WaitFor("id:api-key", wait)
	if len(api.requests) != 0 {
		t.Error("the app called the image API without a key")
	}
}

func TestGalleryErrorOffersRetry(t *testing.T) {
	app := nsgotest.Start(t, App{Sites: Sites{Pose: "http://127.0.0.1:1", Costume: "http://127.0.0.1:1", Face: "http://127.0.0.1:1"}, WorksDir: t.TempDir()}.Root)
	app.Press("id:row-costume")
	app.WaitFor("id:retry", wait)
}

func TestDesktopStudio(t *testing.T) {
	api := &fakeAPI{}
	app, dir := start(t, api, "k", nsgotest.Size(1440, 900))
	app.Find("id:desktop")
	choose(t, app, faceShort)
	app.Press("id:library-costume")
	choose(t, app, outfitCoat)
	app.Press("id:row-pose")
	choose(t, app, poseStand)
	app.Press("id:generate")
	app.WaitFor("id:result", wait)
	if want := []string{"full_std_P001_front.png", "WIN-0001.png", "f001.png"}; !slices.Equal(api.last().Images, want) {
		t.Errorf("images = %v, want %v", api.last().Images, want)
	}
	if ws, _ := listWorks(dir); len(ws) != 1 {
		t.Errorf("works = %d, want 1", len(ws))
	}
	app.Press("id:close-result")
	app.WaitFor("id:preview", wait)
}
