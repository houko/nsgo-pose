package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/yldm-ai/mirra/apps/mirra/locales"
)

// explain is an error as the user reads it: what went wrong in their language, then, on a line of its own, what the service or the network said, where that adds something. An error of no kind it knows, such as a file's, is shown as it is.
func explain(t locales.Messages, err error) string {
	var status *statusError
	var netErr net.Error
	switch {
	case errors.Is(err, ErrNoAPIKey):
		return t.ErrNoKey()
	case errors.Is(err, errNoImage):
		return t.ErrNoImage()
	case errors.As(err, &status):
		detail := status.Message
		if detail == "" {
			detail = status.Status
		}
		return withDetail(statusText(t, status.Code), detail)
	case errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout():
		return t.ErrTimeout()
	case errors.As(err, &netErr):
		return withDetail(t.ErrNetwork(), err.Error())
	}
	return err.Error()
}

// statusText says what an HTTP status from an image API means for the user.
func statusText(t locales.Messages, code int) string {
	switch {
	case code == http.StatusUnauthorized || code == http.StatusForbidden:
		return t.ErrKeyRefused()
	case code == http.StatusNotFound:
		return t.ErrNotFound()
	case code == http.StatusTooManyRequests || code == http.StatusPaymentRequired:
		return t.ErrRateLimited()
	case code >= 500:
		return t.ErrServer()
	}
	return t.ErrRejected()
}

func withDetail(text, detail string) string {
	if detail = strings.TrimSpace(detail); detail == "" {
		return text
	}
	return text + "\n" + detail
}
