package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	databaseURL, err := lookupDatabaseURL()
	if err != nil {
		log.Fatal(err)
	}

	repo, err := newPostgresRepository(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	if err := repo.EnsureSchema(context.Background()); err != nil {
		log.Fatal(err)
	}
	if err := repo.SeedInitialData(context.Background()); err != nil {
		log.Fatal(err)
	}

	rankingService := NewRankingService(repo)
	playerService := NewPlayerService(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest(rankingService, playerService))

	addr := "127.0.0.1:8000"
	log.Printf("Server running on http://%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
