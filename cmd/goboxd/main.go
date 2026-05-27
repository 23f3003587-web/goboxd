package main

import (
	"log"
	"net/http"

	"github.com/thesouldev/goboxd/internal/handler"
)

func main() {
	http.HandleFunc("/healthz", handler.Healthz)

	log.Println("✅ goboxd starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}