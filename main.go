package main

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/sammyshear/lcaaj-transcriber/internal"
	"github.com/sammyshear/lcaaj-transcriber/views"
)

func main() {
	mux := chi.NewMux()

	mux.Handle("/", templ.Handler(views.IndexPage()))
	mux.Post("/api/transcribe", internal.ApiTranscribe)
	mux.HandleFunc("/api/dtranscribe", internal.DatastarTranscribe)

	handler := templ.NewCSSMiddleware(mux, views.MainClass(), views.InputClass())

	log.Fatal(http.ListenAndServe(":8080", handler))
}
