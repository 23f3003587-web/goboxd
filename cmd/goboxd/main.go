package main

import (
	"log"
	"net/http"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/handler"
)

func main() {
	if err := config.LoadLanguages("languages.yaml"); err != nil {
		log.Fatal("Failed to load languages:", err)
	}

	http.HandleFunc("/healthz", handler.Healthz)
	http.HandleFunc("/run", handler.Run)

	log.Println("✅ goboxd started on :8080 | Languages:", len(config.Languages))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
