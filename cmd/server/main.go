package main

import (
	"fmt"
	http2 "github.com/GOVSEteam/strv-vse-go-newsletter/internal/transport/http"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting Newsletter Service...")

	router := http2.NewRouter()

	addr := ":8080"
	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
