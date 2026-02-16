package logging

import (
  "crypto/rand"
  "encoding/hex"
  "log"
  "net/http"
  "time"
)

type ResponseWriter struct {
  http.ResponseWriter
  Status int
}

func (lrw *ResponseWriter) WriteHeader(code int) {
  lrw.Status = code
  lrw.ResponseWriter.WriteHeader(code)
}

func WithRequestID(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    rid := r.Header.Get("X-Request-Id")
    if rid == "" {
      rid = NewRequestID()
    }
    w.Header().Set("X-Request-Id", rid)

    lrw := &ResponseWriter{ResponseWriter: w, Status: http.StatusOK}
    start := time.Now()
    next.ServeHTTP(lrw, r)
    log.Printf("rid=%s method=%s path=%s status=%d dur=%s", rid, r.Method, r.URL.Path, lrw.Status, time.Since(start))
  })
}

func NewRequestID() string {
  buf := make([]byte, 8)
  if _, err := rand.Read(buf); err != nil {
    return ""
  }
  return hex.EncodeToString(buf)
}
