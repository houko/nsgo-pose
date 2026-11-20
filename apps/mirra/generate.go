package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// Config says which image API generates the picture, and in which of the shapes the app speaks.
type Config struct {
	Protocol protocol
	BaseURL  string
	APIKey   string
	Model    string
	Size     string
	// Local marks a server the user runs, which may take no key.
	Local bool
	// ReadLocal reads a picture the app keeps, such as the user's clothes; os.ReadFile when nil.
	ReadLocal func(path string) ([]byte, error)
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
	if r.Pose != nil && r.Pose.People == 2 {
		return r.pairPrompt()
	}
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
	r.writeNote(&b)
	b.WriteString(" Output a single image of one adult person; no text, no watermark.")
	return b.String()
}

// pairPrompt is the prompt of a two-person pose: the look's outfit and face go to the person the pose centers on, and the other person is an adult of their own.
func (r Request) pairPrompt() string {
	var b strings.Builder
	b.WriteString("Image 1 shows two faceless gray mannequins in a pose and a camera angle. Create one photorealistic photo of two adult people in exactly that pose, framing and camera angle, on a plain light studio background. Call the more prominent figure the main person and the other the partner.")
	n := 2
	if r.Costume != nil {
		fmt.Fprintf(&b, " Image %d shows an outfit with no one wearing it: dress the main person in exactly that outfit, with the same garments, colors, materials, patterns and details, fitted naturally to their body and pose. Dress the partner in simple clothes that suit it.", n)
		n++
	} else {
		b.WriteString(" Dress both in simple, plain everyday clothes.")
	}
	if r.Face != nil {
		fmt.Fprintf(&b, " Image %d is a portrait: give the main person that face, with the same facial features, skin tone, hairstyle and hair color, and give the partner a different natural adult face.", n)
	} else {
		b.WriteString(" Give both natural, different adult faces.")
	}
	r.writeNote(&b)
	b.WriteString(" Output a single image of these two adults; no text, no watermark.")
	return b.String()
}

// writeNote adds the user's own words, if any.
func (r Request) writeNote(b *strings.Builder) {
	if note := strings.TrimSpace(r.Note); note != "" {
		b.WriteString(" Also: ")
		b.WriteString(note)
	}
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
	if !cfg.Local && strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrNoAPIKey
	}
	var parts []part
	for _, it := range []*Item{r.Pose, r.Costume, r.Face} {
		if it == nil {
			continue
		}
		p, err := download(ctx, client, cfg.ReadLocal, *it)
		if err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}

	return generate(ctx, client, cfg, r.Prompt(), parts, r.count())
}

// generate sends the prompt and reference images to cfg's API in its protocol and returns n pictures.
func generate(ctx context.Context, client *http.Client, cfg Config, prompt string, parts []part, n int) ([][]byte, error) {
	switch cfg.Protocol {
	case protoAzure:
		return generateAzure(ctx, client, cfg, prompt, parts, n)
	case protoGemini:
		return generateGemini(ctx, client, cfg, prompt, parts, n)
	case protoSeedream:
		return generateSeedream(ctx, client, cfg, prompt, parts, n)
	case protoXAI:
		return generateXAI(ctx, client, cfg, prompt, parts, n)
	case protoOpenRouter:
		return generateOpenRouter(ctx, client, cfg, prompt, parts, n)
	case protoDashScope:
		return generateDashScope(ctx, client, cfg, prompt, parts, n)
	}
	return generateOpenAI(ctx, client, cfg, prompt, parts, n)
}

// generateOpenAI posts the prompt and the images to OpenAI's /images/edits, or an OpenAI-compatible server's.
func generateOpenAI(ctx context.Context, client *http.Client, cfg Config, prompt string, parts []part, n int) ([][]byte, error) {
	return postEdits(ctx, client, cfg, endpoint(cfg, "/images/edits"), bearer(cfg.APIKey), prompt, parts, n)
}

// postEdits posts OpenAI's multipart edit, each image an image[] part, to url with the headers.
func postEdits(ctx context.Context, client *http.Client, cfg Config, url string, headers map[string]string, prompt string, parts []part, n int) ([][]byte, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fields := [][2]string{{"model", or(cfg.Model, defaultModel)}, {"prompt", prompt}, {"size", or(cfg.Size, defaultSize)}, {"n", strconv.Itoa(n)}}
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}
	var out struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("image API: %w", err)
	}
	var b64s, urls []string
	for _, d := range out.Data {
		if d.B64JSON != "" {
			b64s = append(b64s, d.B64JSON)
		} else if d.URL != "" {
			urls = append(urls, d.URL)
		}
	}
	return decodeImages(ctx, client, b64s, urls)
}

// download reads an item's full image: a gallery's over the network, or a photo the user added from the pictures the app keeps.
func download(ctx context.Context, client *http.Client, readLocal func(string) ([]byte, error), it Item) (part, error) {
	var data []byte
	var err error
	if filepath.IsAbs(it.Full) {
		if readLocal == nil {
			readLocal = os.ReadFile
		}
		data, err = readLocal(it.Full)
	} else {
		data, err = get(ctx, client, it.Full, maxImageBytes)
	}
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
