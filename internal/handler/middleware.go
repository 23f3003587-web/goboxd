package handler

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

func Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC RECOVERED: %v\n%s", rec, debug.Stack())
				http.Error(w, `{"error":{"code":"internal_error","message":"Server error"}}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
}

func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	}
}
