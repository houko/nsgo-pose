package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/yldm-ai/mirra/apps/mirra/locales"
)

func TestExplain(t *testing.T) {
	en, zh := locales.In("en"), locales.In("zh-Hans")
	refused := &statusError{Code: http.StatusUnauthorized, Status: "401 Unauthorized", Message: "Incorrect API key provided"}
	for _, c := range []struct {
		err  error
		want string
	}{
		{ErrNoAPIKey, en.ErrNoKey()},
		{fmt.Errorf("generate: %w", refused), en.ErrKeyRefused() + "\nIncorrect API key provided"},
		{&statusError{Code: http.StatusNotFound, Status: "404 Not Found"}, en.ErrNotFound() + "\n404 Not Found"},
		{&statusError{Code: http.StatusTooManyRequests, Status: "429"}, en.ErrRateLimited() + "\n429"},
		{&statusError{Code: http.StatusBadGateway, Status: "502 Bad Gateway"}, en.ErrServer() + "\n502 Bad Gateway"},
		{&statusError{Code: http.StatusBadRequest, Status: "400", Message: "sensitive content"}, en.ErrRejected() + "\nsensitive content"},
		{context.DeadlineExceeded, en.ErrTimeout()},
		{errNoImage, en.ErrNoImage()},
		{os.ErrPermission, os.ErrPermission.Error()},
	} {
		if got := explain(en, c.err); got != c.want {
			t.Errorf("explain(%v) = %q, want %q", c.err, got, c.want)
		}
	}
	if got := explain(zh, refused); got != zh.ErrKeyRefused()+"\nIncorrect API key provided" || got == explain(en, refused) {
		t.Errorf("Chinese explanation %q", got)
	}
}

// TestExplainNetworkErrors reads what the HTTP client returns for a server that is not there and for one that answers too late.
func TestExplainNetworkErrors(t *testing.T) {
	en := locales.In("en")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := srv.URL
	srv.Close()
	err := CheckConnection(context.Background(), http.DefaultClient, Config{BaseURL: addr, APIKey: "k"})
	if got := explain(en, err); len(got) <= len(en.ErrNetwork()) || got[:len(en.ErrNetwork())] != en.ErrNetwork() {
		t.Errorf("a closed server: %q", got)
	}
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		io.WriteString(w, "late")
	}))
	defer slow.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err = CheckConnection(ctx, http.DefaultClient, Config{BaseURL: slow.URL, APIKey: "k"})
	if got := explain(en, err); got != en.ErrTimeout() {
		t.Errorf("a slow server: %q (%v)", got, err)
	}
}
