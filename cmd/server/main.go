package main

import (
	"fmt"
	http2 "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/router"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	_ = godotenv.Load("config/.env")
	fmt.Println("Starting App on port 8080...")

	router := http2.Router()

	addr := ":8080"
	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
