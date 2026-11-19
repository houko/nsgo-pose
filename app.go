// Command nsgo-pose is Mirra (镜衣): it puts a pose, an outfit and a face from three public reference galleries together and has an OpenAI-compatible image model draw the result.
package main

import (
	"context"
	_ "embed"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/houko/nsgo-pose/locales"
	"github.com/nsgo-dev/nsgo"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/ui"
)

func main() {
	app := App{Client: &http.Client{Timeout: 5 * time.Minute}, Sites: DefaultSites}
	nsgo.Run(app.Root, nsgo.WindowSize(1280, 820))
}

//go:embed theme.yaml
var themeYAML []byte

// App is the app with what it reaches outside: tests give it local servers and a temporary directory.
type App struct {
	Client *http.Client
	Sites  Sites
	// WorksDir keeps generated pictures; the app's data directory when empty.
	WorksDir string
}

// wideWidth is the window width, in points, from which the app shows the desktop studio's three columns instead of the phone's tabs.
const wideWidth = 900

// Sections of the app, in the tab bar's order.
const (
	sectionStudio = iota
	sectionWardrobe
	sectionWorks
	sectionSettings
)

// steps is the order the look is put together in.
var steps = []Kind{KindFace, KindCostume, KindPose}

type page int

const (
	pageHome page = iota
	pageEditor
	pageWork
)

type route struct {
	page page
	work work
}

const apiKeyAccount = "api-key"

// studio is the app's state, shared by the phone and desktop layouts.
type studio struct {
	App
	cx      *core.Scope
	toaster *ui.Toaster
	section *core.Signal[int]
	// step is the gallery being chosen from: the phone editor's step and the desktop library's tab.
	step *core.Signal[Kind]
	// chosen holds the look's face, outfit and pose, by Kind.
	chosen [3]*core.Signal[*Item]
	// filter is each gallery's chip, "" for all.
	filter [3]*core.Signal[string]
	note   *core.Signal[string]
	count  *core.Signal[int]

	busy   *core.Signal[bool]
	cancel context.CancelFunc
	// shown is the desktop canvas's last result, shown until the look changes.
	shown *core.Signal[*work]

	baseURL, apiKey, model, size *core.Signal[string]

	works    *core.Signal[[]work]
	catalogs [3]*core.Resource[Catalog]

	studioNav, worksNav *base.Nav[route]
}

var fill = core.Style{FlexGrow: 1, FlexShrink: 1, FlexBasis: core.Pt(0)}

// Root is the app: the desktop studio in a wide window, the phone's tabs otherwise, with the generating screen and toasts over either.
func (a App) Root(cx *core.Scope) core.Node {
	theme, err := core.ParseTheme(themeYAML)
	if err != nil {
		panic(err) // the embedded theme is part of the build
	}
	cx.SetTheme(theme)
	s := newStudio(cx, a)
	colors := ui.ColorsOf(cx)
	env := cx.Env()
	return core.View{
		StyleFunc: func() core.Style {
			return core.Style{FlexGrow: 1, FlexShrink: 1, Background: colors.Background(), Color: colors.Text()}
		},
		Children: []core.Node{
			core.Show{
				When: func() bool { return env.Window().Width >= wideWidth },
				Then: s.desktop,
				Else: s.phone,
			},
			core.Show{When: s.busy.Get, Then: s.generatingScreen},
			ui.Toasts{Toaster: s.toaster, Placement: ui.ToastTop},
		},
	}
}

