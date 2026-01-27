package main

import (
	"app/internal"
	"log/slog"
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
	slog.Info("Application started successfully")

	// Handle OS signals for graceful shutdown.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan
	slog.Info("Received shutdown signal")

	// Stop the application gracefully.
	application.Stop()
	slog.Info("Application stopped successfully")
}
