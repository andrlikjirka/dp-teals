package main

import (
	"github.com/andrlikjira/dp-teals/internal/server"
	"github.com/andrlikjirka/logger"
	"github.com/go-chi/chi/v5"
)

func main() {
	log := logger.New("")
	router := chi.NewRouter()

	config := server.MustLoadConfig("../.env")

	server := server.New(config, log, router)

	if err := server.Run(); err != nil {
		log.Error("Server stopped", "error", err)
	}
}
