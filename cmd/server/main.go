package main

import (
	"net/http"
	"pg-access-bot/internal/handler/access"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Post("/request-access", access.RequestAccess)

	http.ListenAndServe(":8080", r)
}
