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
		port = "8095"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"delivery-webhook"}`)
	})

	http.HandleFunc("/webhook/grab", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"accepted","message":"Grab webhook received"}`)
	})

	http.HandleFunc("/webhook/lineman", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"accepted","message":"LineMan webhook received"}`)
	})

	log.Printf("Starting delivery webhook service on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
