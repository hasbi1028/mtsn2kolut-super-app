package middleware

import (
	"log"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		reqID := chimw.GetReqID(r.Context())
		log.Printf(
			"http request_id=%s method=%s path=%s status=%d bytes=%d duration_ms=%d remote=%s",
			reqID,
			r.Method,
			r.URL.RequestURI(),
			rec.status,
			rec.bytes,
			time.Since(start).Milliseconds(),
			r.RemoteAddr,
		)
	})
}
