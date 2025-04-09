package main

import (
	"app/internal"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	application := internal.NewApp()

	// Start the application in a separate goroutine.
	go func() {
		application.Start()
	}()
	log.Println("Application started successfully")

	// Handle OS signals for graceful shutdown.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan
	log.Println("Received shutdown signal")

	// Stop the application gracefully.
	application.Stop()
	log.Println("Application stopped successfully")
}
