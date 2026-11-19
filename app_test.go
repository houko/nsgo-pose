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
func start(t *testing.T, api *fakeAPI, key string) (*nsgotest.App, string) {
	t.Helper()
	srv, sites := newServer(t, api)
	dir := t.TempDir()
	secrets := map[string]string{}
	if key != "" {
		secrets[apiKeyAccount] = key
	}
	app := nsgotest.Start(t, App{Client: srv.Client(), Sites: sites, WorksDir: dir}.Root,
		nsgotest.Storage(map[string]string{"baseURL": srv.URL + "/v1", "model": "test-model"}),
		nsgotest.Secrets(secrets))
	return app, dir
}

func TestPoseOutfitAndFaceMakeAPicture(t *testing.T) {
	api := &fakeAPI{}
	app, dir := start(t, api, "sk-test")

	if app.Enabled("id:generate") {
		t.Error("Generate is enabled with nothing chosen")
	}
	app.Find("id:hint")

	app.Press("id:slot-pose")
	app.WaitFor("label:片手を腰に当てて立つ・正面", wait)
	if app.Exists("label:x・正面") {
		t.Error("the picker offers an excluded pose")
	}
	pick(t, app, "label:片手を腰に当てて立つ・正面")
	app.WaitFor("id:slot-costume", wait)

	app.Press("id:slot-costume")
	app.WaitFor("label:camel wool coat", wait)
	pick(t, app, "label:camel wool coat")

	app.Press("id:slot-face")
	app.WaitFor("label:ショート", wait)
	pick(t, app, "label:ショート")

	app.WaitFor("id:slot-face-clear", wait)
	if app.Exists("id:hint") {
		t.Error("the hint still shows with a pose chosen")
	}
	app.Press("id:generate")
	app.WaitFor("id:work-image", wait)

	req := api.last()
	if req.Auth != "Bearer sk-test" || req.Model != "test-model" {
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

func TestClearingAChoice(t *testing.T) {
	app, _ := start(t, &fakeAPI{}, "k")
	app.Press("id:slot-face")
	app.WaitFor("label:ショート", wait)
	pick(t, app, "label:ショート")
	app.WaitFor("id:slot-face-clear", wait)
	app.Press("id:slot-face-clear")
	app.WaitForGone("id:slot-face-clear", wait)
}

func TestPhotoTakesOnlyAFace(t *testing.T) {
	app, _ := start(t, &fakeAPI{}, "k")
	app.Press("id:mode-photo")
	app.Find("id:slot-photo")
	app.Find("id:slot-face")
	if app.Exists("id:slot-pose") {
		t.Error("the pose slot shows in photo mode")
	}
	if app.Exists("id:slot-costume") {
		t.Error("the outfit slot shows in photo mode; the outfit gallery's terms forbid putting its outfits on real people")
	}
	if app.Enabled("id:generate") {
		t.Error("Generate is enabled without a photo")
	}
	app.Find("id:hint")
}

func TestGeneratingWithoutAKeyOpensSettings(t *testing.T) {
	api := &fakeAPI{}
	app, _ := start(t, api, "")
	app.Press("id:slot-pose")
	app.WaitFor("label:振り返る・横", wait)
	pick(t, app, "label:振り返る・横")
	app.WaitFor("id:generate", wait)
	app.Press("id:generate")
	app.WaitFor("id:api-key", wait)
	if len(api.requests) != 0 {
		t.Error("the app called the image API without a key")
	}
}

func TestGalleryErrorOffersRetry(t *testing.T) {
	app := nsgotest.Start(t, App{Sites: Sites{Pose: "http://127.0.0.1:1", Costume: "http://127.0.0.1:1", Face: "http://127.0.0.1:1"}, WorksDir: t.TempDir()}.Root)
	app.Press("id:slot-costume")
	app.WaitFor("id:retry", wait)
}

func TestAnOutfitStaysWithThePose(t *testing.T) {
	app, _ := start(t, &fakeAPI{}, "k")
	app.Press("id:slot-costume")
	app.WaitFor("label:camel wool coat", wait)
	pick(t, app, "label:camel wool coat")
	app.WaitFor("id:slot-costume-clear", wait)
	app.Press("id:mode-photo")
	app.WaitForGone("id:slot-costume", wait)
	app.Press("id:mode-pose")
	app.WaitFor("id:slot-costume-clear", wait)
}

// pick presses a picker's cell, which closes the picker. On the macOS target a press while a pushed page is still moving in is lost, and neither Settle nor the view tree says when the move ends, so on a target the press waits out the transition first: https://github.com/nsgo-dev/nsgo/issues/854
func pick(t *testing.T, app *nsgotest.App, selector string) {
	t.Helper()
	if nsgotest.Target() != "" {
		time.Sleep(time.Second)
	}
	app.Press(selector)
	app.WaitForGone("id:groups", wait)
}
