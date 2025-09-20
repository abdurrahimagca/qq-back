package main

import (
	"log"

	"github.com/abdurrahimagca/qq-back/internal/bootstrap"
	"github.com/abdurrahimagca/qq-back/internal/environment"
)

func main() {
	log.Println("Starting server...")

	environment, err := environment.Load()
	if err != nil {
		log.Fatal("Error loading environment", err)
	}

	app := bootstrap.New(environment)
	app.Bootstrap()
	app.StartServer()
}
