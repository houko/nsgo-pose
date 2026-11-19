package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

// Kind is one of the three reference galleries.
type Kind int

const (
	KindPose Kind = iota
	KindCostume
	KindFace
)

// Sites are the galleries' origins, overridable in tests.
type Sites struct {
	Pose, Costume, Face string
}

// DefaultSites are the public galleries the app reads.
var DefaultSites = Sites{
	Pose:    "https://ai-pose-gallery.pages.dev",
	Costume: "https://ai-costume-gallery.pages.dev",
	Face:    "https://ai-portrait-gallery.pages.dev",
}

func (s Sites) origin(k Kind) string {
	switch k {
	case KindPose:
		return s.Pose
	case KindCostume:
		return s.Costume
	default:
		return s.Face
	}
}

// Item is one reference image: Thumb is shown while browsing, Full is downloaded once, only when the user generates with it.
type Item struct {
	Kind Kind
	ID   string
	// Group is what the picker's chips filter by: a pose's camera angle, an outfit's category, a face's hair length.
	Group string
	// Title is a pose's name, an outfit's keywords, or a face's hairstyle, which may be empty.
	Title string
	Thumb string
	Full  string
}

// Catalog is a gallery's usable items and their groups in the gallery's order.
type Catalog struct {
	Items  []Item
	Groups []string
}

// InGroup returns the items of one group, all of them for "".
func (c Catalog) InGroup(group string) []Item {
	if group == "" {
		return c.Items
	}
	var out []Item
	for _, it := range c.Items {
		if it.Group == group {
			out = append(out, it)
		}
	}
	return out
}

// Content the app never offers. The pose gallery has a line of two-person sexual positions and explicit single poses, and the costume gallery has lingerie and "sexy" categories; none of these belong in an app that puts real-looking faces and users' own photos into the result.
var (
	excludedPoseLines      = map[string]bool{"2人絡み": true}
	excludedPoseCategories = map[string]bool{"セクシー・露骨寄り系": true, "胸寄せ・谷間強調系": true}
	excludedCostumeCats    = map[string]bool{"ランジェリー": true, "セクシー": true}
)

// maxIndexBytes bounds a gallery's index; the real ones are under 1 MB.
const maxIndexBytes = 16 << 20

// FetchCatalog reads a gallery's index: the data array its page embeds for poses and costumes, and data/images.js for faces. It reads metadata only; no image is downloaded.
func FetchCatalog(ctx context.Context, client *http.Client, sites Sites, k Kind) (Catalog, error) {
	origin := strings.TrimRight(sites.origin(k), "/")
	path := "/"
	if k == KindFace {
		path = "/data/images.js"
	}
	body, err := get(ctx, client, origin+path, maxIndexBytes)
	if err != nil {
		return Catalog{}, err
	}
	switch k {
	case KindPose:
		return parsePoses(body, origin)
	case KindCostume:
		return parseCostumes(body, origin)
	default:
		return parseFaces(body, origin)
	}
}

func get(ctx context.Context, client *http.Client, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("GET %s: larger than %d bytes", url, limit)
	}
	return body, nil
}

// embeddedArray decodes the JSON array that follows marker in a page's script.
func embeddedArray(page []byte, marker string, v any) error {
	i := bytes.Index(page, []byte(marker))
	if i < 0 {
		return fmt.Errorf("gallery format changed: %q not found", marker)
	}
	if err := json.NewDecoder(bytes.NewReader(page[i+len(marker):])).Decode(v); err != nil {
		return fmt.Errorf("gallery format changed: %w", err)
	}
	return nil
}

type groupOrder struct {
	seen map[string]bool
	list []string
}

func (g *groupOrder) add(name string) {
	if g.seen == nil {
		g.seen = map[string]bool{}
	}
	if !g.seen[name] {
		g.seen[name] = true
		g.list = append(g.list, name)
	}
}

