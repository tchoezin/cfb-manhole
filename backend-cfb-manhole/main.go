package main

import (
	"log"
	"net/http"
)

func main() {
	service := NewRankingService(scoringRepository{})

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest(service))

	addr := "127.0.0.1:8000"
	log.Printf("Server running on http://%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
