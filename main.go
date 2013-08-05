/*
pi-chart

A remote http-based display server for the Raspberry Pi.

Created for use as a heads-up information display on large screens attached to the Raspberry Pi. Uses OpenVG to draw highly-accelerated 2D graphics.

TODO: json/XML-based layout
TODO: Command-line flags
TODO: Variable resolution
TODO: Transitional animations
TODO: Unit tests.
*/
package main

import (
	"github.com/genbattle/openvg"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	screenWidth, screenHeight int
	submitPage                []byte
	reqChan                   chan *http.Request
)

func drawThread(req <-chan *http.Request) {
	screenWidth, screenHeight := openvg.Init()
	openvg.Start(screenWidth, screenHeight)
	log.Println("Finished OpenVG Init in drawing thread")
	var current *http.Request
	var images []*image.Image
	defer openvg.Finish() // Never gets called?
	// Poll endlessly for requests to draw
	for {
		log.Println("Drawing thread waiting for request...")
		current = <-reqChan
		// Parse the POST form
		current.ParseMultipartForm(10485760) // Parse the form with 10MB buffer
		// Get all images (files) from the form
		for i := range current.MultipartForm.File {
			for j := range current.MultipartForm.File[i] {
				img, err := extractImage(current.MultipartForm.File[i][j])
				if err != nil {
					log.Println("Error while extracting image ", i, " from POST form")
					log.Println(err)
					continue
				}
				images = append(images, &img)
			}
		}
		// Get all images (urls) from the form
		for i := range current.MultipartForm.Value {
			for j := range current.MultipartForm.Value[i] {
				img, err := downloadImage(current.MultipartForm.Value[i][j])
				if err != nil {
					log.Println("Error while downloading image from url", current.MultipartForm.Value[i][0], ", from form field", i)
					log.Println(err)
					continue
				}
				images = append(images, &img)
			}
		}

		// Download image
		log.Println("Drawing image width ", screenWidth, " height ", screenHeight)
		openvg.Start(screenWidth, screenHeight) // Start the picture
		openvg.BackgroundColor("black")         // Black background

		// Display images in a simple scanning grid
		widthCount := 0  // total row width
		heightCount := 0 // total screen height
		rowHeight := 0   // max height of row
		for i := range images {
			bounds := (*images[i]).Bounds()
			widthCount += bounds.Dx()
			if widthCount < screenWidth {
				if rowHeight < bounds.Dy() {
					rowHeight = bounds.Dy()
				}
				openvg.ImageGo(float32(widthCount-bounds.Dx()), float32(screenHeight-heightCount-bounds.Dy()), *images[i])
			} else {
				widthCount = 0
				heightCount += rowHeight
				if heightCount > screenHeight {
					log.Println("Ran out of room to print images, ignoring some")
					break
				}
			}
		}
		openvg.End()
	}
}

// Displays the POSTed data on the screen
func handlePOST(w http.ResponseWriter, r *http.Request) {
	// Post the data to the screen
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
