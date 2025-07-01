package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8096"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"payment-webhook"}`)
	})

	http.HandleFunc("/webhook/omise", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"accepted","message":"Omise webhook received"}`)
	})

	http.HandleFunc("/webhook/2c2p", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"accepted","message":"2C2P webhook received"}`)
	})

	log.Printf("Starting payment webhook service on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
