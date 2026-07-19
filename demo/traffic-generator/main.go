package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	target := os.Getenv("CHECKOUT_URL")
	if target == "" {
		target = "http://checkout:8081/checkout?items=8"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		response, err := client.Get(target)
		if err != nil {
			log.Printf("checkout request: %v", err)
			continue
		}
		response.Body.Close()
	}
}
