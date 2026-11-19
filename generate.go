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
	"strconv"
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
	// maxCount bounds how many pictures one request asks for.
	maxCount = 4
)

// Request is what the user put together: a pose, with an outfit and a face when chosen, all from the galleries. The app takes no photo of the user: the pose and outfit galleries' terms forbid depicting real people with their images.
type Request struct {
	Pose    *Item
	Costume *Item
	Face    *Item
	Note    string
	// Count is how many pictures to ask for, 1 when 0.
	Count int
}

// Errors the UI explains in the user's language.
var (
	ErrNoPose   = errors.New("no pose chosen")
	ErrNoAPIKey = errors.New("no API key")
)

// Validate reports what the request still lacks.
func (r Request) Validate() error {
	if r.Pose == nil {
		return ErrNoPose
	}
	return nil
}

func (r Request) count() int { return min(max(r.Count, 1), maxCount) }

// Prompt says what each attached image is for, in the order the images are sent.
func (r Request) Prompt() string {
	var b strings.Builder
	b.WriteString("Image 1 is a faceless gray mannequin showing a pose and a camera angle. Create one photorealistic photo of a single adult person in exactly that pose, framing and camera angle, on a plain light studio background.")
	n := 2
	if r.Costume != nil {
		fmt.Fprintf(&b, " Image %d shows an outfit with no one wearing it: dress the person in exactly that outfit, with the same garments, colors, materials, patterns and details, fitted naturally to their body and pose.", n)
		n++
	} else {
		b.WriteString(" Dress them in simple, plain everyday clothes.")
	}
	if r.Face != nil {
		fmt.Fprintf(&b, " Image %d is a portrait: give the person that face, with the same facial features, skin tone, hairstyle and hair color.", n)
	} else {
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

// Generate downloads the request's gallery images, one each, sends them with the prompt to cfg's /images/edits, and returns the pictures' bytes.
func Generate(ctx context.Context, client *http.Client, cfg Config, r Request) ([][]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrNoAPIKey
	}
	var parts []part
	for _, it := range []*Item{r.Pose, r.Costume, r.Face} {
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
	fields := [][2]string{{"model", or(cfg.Model, defaultModel)}, {"prompt", r.Prompt()}, {"size", or(cfg.Size, defaultSize)}, {"n", strconv.Itoa(r.count())}}
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
	var pictures [][]byte
	for _, d := range out.Data {
		switch {
		case d.B64JSON != "":
			pic, err := base64.StdEncoding.DecodeString(d.B64JSON)
			if err != nil {
				return nil, fmt.Errorf("image API: %w", err)
			}
			pictures = append(pictures, pic)
		case d.URL != "":
			pic, err := get(ctx, client, d.URL, maxImageBytes)
			if err != nil {
				return nil, err
			}
			pictures = append(pictures, pic)
		}
	}
	if len(pictures) == 0 {
		return nil, errors.New("image API returned no image")
	}
	return pictures, nil
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
