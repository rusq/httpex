// Package httpex provides some useful HTTP middlewares.
package httpex

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Neuter is the middleware that returns 404 on any non-direct directory path
// except the root.
// https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings
func Neuter(path string, next http.Handler) http.Handler {
	var rootLen = len(path)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, path) && len(r.URL.Path) > rootLen {
			log.Printf("neutralised: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// statusRecorder is a wrapper for http.ResponseWriter that keeps track of the
// response status.
type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (sr *statusRecorder) WriteHeader(status int) {
	sr.Status = status
	sr.ResponseWriter.WriteHeader(status)
}

type reqID int

var reqIDkey reqID

type RequestID string

const RUnknown RequestID = "unknown"

func newReqIDContext(parent context.Context, id RequestID) context.Context {
	return context.WithValue(parent, reqIDkey, id)
}

func ContextRequestID(ctx context.Context) (RequestID, bool) {
	id, ok := ctx.Value(reqIDkey).(RequestID)
	return id, ok
}

var counter atomic.Int64

// LogMiddleware logs the request.
func LogMiddleware(next http.Handler, lg *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wr := &statusRecorder{w, 200}
		reqID := RequestID(strconv.FormatInt(counter.Add(1), 10))
		r = r.WithContext(newReqIDContext(r.Context(), reqID))
		next.ServeHTTP(wr, r)
		lg.Printf("[%s] HTTP %s %s - %d %5dms (%s) %s", reqID, r.Method, r.URL.Path, wr.Status, time.Since(start).Milliseconds(), RequestIPAddr(r), r.UserAgent())
	})
}

func RequestIPAddr(r *http.Request) string {
	return r.RemoteAddr // for now
}