func newStudio(cx *core.Scope, a App) *studio {
	if a.Client == nil {
		a.Client = http.DefaultClient
	}
	store := cx.Storage()
	stored := func(key, fallback string) *core.Signal[string] {
		v, ok := store.GetItem(key)
		if !ok {
			v = fallback
		}
		return core.NewSignal(cx, v)
	}
	s := &studio{
		App:     a,
		cx:      cx,
		toaster: ui.NewToaster(cx),
		section: core.NewSignal(cx, sectionStudio),
		step:    core.NewSignal(cx, KindFace),
		note:    core.NewSignal(cx, ""),
		count:   core.NewSignal(cx, 1),
		busy:    core.NewSignal(cx, false),
		shown:   core.NewSignal[*work](cx, nil),
		baseURL: stored("baseURL", defaultBaseURL),
		apiKey:  core.NewSignal(cx, ""),
		model:   stored("model", defaultModel),
		size:    stored("size", defaultSize),
		works:   core.NewSignalFunc[[]work](cx, nil, nil),
	}
	for k := range s.chosen {
		kind := Kind(k)
		s.chosen[k] = core.NewSignal[*Item](cx, nil)
		s.filter[k] = core.NewSignal(cx, "")
		s.catalogs[k] = core.NewResource(cx, func() (Kind, bool) { return kind, true },
			func(ctx context.Context, k Kind, _ core.ResourceInfo) (Catalog, error) {
				return FetchCatalog(ctx, s.Client, s.Sites, k)
			})
	}
	// Poses open on the front view rather than every angle of every pose at once.
	s.filter[KindPose].Set(poseAngleOrder[0])
	s.studioNav = base.NewNav(cx, route{page: pageHome})
	s.worksNav = base.NewNav(cx, route{page: pageHome})

	secrets := cx.Secrets()
	dir := s.worksDir()
	go func() {
		key, ok, keyErr := secrets.Get(apiKeyAccount)
		list, listErr := listWorks(dir)
		cx.Post(func() {
			if keyErr == nil && ok && s.apiKey.Get() == "" {
				s.apiKey.Set(key)
			}
			if listErr == nil {
				s.works.Set(list)
			}
		})
	}()
	return s
}

func (s *studio) worksDir() string {
	if s.WorksDir != "" {
		return s.WorksDir
	}
	if d := s.cx.DataDir(); d != "" {
		return filepath.Join(d, "works")
	}
	return filepath.Join(os.TempDir(), "nsgo-pose", "works")
}

func (s *studio) choose(k Kind, it Item) {
	s.chosen[k].Set(&it)
	s.shown.Set(nil)
}

// request is the look as chosen; reading it inside a binding follows every choice.
func (s *studio) request() Request {
	return Request{Face: s.chosen[KindFace].Get(), Costume: s.chosen[KindCostume].Get(), Pose: s.chosen[KindPose].Get(),
		Note: s.note.Get(), Count: s.count.Get()}
}

func (s *studio) wide() bool { return s.cx.Env().Window().Width >= wideWidth }

func (s *studio) toastError(title string, err error) {
	s.toaster.Show(ui.ToastContent{Type: ui.ToastError, Title: title, Description: err.Error()})
}

// generate asks the image API for the look, keeps the pictures in Works and shows the first: on the desktop canvas, or as a page on a phone.
func (s *studio) generate() {
	t := locales.For(s.cx)
	req := s.request()
	if req.Validate() != nil || s.busy.Get() {
		return
	}
	cfg := Config{BaseURL: s.baseURL.Get(), APIKey: s.apiKey.Get(), Model: s.model.Get(), Size: s.size.Get()}
	if strings.TrimSpace(cfg.APIKey) == "" {
		s.toaster.Show(ui.ToastContent{Type: ui.ToastWarning, Title: t.NeedKey()})
		s.section.Set(sectionSettings)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	s.cancel = cancel
	s.busy.Set(true)
	dir, now, client := s.worksDir(), s.cx.Now(), s.Client
	go func() {
		pictures, err := Generate(ctx, client, cfg, req)
		// Whether the user cancelled is read before this goroutine's own cancel, which only releases the context.
		cancelled := errors.Is(ctx.Err(), context.Canceled)
		cancel()
		var made []work
		for i, pic := range pictures {
			if err != nil {
				break
			}
			var w work
			// Pictures of one request keep their order by a nanosecond apart.
			w, err = saveWork(dir, pic, now.Add(time.Duration(len(pictures)-i)))
			made = append(made, w)
		}
		s.cx.Post(func() {
			s.busy.Set(false)
			s.cancel = nil
			if cancelled {
				return
			}
			if err != nil {
				s.toastError(t.GenerateFailed(), err)
				return
			}
			s.works.Update(func(ws []work) []work { return append(append([]work(nil), made...), ws...) })
			if len(made) > 1 {
				s.toaster.Show(ui.ToastContent{Type: ui.ToastSuccess, Title: t.GeneratedMany(len(made))})
			}
			first := made[0]
			if s.wide() {
				s.section.Set(sectionStudio)
				s.shown.Set(&first)
				return
			}
			s.section.Set(sectionStudio)
			s.studioNav.SetRoutes(route{page: pageHome}, route{page: pageWork, work: first})
		})
	}()
}

func (s *studio) cancelGenerating() {
	if s.cancel != nil {
		s.cancel()
	}
}
