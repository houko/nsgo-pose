package main

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/nsgotest"
)

// TestMain gives the platform the main thread under nsgo test -target.
func TestMain(m *testing.M) { nsgotest.Main(m) }

const wait = 5 * time.Second

// start runs the app against local galleries and a local image API, keeping its data in a temporary directory; key "" leaves the cloud service without one.
func start(t *testing.T, api *fakeAPI, key string, opts ...nsgotest.Option) (*nsgotest.App, string) {
	t.Helper()
	srv, sites := newServer(t, api)
	dir := t.TempDir()
	secrets := map[string]string{}
	if key != "" {
		secrets[apiKeyAccount] = key
	}
	// The tests read English words; a target otherwise shows the machine's language.
	opts = append([]nsgotest.Option{
		nsgotest.Environment(core.Environment{Locale: "en"}),
		nsgotest.Storage(map[string]string{keyCloudURL: srv.URL + "/v1", keyLocalURL: srv.URL + "/v1", keyModel: "test-model"}),
		nsgotest.Secrets(secrets),
	}, opts...)
	app := nsgotest.Start(t, App{Client: srv.Client(), Sites: sites, DataDir: dir}.Root, opts...)
	return app, dir
}

func works(t *testing.T, dir string) []work {
	t.Helper()
	ws, err := listWorks(filePictures{}, filepath.Join(dir, "works"))
	if err != nil {
		t.Fatal(err)
	}
	return ws
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

// The cells' labels are the galleries' own words, the same in every language; a face is labeled by its hair length and ID, so its selector matches the ID alone.
const (
	faceShort   = "label:/ · f001$/"
	outfitCoat  = "label:camel wool coat"
	outfitKnit  = "label:ivory knit denim skirt"
	poseStand   = "label:片手を腰に当てて立つ"
	poseTurning = "label:振り返る"
)

// makeLook puts a face, an outfit and a pose together on a phone, and generates.
func makeLook(t *testing.T, app *nsgotest.App) {
	t.Helper()
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
	app.WaitFor("id:shot-image", wait)
}

func TestPhoneLookIsKeptOnlyWhenSaved(t *testing.T) {
	api := &fakeAPI{}
	app, dir := start(t, api, "sk-test")
	if app.Enabled("id:generate") {
		t.Error("Generate is enabled with nothing chosen")
	}
	makeLook(t, app)

	req := api.last()
	if req.Auth != "Bearer sk-test" || req.Model != "test-model" || req.N != "1" {
		t.Errorf("request = %+v", req)
	}
	if want := []string{"full_std_P001_front.png", "WIN-0001.png", "f001.png"}; !slices.Equal(req.Images, want) {
		t.Errorf("images = %v, want %v", req.Images, want)
	}
	if n := len(works(t, dir)); n != 0 {
		t.Fatalf("a picture not saved is in Works: %d", n)
	}
	app.Press("id:save-to-works")
	app.WaitFor("id:saved", wait)
	ws := works(t, dir)
	if len(ws) != 1 || ws[0].No != 1 || ws[0].Costume == nil || ws[0].Costume.ID != "WIN-0001" || ws[0].Pose == nil || ws[0].Face == nil {
		t.Fatalf("works = %+v", ws)
	}
}

func TestRegenerateAndSwapOutfit(t *testing.T) {
	api := &fakeAPI{}
	app, _ := start(t, api, "k")
	makeLook(t, app)
	app.Press("id:regenerate")
	app.WaitUntil(wait, func() bool { return len(api.requests) == 2 })
	app.WaitFor("text:/TRY 2/", wait)

	app.Press("id:swap-0")
	app.WaitUntil(wait, func() bool { return len(api.requests) == 3 })
	app.WaitFor("text:/TRY 3/", wait)
	if got := api.last().Images; len(got) != 3 || got[1] != "SPR-0001.png" {
		t.Errorf("swapped images = %v, want the other outfit", got)
	}
}

func TestFavoriteSavesAndFilters(t *testing.T) {
	app, dir := start(t, &fakeAPI{}, "k")
	makeLook(t, app)
	app.Press("id:favorite")
	app.WaitFor("id:saved", wait)
	if ws := works(t, dir); len(ws) != 1 || !ws[0].Favorite {
		t.Fatalf("works = %+v, want one favorite", ws)
	}
	app.Press("id:close-shot")
	app.Press("text:Works")
	app.Press("id:works-favorites")
	app.WaitFor("label:Open No.01", wait)
}

func TestDeleteAShot(t *testing.T) {
	app, dir := start(t, &fakeAPI{}, "k", nsgotest.PendingAlerts())
	makeLook(t, app)
	app.Press("id:save-to-works")
	app.WaitFor("id:saved", wait)
	app.Press("id:delete")
	app.Answer(1)
	app.WaitForGone("id:shot-image", wait)
	if ws := works(t, dir); len(ws) != 0 {
		t.Errorf("works after deleting = %+v", ws)
	}
}

func TestCancelGenerating(t *testing.T) {
	api := &fakeAPI{hang: true}
	app, _ := start(t, api, "k")
	app.Press("id:row-pose")
	choose(t, app, poseStand)
	app.Press("id:generate")
	app.WaitFor("id:generating", wait)
	app.Press("id:cancel")
	app.WaitForGone("id:generating", wait)
	if app.Exists("role:alert") {
		t.Error("cancelling showed an error")
	}
	if app.Exists("id:shot-image") {
		t.Error("a cancelled look showed a picture")
	}
}

func TestNoServiceOpensSettings(t *testing.T) {
	api := &fakeAPI{}
	app, _ := start(t, api, "")
	if got := app.Find("id:generate").Props.Label; !strings.Contains(got, "AI service") {
		t.Errorf("Generate says %q without a service", got)
	}
	app.Press("id:generate")
	app.WaitFor("id:settings", wait)
	if len(api.requests) != 0 {
		t.Error("the app called the image API without a key")
	}
}

func TestLocalServiceNeedsNoKey(t *testing.T) {
	api := &fakeAPI{}
	app, _ := start(t, api, "")
	app.Press("text:Me")
	app.Press("id:service-local")
	app.Press("id:test-connection")
	app.WaitFor("id:conn-ok", wait)
	app.Press("text:Studio")
	app.Press("id:row-pose")
	choose(t, app, poseStand)
	app.Press("id:generate")
	app.WaitFor("id:shot-image", wait)
	if got := api.last().Auth; got != "" {
		t.Errorf("a keyless local request sent Authorization %q", got)
	}
}

func TestConnectionTestReportsFailure(t *testing.T) {
	app, _ := start(t, &fakeAPI{}, "bad")
	app.Press("text:Me")
	app.Press("id:test-connection")
	app.WaitFor("id:conn-failed", wait)
	app.Find("text:/can't connect/")
}

func TestWardrobeFavorites(t *testing.T) {
	app, _ := start(t, &fakeAPI{}, "k")
	app.Press("text:Wardrobe")
	app.Press("id:closet-favorites")
	app.Find("text:No favorites yet")
	app.Press("id:closet-gallery")
	app.WaitFor(outfitKnit, wait)
	app.Press("label:Add to favorites >> nth=0")
	app.Press("id:closet-favorites")
	app.WaitFor(outfitKnit, wait)
}

func TestUploadedClothesAreWorn(t *testing.T) {
	if nsgotest.Target() != "" {
		t.Skip("a target never hands the test's PickImages answer to the app: https://github.com/nsgo-dev/nsgo/issues/866")
	}
	api := &fakeAPI{}
	app, _ := start(t, api, "k")
	app.Press("text:Wardrobe")
	app.PickImages = func(core.PickRequest) []core.PickedFile {
		return []core.PickedFile{{Name: "shirt.png", MIME: "image/png", Data: tinyPNG(t)}}
	}
	app.Press("id:upload")
	app.WaitFor("label:My clothes 01", wait)
	app.WaitFor("id:wearing", wait)

	app.Press("id:wear-generate")
	app.WaitFor("id:steps", wait)
	choose(t, app, poseStand)
	app.Press("id:generate")
	app.WaitFor("id:shot-image", wait)
	if got := api.last().Images; len(got) != 2 || got[1] == "WIN-0001.png" || !strings.HasSuffix(got[1], ".png") {
		t.Errorf("images = %v, want the pose and the uploaded clothes", got)
	}
}

func TestGalleryErrorOffersRetry(t *testing.T) {
	app := nsgotest.Start(t, App{Sites: Sites{Pose: "http://127.0.0.1:1", Costume: "http://127.0.0.1:1", Face: "http://127.0.0.1:1"}, DataDir: t.TempDir()}.Root)
	app.Press("id:row-costume")
	app.WaitFor("id:retry", wait)
}

func TestDesktopStudio(t *testing.T) {
	api := &fakeAPI{}
	app, dir := start(t, api, "k", nsgotest.Size(1440, 1000))
	app.Find("id:desktop")
	choose(t, app, faceShort)
	app.Press("id:library-costume")
	choose(t, app, outfitCoat)
	app.Press("id:quick-pose")
	choose(t, app, poseStand)
	app.Press("id:zoom-in")
	app.Find("text:125%")
	app.Press("id:generate")
	app.WaitFor("id:shot-image", wait)
	if want := []string{"full_std_P001_front.png", "WIN-0001.png", "f001.png"}; !slices.Equal(api.last().Images, want) {
		t.Errorf("images = %v, want %v", api.last().Images, want)
	}
	app.Press("id:save-to-works")
	app.WaitFor("id:saved", wait)
	if ws := works(t, dir); len(ws) != 1 {
		t.Errorf("works = %d, want 1", len(ws))
	}
	app.Press("id:close-shot")
	app.WaitFor("id:zoom", wait)
	app.Press("id:open-settings")
	app.WaitFor("id:settings", wait)
}

// TestPicturesInMemory runs the look as the web version does, with nothing written to disk: the picture is saved to Works in memory and the page says it lasts only as long as the page.
func TestPicturesInMemory(t *testing.T) {
	api := &fakeAPI{}
	srv, sites := newServer(t, api)
	mem := newMemPictures()
	app := nsgotest.Start(t, App{Client: srv.Client(), Sites: sites, DataDir: "/web", Pictures: mem}.Root,
		nsgotest.Environment(core.Environment{Locale: "en"}),
		nsgotest.Storage(map[string]string{keyCloudURL: srv.URL + "/v1"}),
		nsgotest.Secrets(map[string]string{apiKeyAccount: "k"}))
	app.Press("id:row-pose")
	choose(t, app, poseStand)
	app.Press("id:generate")
	app.WaitFor("id:shot-image", wait)
	if got := app.Find("id:shot-image").Image.Source; len(got.Data) == 0 {
		t.Errorf("the result shows from %q, not the memory it is kept in", got.URI)
	}
	app.Press("id:save-to-works")
	app.WaitFor("id:saved", wait)
	ws, err := listWorks(mem, "/web/works")
	if err != nil || len(ws) != 1 {
		t.Fatalf("works in memory = %v, %v", ws, err)
	}
	app.Press("id:close-shot")
	app.Press("text:Works")
	app.WaitFor("id:page-only", wait)
}
