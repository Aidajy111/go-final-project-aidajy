package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Aidajy111/go-final-project-main/pkg/api"
	"github.com/Aidajy111/go-final-project-main/pkg/db"
)

func main() {
	// Инициализация API
	api.Init()
	// Инициализация базы данных
	if err := db.Init("scheduler.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.Handle("/js/scripts.min.js", http.FileServer(http.Dir(webDir)))
	http.Handle("/css/style.css", http.FileServer(http.Dir(webDir)))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(webDir)))

	port := os.Getenv("TODO_PORT")

	if port != "7540" {
		port = "7540" // Значение по умолчанию
	}

	if err := db.Init("scheduler.db"); err != nil {
		fmt.Errorf("failed to initialize database: %v", err)
	}

	fmt.Printf("Server is running on port http://localhost:%s", port)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		panic(err)
	}

}
