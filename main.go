package main

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/sammyshear/lcaaj-transcriber/views"
)

func main() {
	mux := chi.NewMux()

	mux.Handle("/", templ.Handler(views.IndexPage()))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
