package mypod

import (
	"net/http"
	"time"

	"github.com/parkr/radar"
	"github.com/technoweenie/grohl"
)

// This injects additional context into the grohl context
func AdditionalLogContextHandler(h http.Handler) http.Handler {
	return &additionalLoggingHandler{next: h}
}

type additionalLoggingHandler struct {
	next http.Handler
}

func (h *additionalLoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logCtx := radar.GetLogContext(r)
	rl := &responseLogger{rw: w, start: time.Now()}
	h.next.ServeHTTP(rl, r)
	_ = logCtx.Log(grohl.Data{
		"at":      "done",
		"status":  rl.status,
		"size":    rl.size,
		"elapsed": time.Since(rl.start).String(),
		"client":  remoteAddr(r),
	})
}

func remoteAddr(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return forwardedFor
	}
	return r.RemoteAddr
}

type responseLogger struct {
	rw     http.ResponseWriter
	start  time.Time
	status int
	size   int
}

func (rl *responseLogger) Header() http.Header {
	return rl.rw.Header()
}

func (rl *responseLogger) Write(bytes []byte) (int, error) {
	if rl.status == 0 {
		rl.status = http.StatusOK
	}

	size, err := rl.rw.Write(bytes)

	rl.size += size

	return size, err
}

func (rl *responseLogger) WriteHeader(status int) {
	rl.status = status

	rl.rw.WriteHeader(status)
}

func (rl *responseLogger) Flush() {
	f, ok := rl.rw.(http.Flusher)

	if ok {
		f.Flush()
	}
}
