package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// The image APIs other than OpenAI's multipart /images/edits, each as its provider documents it. Every one takes the reference images inline, as base64, since the app holds them in memory and has no public URL for them.

// errNoImage is an API's answer that held no picture.
var errNoImage = errors.New("image API returned no image")

// dataURI is p as a data: URI.
func dataURI(p part) string {
	return "data:" + p.mime + ";base64," + base64.StdEncoding.EncodeToString(p.data)
}

// postJSON posts body as JSON to url with the headers and decodes a 2xx answer into out; any other answer is an error with the API's message.
func postJSON(ctx context.Context, client *http.Client, url string, headers map[string]string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return apiError(resp)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4*maxImageBytes)).Decode(out); err != nil {
		return fmt.Errorf("image API: %w", err)
	}
	return nil
}

// statusError is an API's failed answer: its HTTP status and the message it put in its body, if any.
type statusError struct {
	Code            int
	Status, Message string
}

func (e *statusError) Error() string {
	if e.Message != "" {
		return e.Status + ": " + e.Message
	}
	return "image API: " + e.Status
}

// apiError is a failed answer as an error, with the message the API put in its body: {"error":{"message"}} for most, {"message"} for DashScope.
func apiError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	var out struct {
		Error   json.RawMessage `json:"error"`
		Message string          `json:"message"`
	}
	msg := ""
	if json.Unmarshal(body, &out) == nil {
		var nested struct {
			Message string `json:"message"`
		}
		switch {
		case len(out.Error) > 0 && json.Unmarshal(out.Error, &nested) == nil && nested.Message != "":
			msg = nested.Message
		case len(out.Error) > 0 && out.Error[0] == '"':
			_ = json.Unmarshal(out.Error, &msg)
		default:
			msg = out.Message
		}
	}
	return &statusError{Code: resp.StatusCode, Status: resp.Status, Message: msg}
}

// bearer is the Authorization header most APIs take the key in, none without a key.
func bearer(key string) map[string]string {
	if key = strings.TrimSpace(key); key != "" {
		return map[string]string{"Authorization": "Bearer " + key}
	}
	return map[string]string{}
}

// endpoint joins the configured address and a path.
func endpoint(cfg Config, path string) string {
	return strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/") + path
}

// decodeImages turns an answer's images, base64 or links, into their bytes, downloading the links.
func decodeImages(ctx context.Context, client *http.Client, b64s, urls []string) ([][]byte, error) {
	var out [][]byte
	for _, b := range b64s {
		if i := strings.Index(b, ";base64,"); strings.HasPrefix(b, "data:") && i >= 0 {
			b = b[i+len(";base64,"):]
		}
		pic, err := base64.StdEncoding.DecodeString(b)
		if err != nil {
			return nil, fmt.Errorf("image API: %w", err)
		}
		out = append(out, pic)
	}
	for _, u := range urls {
		pic, err := get(ctx, client, u, maxImageBytes)
		if err != nil {
			return nil, err
		}
		out = append(out, pic)
	}
	if len(out) == 0 {
		return nil, errNoImage
	}
	return out, nil
}

// repeat makes n pictures with an API that makes one a call, calling it n times at once.
func repeat(ctx context.Context, n int, one func(context.Context) ([][]byte, error)) ([][]byte, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make([][][]byte, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			results[i], errs[i] = one(ctx)
			if errs[i] != nil {
				cancel()
			}
		})
	}
	wg.Wait()
	// A call's own failure says what went wrong; the others were only cancelled because of it.
	for _, err := range errs {
		if err != nil && !errors.Is(err, context.Canceled) {
			return nil, err
		}
	}
	var out [][]byte
	for i := range n {
		if errs[i] != nil {
			return nil, errs[i]
		}
		out = append(out, results[i]...)
	}
	return out, nil
}

// dimensions reads a size written as "1024x1536" or "1024*1536".
func dimensions(size string) (w, h int, ok bool) {
	a, b, found := strings.Cut(strings.ReplaceAll(strings.TrimSpace(size), "*", "x"), "x")
	if !found {
		return 0, 0, false
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(a))
	h, err2 := strconv.Atoi(strings.TrimSpace(b))
	if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
		return 0, 0, false
	}
	return w, h, true
}

