package main

// This file launches a GCA server. Most of the work is being done in
// 'NewGCAServer()', the main purpose of this file is to set up OS related
// tasks such as creating the homedir for the server and listening for quit
// signals from the OS.

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/glowlabs-org/gca-backend/server"
)

// main is the entry point of the application.
func main() {
	// Get the user's home directory in an OS-agnostic manner.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error obtaining user's home directory:", err)
		os.Exit(1)
	}

	// Create the server directory path within the user's home directory.
	serverDir := filepath.Join(homeDir, "gca-server")

	// Internal test mode enables internal APIs, and sets the logging level to Info.
	var internalTestMode bool
	if len(os.Args) == 2 && os.Args[1] == "--internal-test" {
		internalTestMode = true
	}

	// Initialize a new GCAServer instance with the server directory.
	gcaServer, err := server.NewGCAServer(serverDir, internalTestMode)
	if err != nil {
		fmt.Println("Unable to launch GCA server:", err)
		os.Exit(1)
	}

	if internalTestMode {
		fmt.Println("This server is using internal test mode, and should not be used in production.")
	}

	// Create a channel to listen for operating system signals.
	// The channel c is buffered with a size of 1.
	c := make(chan os.Signal, 1)

	// Notify the channel c upon receiving either an Interrupt signal or a SIGTERM signal.
	// This helps us gracefully shut down the application.
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Goroutine that waits for an Interrupt or SIGTERM signal.
	// It will call Close() on the GCAServer instance and then exit the program.
	go func() {
		// Block until a signal is received.
		<-c

		// Begin shutdown. Count in the terminal how long shutdown is
		// taking. The goroutine that counts how long shutdown is
		// taking will automatically be killed when os.Exit is called,
		// there's no need to clean up that loop.
		fmt.Println("Close signal received, shutting down server. ETA 90 seconds.")
		go func() {
			times := 0
			for {
				time.Sleep(time.Second * 5)
				times++
				fmt.Println(times*5, "seconds")
			}
		}()
		gcaServer.Close() // Close the GCAServer.
		fmt.Println()     // Print a newline for cleaner terminal output.
		os.Exit(0)        // Exit the program with a successful status code.
	}()

	// An empty select block is used to keep the main function alive indefinitely.
	// This is necessary because the main function would exit otherwise, killing any child goroutines.
	select {} // Block forever.
}
