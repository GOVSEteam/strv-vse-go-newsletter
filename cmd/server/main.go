package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/router"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file. Errors are ignored, so if it's not present, app will rely on actual env vars.
	_ = godotenv.Load()

	fmt.Println("Starting App on port 8080...")

	// Initialize all dependencies and setup routes
	appRouter := router.Router() // Corrected to use the existing Router function

	addr := ":8080"
	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, appRouter); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