// aspectRatio is the ratio among allowed closest to size's, for APIs that take a ratio rather than pixels; a size that is already a ratio, such as "2:3", passes as it is.
func aspectRatio(size string, allowed []string) string {
	if strings.Contains(size, ":") {
		return strings.TrimSpace(size)
	}
	w, h, ok := dimensions(size)
	if !ok {
		w, h, _ = dimensions(defaultSize)
	}
	want := math.Log(float64(w) / float64(h))
	best, bestDiff := allowed[0], math.Inf(1)
	for _, r := range allowed {
		a, b, _ := strings.Cut(r, ":")
		x, err1 := strconv.ParseFloat(a, 64)
		y, err2 := strconv.ParseFloat(b, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		if d := math.Abs(math.Log(x/y) - want); d < bestDiff {
			best, bestDiff = r, d
		}
	}
	return best
}

// ---- Azure OpenAI ----

// generateAzure posts OpenAI's multipart edit to Azure OpenAI's v1 API, whose image routes are in its preview version, with the key in an api-key header.
func generateAzure(ctx context.Context, client *http.Client, cfg Config, prompt string, parts []part, n int) ([][]byte, error) {
	return postEdits(ctx, client, cfg, endpoint(cfg, "/images/edits")+"?api-version=preview", map[string]string{"api-key": strings.TrimSpace(cfg.APIKey)}, prompt, parts, n)
}

// ---- Google Gemini ----

var geminiRatios = []string{"1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9"}

// generateGemini asks Gemini's generateContent for an image from the prompt and the images as inline parts. It answers with one image a call, so n pictures take n calls.
func generateGemini(ctx context.Context, client *http.Client, cfg Config, prompt string, parts []part, n int) ([][]byte, error) {
	type inline struct {
		MimeType string `json:"mimeType"`
		Data     string `json:"data"`
	}
	type geminiPart struct {
		Text       string  `json:"text,omitempty"`
		InlineData *inline `json:"inlineData,omitempty"`
	}
	content := []geminiPart{{Text: prompt}}
	for _, p := range parts {
		content = append(content, geminiPart{InlineData: &inline{MimeType: p.mime, Data: base64.StdEncoding.EncodeToString(p.data)}})
	}
	body := map[string]any{
		"contents": []any{map[string]any{"role": "user", "parts": content}},
		"generationConfig": map[string]any{
			"responseModalities": []string{"TEXT", "IMAGE"},
			"imageConfig":        map[string]any{"aspectRatio": aspectRatio(cfg.Size, geminiRatios)},
		},
	}
	url := endpoint(cfg, "/models/"+strings.TrimSpace(cfg.Model)+":generateContent")
	headers := map[string]string{"x-goog-api-key": strings.TrimSpace(cfg.APIKey)}
	return repeat(ctx, n, func(ctx context.Context) ([][]byte, error) {
		var out struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						InlineData *inline `json:"inlineData"`
					} `json:"parts"`
				} `json:"content"`
				FinishReason string `json:"finishReason"`
			} `json:"candidates"`
			PromptFeedback *struct {
				BlockReason string `json:"blockReason"`
			} `json:"promptFeedback"`
		}
		if err := postJSON(ctx, client, url, headers, body, &out); err != nil {
			return nil, err
		}
		var b64s []string
		for _, c := range out.Candidates {
			for _, p := range c.Content.Parts {
				if p.InlineData != nil && strings.HasPrefix(p.InlineData.MimeType, "image/") {
					b64s = append(b64s, p.InlineData.Data)
				}
			}
		}
		if len(b64s) == 0 {
			if out.PromptFeedback != nil && out.PromptFeedback.BlockReason != "" {
				return nil, fmt.Errorf("image API: blocked: %s", out.PromptFeedback.BlockReason)
			}
			if len(out.Candidates) > 0 && out.Candidates[0].FinishReason != "" {
				return nil, fmt.Errorf("image API returned no image: %s", out.Candidates[0].FinishReason)
			}
		}
		return decodeImages(ctx, client, b64s, nil)
	})
}

// ---- Volcengine Ark (Seedream) ----

// generateSeedream asks Ark's /images/generations for an image from the prompt and the images as data URIs, without the watermark Ark adds by default. Seedream 5.x makes one image a call, so n pictures take n calls.
func generateSeedream(ctx context.Context, client *http.Client, cfg Config, prompt string, parts []part, n int) ([][]byte, error) {
	images := make([]string, len(parts))
	for i, p := range parts {
		images[i] = dataURI(p)
	}
	body := map[string]any{"model": strings.TrimSpace(cfg.Model), "prompt": prompt, "image": images, "size": strings.TrimSpace(cfg.Size),
		"response_format": "b64_json", "watermark": false}
	return repeat(ctx, n, func(ctx context.Context) ([][]byte, error) {
		return dataImages(ctx, client, endpoint(cfg, "/images/generations"), bearer(cfg.APIKey), body)
	})
}