func parsePoses(page []byte, origin string) (Catalog, error) {
	var rows []struct {
		ID       string `json:"id"`
		Line     string `json:"系統"`
		Category string `json:"分類"`
		Name     string `json:"ポーズ名"`
		Angle    string `json:"アングル"`
		Thumb    string `json:"サムネ"`
		Full     string `json:"原寸"`
	}
	if err := embeddedArray(page, "const D=", &rows); err != nil {
		return Catalog{}, err
	}
	var c Catalog
	var groups groupOrder
	lines := map[string]string{}
	for _, r := range rows {
		if excludedPoseLines[r.Line] || excludedPoseCategories[r.Category] || r.ID == "" || r.Full == "" {
			continue
		}
		groups.add(r.Angle)
		lines[r.ID] = r.Line
		c.Items = append(c.Items, Item{Kind: KindPose, ID: r.ID, Group: r.Angle, Title: r.Name,
			Thumb: origin + "/" + r.Thumb, Full: origin + "/" + r.Full})
	}
	// The gallery lists its glamour line first; the app opens on the everyday poses.
	slices.SortStableFunc(c.Items, func(a, b Item) int { return rank(poseLineOrder, lines[a.ID]) - rank(poseLineOrder, lines[b.ID]) })
	slices.SortStableFunc(groups.list, func(a, b string) int { return rank(poseAngleOrder, a) - rank(poseAngleOrder, b) })
	c.Groups = groups.list
	return nonEmpty(c)
}

var (
	poseLineOrder  = []string{"全身1人・標準系", "上半身1人", "全身1人・セクシー系"}
	poseAngleOrder = []string{"正面", "斜め", "横", "上から", "下から"}
	faceHairOrder  = []string{"ショート", "ミディアム", "セミロング", "ロング"}
)

// rank is where v comes in order, after every listed value when it is not listed.
func rank(order []string, v string) int {
	if i := slices.Index(order, v); i >= 0 {
		return i
	}
	return len(order)
}

func parseCostumes(page []byte, origin string) (Catalog, error) {
	var rows []struct {
		ID       string `json:"id"`
		Category string `json:"カテゴリ"`
		Keywords string `json:"キーワード"`
		Thumb    string `json:"サムネ"`
		Full     string `json:"原寸"`
	}
	if err := embeddedArray(page, "const D=", &rows); err != nil {
		return Catalog{}, err
	}
	var c Catalog
	var groups groupOrder
	for _, r := range rows {
		if excludedCostumeCats[r.Category] || r.ID == "" || r.Full == "" {
			continue
		}
		groups.add(r.Category)
		c.Items = append(c.Items, Item{Kind: KindCostume, ID: r.ID, Group: r.Category, Title: r.Keywords,
			Thumb: origin + "/" + r.Thumb, Full: origin + "/" + r.Full})
	}
	c.Groups = groups.list
	return nonEmpty(c)
}

// parseFaces keeps the gallery's front-facing faces, the "face" section, grouped by hair length; its "styling" section shows whole outfits at any angle, which make weaker face references.
func parseFaces(script []byte, origin string) (Catalog, error) {
	var rows []struct {
		ID      string `json:"id"`
		Section string `json:"section"`
		Hair    string `json:"hair"`
		Style   string `json:"style"`
		Full    string `json:"full"`
		Thumb   string `json:"thumb"`
	}
	if err := embeddedArray(script, "window.GALLERY_IMAGES =", &rows); err != nil {
		return Catalog{}, err
	}
	var c Catalog
	var groups groupOrder
	for _, r := range rows {
		if r.Section != "face" || r.ID == "" || r.Full == "" {
			continue
		}
		groups.add(r.Hair)
		c.Items = append(c.Items, Item{Kind: KindFace, ID: r.ID, Group: r.Hair, Title: r.Style,
			Thumb: origin + "/" + r.Thumb, Full: origin + "/" + r.Full})
	}
	slices.SortStableFunc(groups.list, func(a, b string) int { return rank(faceHairOrder, a) - rank(faceHairOrder, b) })
	c.Groups = groups.list
	return nonEmpty(c)
}

func nonEmpty(c Catalog) (Catalog, error) {
	if len(c.Items) == 0 {
		return Catalog{}, errors.New("gallery format changed: no items")
	}
	return c, nil
}
