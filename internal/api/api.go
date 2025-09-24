package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	return r
}

func startGame(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Start game -- not implemented yet"))
}
