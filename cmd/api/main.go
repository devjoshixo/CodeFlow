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
		port = "8080"
	}
	log.Printf("Port is: %s", port)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status": "ok"}`)
	})

	log.Printf("Server started on the port localhost:%s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %s", err)
	}

}