// dataImages posts body and reads the OpenAI-shaped answer, {"data":[{"b64_json"|"url"}]}, that Ark, OpenRouter and xAI share.
func dataImages(ctx context.Context, client *http.Client, url string, headers map[string]string, body any) ([][]byte, error) {
	var out struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
			Error   *struct {
				Message string `json:"message"`
			} `json:"error"`
		} `json:"data"`
	}
	if err := postJSON(ctx, client, url, headers, body, &out); err != nil {
		return nil, err
	}
	var b64s, urls []string
	var failed string
	for _, d := range out.Data {
		switch {
		case d.B64JSON != "":
			b64s = append(b64s, d.B64JSON)
		case d.URL != "":
			urls = append(urls, d.URL)
		case d.Error != nil:
			failed = d.Error.Message
		}
	}
	if len(b64s)+len(urls) == 0 && failed != "" {
		return nil, fmt.Errorf("image API: %s", failed)
	}
	return decodeImages(ctx, client, b64s, urls)
}

// ---- xAI ----

var xaiRatios = []string{"1:1", "3:4", "4:3", "9:16", "16:9", "2:3", "3:2", "1:2", "2:1"}

// generateXAI posts xAI's JSON /images/edits, its images as data URIs.
func generateXAI(ctx context.Context, client *http.Client, cfg Config, prompt string, parts []part, n int) ([][]byte, error) {
	images := make([]map[string]string, len(parts))
	for i, p := range parts {
		images[i] = map[string]string{"url": dataURI(p)}
	}
	body := map[string]any{"model": strings.TrimSpace(cfg.Model), "prompt": prompt, "images": images, "n": n,
		"aspect_ratio": aspectRatio(cfg.Size, xaiRatios), "resolution": "1k", "response_format": "b64_json"}
	return dataImages(ctx, client, endpoint(cfg, "/images/edits"), bearer(cfg.APIKey), body)
}

// ---- OpenRouter ----

var openRouterRatios = []string{"1:1", "2:3", "3:2", "3:4", "4:3", "9:16", "16:9"}

// generateOpenRouter posts OpenRouter's /images, its images as input references.
func generateOpenRouter(ctx context.Context, client *http.Client, cfg Config, prompt string, parts []part, n int) ([][]byte, error) {
	refs := make([]map[string]any, len(parts))
	for i, p := range parts {
		refs[i] = map[string]any{"type": "image_url", "image_url": map[string]string{"url": dataURI(p)}}
	}
	body := map[string]any{"model": strings.TrimSpace(cfg.Model), "prompt": prompt, "input_references": refs, "n": n,
		"aspect_ratio": aspectRatio(cfg.Size, openRouterRatios), "resolution": "1K", "response_format": "b64_json"}
	return dataImages(ctx, client, endpoint(cfg, "/images"), bearer(cfg.APIKey), body)
}

// ---- Alibaba Cloud Model Studio (DashScope) ----

// generateDashScope posts Qwen image editing's multimodal generation, the images then the prompt in one message. It answers with links, which are downloaded.
func generateDashScope(ctx context.Context, client *http.Client, cfg Config, prompt string, parts []part, n int) ([][]byte, error) {
	var content []map[string]string
	for _, p := range parts {
		content = append(content, map[string]string{"image": dataURI(p)})
	}
	content = append(content, map[string]string{"text": prompt})
	params := map[string]any{"n": n, "watermark": false}
	if w, h, ok := dimensions(cfg.Size); ok {
		params["size"] = fmt.Sprintf("%d*%d", w, h)
	}
	body := map[string]any{"model": strings.TrimSpace(cfg.Model),
		"input":      map[string]any{"messages": []any{map[string]any{"role": "user", "content": content}}},
		"parameters": params}
	var out struct {
		Output struct {
			Choices []struct {
				Message struct {
					Content []struct {
						Image string `json:"image"`
					} `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		} `json:"output"`
	}
	if err := postJSON(ctx, client, endpoint(cfg, dashScopePath), bearer(cfg.APIKey), body, &out); err != nil {
		return nil, err
	}
	var urls []string
	for _, c := range out.Output.Choices {
		for _, it := range c.Message.Content {
			if it.Image != "" {
				urls = append(urls, it.Image)
			}
		}
	}
	return decodeImages(ctx, client, nil, urls)
}

const dashScopePath = "/services/aigc/multimodal-generation/generation"
