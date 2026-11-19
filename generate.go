package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path"
	"strings"
)

// Config says which OpenAI-compatible image API generates the picture.
type Config struct {
	BaseURL string
	APIKey  string
	Model   string
	Size    string
}

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-image-2"
	defaultSize    = "1024x1536"
)

// Source is a reference image: a gallery item's full image, downloaded when generating, or a photo the user picked.
type Source struct {
	Item  *Item
	Photo []byte
	Name  string
	MIME  string
}

// Request is what the user put together: a pose with an outfit, a face or both, or the user's own photo with a face.
type Request struct {
	Base    Source
	Costume *Item
	Face    *Item
	Note    string
}

// Errors the UI explains in the user's language.
var (
	ErrNoBase      = errors.New("no base image")
	ErrNothingToDo = errors.New("a photo needs a face")
	ErrPhotoOutfit = errors.New("outfits go on poses, not photos")
	ErrNoAPIKey    = errors.New("no API key")
)

// Validate reports what the request still lacks.
func (r Request) Validate() error {
	switch {
	case r.Base.Item == nil && len(r.Base.Photo) == 0:
		return ErrNoBase
	// The outfit gallery's terms forbid depicting real people with its images, so an outfit never goes on the user's photo.
	case r.Base.Item == nil && r.Costume != nil:
		return ErrPhotoOutfit
	case r.Base.Item == nil && r.Face == nil:
		return ErrNothingToDo
	}
	return nil
}

// Prompt says what each attached image is for, in the order the images are sent.
func (r Request) Prompt() string {
	var b strings.Builder
	n := 1
	if r.Base.Item != nil {
		fmt.Fprintf(&b, "Image %d is a faceless gray mannequin showing a pose and a camera angle. Create one photorealistic photo of a single adult person in exactly that pose, framing and camera angle, on a plain light studio background.", n)
	} else {
		fmt.Fprintf(&b, "Image %d is a photo of a person. Keep their pose, body shape, camera angle, framing, lighting and background exactly as they are.", n)
	}
	n++
	if r.Costume != nil {
		fmt.Fprintf(&b, " Image %d shows an outfit with no one wearing it: dress the person in exactly that outfit, with the same garments, colors, materials, patterns and details, fitted naturally to their body and pose.", n)
		n++
	} else if r.Base.Item != nil {
		b.WriteString(" Dress them in simple, plain everyday clothes.")
	}
	if r.Face != nil {
		if r.Base.Item != nil {
			fmt.Fprintf(&b, " Image %d is a portrait: give the person that face, with the same facial features, skin tone, hairstyle and hair color.", n)
		} else {
			fmt.Fprintf(&b, " Image %d is a portrait: replace the person's face and hair with that face, hairstyle and hair color, blended naturally into the photo's lighting.", n)
		}
	} else if r.Base.Item != nil {
		b.WriteString(" Give them a natural adult face.")
	}
	if note := strings.TrimSpace(r.Note); note != "" {
		b.WriteString(" Also: ")
		b.WriteString(note)
	}
	b.WriteString(" Output a single image of one adult person; no text, no watermark.")
	return b.String()
}

// maxImageBytes bounds a downloaded reference or generated image.
const maxImageBytes = 50 << 20

type part struct {
	name, mime string
	data       []byte
}

// Generate downloads the request's gallery images, one each, sends them with the prompt to cfg's /images/edits, and returns the picture's bytes.
func Generate(ctx context.Context, client *http.Client, cfg Config, r Request) ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrNoAPIKey
	}
	var parts []part
	if r.Base.Item != nil {
		p, err := download(ctx, client, *r.Base.Item)
		if err != nil {
			return nil, err
		}
		parts = append(parts, p)
	} else {
		parts = append(parts, part{name: or(r.Base.Name, "photo.jpg"), mime: or(r.Base.MIME, "image/jpeg"), data: r.Base.Photo})
	}
	for _, it := range []*Item{r.Costume, r.Face} {
		if it == nil {
			continue
		}
		p, err := download(ctx, client, *it)
		if err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fields := [][2]string{{"model", or(cfg.Model, defaultModel)}, {"prompt", r.Prompt()}, {"size", or(cfg.Size, defaultSize)}, {"n", "1"}}
	for _, f := range fields {
		if err := w.WriteField(f[0], f[1]); err != nil {
			return nil, err
		}
	}
	for _, p := range parts {
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image[]"; filename=%q`, p.name))
		h.Set("Content-Type", p.mime)
		fw, err := w.CreatePart(h)
		if err != nil {
			return nil, err
		}
		if _, err := fw.Write(p.data); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	endpoint := strings.TrimRight(or(cfg.BaseURL, defaultBaseURL), "/") + "/images/edits"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(cfg.APIKey))
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	decodeErr := json.NewDecoder(resp.Body).Decode(&out)
	if resp.StatusCode != http.StatusOK {
		if decodeErr == nil && out.Error != nil && out.Error.Message != "" {
			return nil, fmt.Errorf("%s: %s", resp.Status, out.Error.Message)
		}
		return nil, fmt.Errorf("image API: %s", resp.Status)
	}
	if decodeErr != nil {
		return nil, fmt.Errorf("image API: %w", decodeErr)
	}
	if len(out.Data) == 0 {
		return nil, errors.New("image API returned no image")
	}
	if b64 := out.Data[0].B64JSON; b64 != "" {
		return base64.StdEncoding.DecodeString(b64)
	}
	if u := out.Data[0].URL; u != "" {
		return get(ctx, client, u, maxImageBytes)
	}
	return nil, errors.New("image API returned no image")
}

func download(ctx context.Context, client *http.Client, it Item) (part, error) {
	data, err := get(ctx, client, it.Full, maxImageBytes)
	if err != nil {
		return part{}, err
	}
	name := path.Base(it.Full)
	return part{name: name, mime: http.DetectContentType(data), data: data}, nil
}

func or(v, fallback string) string {
	if v = strings.TrimSpace(v); v != "" {
		return v
	}
	return fallback
}
