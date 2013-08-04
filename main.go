/*
pi-chart

A remote http-based display server for the Raspberry Pi.

Created for use as a heads-up information display on large screens attached to the Raspberry Pi. Uses OpenVG to draw highly-accelerated 2D graphics.

TODO: json/XML-based layout
TODO: Command-line flags
TODO: Variable resolution
TODO: Transitional animations
*/
package main

import (
	"github.com/genbattle/openvg"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"errors"
)

var (
	screenWidth, screenHeight int
	submitPage                []byte
	reqChan                   chan *http.Request
)

// Download and decode an image from a url
func downloadImage(url string) (image.Image, error) {
	return nil, nil
}

// Extract and decode an image submitted as part of a form POST
func extractImage(req *http.Request) (image.Image, error) {
	file, header, err := req.FormFile("imagefile")
	if err != nil {
		log.Println("Error while getting form file from request")
		return nil, err
	}
	// Check file MIME type
	var img image.Image
	switch header.Header["Content-Type"][0] {
	case "image/jpeg", "image/png":
		img, _, err = image.Decode(file)
		if err != nil {
			log.Println("Error while decoding request image data")
			return nil, err
		}
	default:
		log.Println("Unsupported image format ", header.Header["Content-Type"])
		return nil, errors.New("Unsupported image format")
	}
	return img, nil
}

func drawThread(req <-chan *http.Request) {
	screenWidth, screenHeight := openvg.Init()
	openvg.Start(screenWidth, screenHeight)
	log.Println("Finished OpenVG Init in drawing thread")
	var current *http.Request
	defer openvg.Finish() // Never gets called?
	// Poll endlessly for requests to draw
	for {
		log.Println("Drawing thread waiting for request...")
		current = <-reqChan
		// Choose what to do based on what form fields are populated
		current.ParseMultipartForm(10485760) // Parse the form with 10MB buffer
		log.Println(len(current.MultipartForm.Value))
		log.Println(len(current.MultipartForm.File))

		// Extract image from request
		img, err := extractImage(current)
		if err != nil {
			log.Println("Error while extracting image from POST form")
			log.Fatal(err)
		}
		// Download image
		log.Println("Drawing image width ", screenWidth, " height ", screenHeight)
		openvg.Start(screenWidth, screenHeight) // Start the picture
		openvg.BackgroundColor("black")         // Black background
		openvg.FillRGB(44, 100, 232, 1)         // Big blue marble
		openvg.FillColor("white")               // White text
		openvg.ImageGo(100, 100, img)
		// openvg.TextMid(float32(screenWidth / 2), float32(screenHeight / 2), "hello, world", "serif", screenWidth/10) // Greetings
		openvg.End()
	}
}

// Displays the POSTed data on the screen
func handlePOST(w http.ResponseWriter, r *http.Request) {
	// Figure out what sort of data we're dealing with
	// TODO: Check MIME type?
	// Post the data to the screen
	// TODO: Replace generic hell world code with something useful
	log.Println("Sending request to drawing thread")
	reqChan <- r
	handleGET(w, r)
}

// Check the HTTP request method
func handleGET(w http.ResponseWriter, r *http.Request) {
	w.Write(submitPage)
	log.Println("Wrote Response length ", len(submitPage))
}

// Check the HTTP request method
func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST", "PUT":
		handlePOST(w, r)
	default:
		// TODO: Respond with page that allows user to send images or data to the server
		handleGET(w, r)
	}
}

func main() {
	// TODO: Set up command-line arguments for port, etc.
	// Load default response page
	log.Println("Starting up...")
	file, err := os.Open("submit.html")
	if err != nil {
		log.Println("Error while loading default response page")
		log.Fatal(err)
	}
	submitPage, err = ioutil.ReadAll(file)
	if err != nil {
		log.Println("Error while reading default response page from disk")
		log.Fatal(err)
	}
	// Set up the OpenVG rendering
	reqChan = make(chan *http.Request)
	go drawThread(reqChan)
	// Create default handler
	http.HandleFunc("/", handle)
	log.Println("Finished init, listening...")
	// Start HTTP server listening on port 8787
	err = http.ListenAndServe(":8787", nil)
	if err != nil {
		log.Println("Error returned by HTTP server:")
		log.Fatal(err)
	}
}
