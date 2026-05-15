package main

import (
	"codeflow/internal/platform/migrator"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Port is: %s", port)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status": "ok"}`)
	})

	log.Printf("Server started on the port localhost:%s", port)

	if err := migrator.Run(os.Getenv("DATABASE_URL")); err != nil {
		fmt.Println("Got error in migration")
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %s", err)
	}

}
