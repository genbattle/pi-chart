package main

import (
	"github.com/genbattle/openvg"
	"log"
	"net/http"
)

var screenWidth, screenHeight int

// Displays the POSTed data on the screen
func handlePOST(w http.ResponseWriter, r *http.Request) {
	// Figure out what sort of data we're dealing with
	// TODO: Check MIME type?
	// Post the data to the screen
	// TODO: Replace generic hell world code with something useful
	openvg.Start(width, height)                               // Start the picture
	openvg.BackgroundColor("black")                           // Black background
	openvg.FillRGB(44, 100, 232, 1)                           // Big blue marble
	openvg.Circle(w2, 0, w)                                   // The "world"
	openvg.FillColor("white")                                 // White text
	openvg.TextMid(w2, h2, "hello, world", "serif", width/10) // Greetings
	openvg.End()
}

// Check the HTTP request method
func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case POST:
		handlePOST(w, r)
	default:
		log.Println("Receieved non-POST ", r.Method, " Request, ignoring")
	}
}

func main() {
	// TODO: Set up command-line arguments for port, etc.
	// Set up the OpenVG rendering
	screenWidth, screenHeight := openvg.Init()
	openvg.Start(screenWidth, screenHeight)
	defer openvg.Finish()
	// Create default handler
	http.HandleFunc("", handle)
	// Start HTTP server listening on port 8787
	err := http.ListenAndServe(":8787", nil)
	if err != nil {
		log.Println("Error returned by HTTP server:")
		log.Fatal(err)
	}
}
