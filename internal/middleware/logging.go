package middleware

import (
	"log"
	"net/http"
	"time"
)

type ResponseRecorder struct {
	statusCode int
	hasWritten bool
	http.ResponseWriter
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.hasWritten = true
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	if !r.hasWritten {
		r.WriteHeader(http.StatusOK)
	}

	return r.ResponseWriter.Write(data)
}

//func (r *ResponseRecorder) Header() http.Header {
//	return r.ResponseWriter.Header()
//}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &ResponseRecorder{
			statusCode:     http.StatusOK,
			hasWritten:     false,
			ResponseWriter: w,
		}

		// pass to rate limiter handler in middleware chain to write response
		next.ServeHTTP(recorder, r)

		// log output : <datetime> <status_code> <latency>ms <IP> <method> <path>
		// Ex: 2026/03/24 4:55:20    200    142ms    ::1    GET /get
		log.Printf(
			" %-4d %-14s %-21s %s %s", recorder.statusCode,
			time.Since(start), r.RemoteAddr, r.Method, r.URL.Path,
		)
	})
}
