package main

import (
	"context"
	"fmt"
	"go-trello/routes"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
)

func main() {
	mainRouter := mux.NewRouter()

	// Get the API routes
	apiRouter := routes.ApiRoutes()

	// Get the web routes
	webRouter := routes.WebRoutes()

	ctx, cancel := context.WithCancel(context.Background())

	// Create a channel to listen for the termination signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Register the API routes under /api
	mainRouter.PathPrefix("/api").Handler(http.StripPrefix("/api", apiRouter))

	// Define the route to stop the server
	mainRouter.PathPrefix("/stop-server").Handler(http.StripPrefix("/stop-server", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is shutting down...")
		cancel()
	})))

	box := packr.New("MyBox", "./resources")
	mainRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))

	// Register the web routes under /
	mainRouter.PathPrefix("/").Handler(webRouter)

	// pop box here
	// a := app.New()
	// w := a.NewWindow("Hello World")

	// w.SetContent(widget.NewLabel("Hello World!"))
	// w.ShowAndRun()

	// Create the HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mainRouter,
	}

	// Start the server in a goroutine
	go func() {
		url := "http://localhost:8080"
		fmt.Println("Starting server on " + url)
		// errUrl := openURL(url)
		// if errUrl != nil {
		// 	fmt.Println("Server error:", errUrl)
		// }
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Println("Server error:", err)
		}
	}()

	// Wait for the termination signal or cancellation of the context
	select {
	case <-stop:
		fmt.Println("Termination signal received, shutting down...")
	case <-ctx.Done():
		fmt.Println("Context canceled, shutting down...")
	}

	// Allow a grace period for outstanding requests to finish
	gracefulShutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	// Shutdown the server gracefully
	err := server.Shutdown(gracefulShutdownCtx)
	if err != nil {
		fmt.Println("Error during server shutdown:", err)
	}
}

func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		// macOS
		cmd = exec.Command("open", url)
	case "windows":
		// Windows
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		// Linux and other Unix-like systems
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		return err
	}

	return nil
}
