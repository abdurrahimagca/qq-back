package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting server...")
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello, This is working World!"})
	})

	log.Println("Server is running on port 3003")
	log.Fatal(http.ListenAndServe(":3003", mux))
}
