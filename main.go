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
	"errors"
	"github.com/genbattle/openvg"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"mime/multipart"
)

var (
	screenWidth, screenHeight int
	submitPage                []byte
	reqChan                   chan *http.Request
)

// Download and decode an image from a url
func downloadImage(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		log.Println("Failed to fetch image from URL ", url)
		return nil, err
	}
	// read the image data from the response
	img, _, err := image.Decode(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Println("Error while decoding image from response body")
		return nil, err
	}
	return img, nil
}

// Extract and decode an image submitted as part of a form POST
func extractImage(header *multipart.FileHeader) (image.Image, error) {
	file, err := header.Open()
	if err != nil {
		log.Println("Error while getting form file from header")
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
			img, err := extractImage(current.MultipartForm.File[i][0]) //TODO: should we check for more than one file header per key here?
			if err != nil {
				log.Println("Error while extracting image ", i, " from POST form")
				log.Println(err)
				continue
			}
			images = append(images, &img)
		}
		// Get all images (urls) from the form
		for i := range current.MultipartForm.Value {
			img, err := downloadImage(current.MultipartForm.Value[i][0]) //TODO: should we check for more than one url per key here?
			if err != nil {
				log.Println("Error while downloading image from url", current.MultipartForm.Value[i][0], ", from form field", i)
				log.Println(err)
				continue
			}
			images = append(images, &img)
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
				openvg.ImageGo(float32(widthCount-bounds.Dx()), float32(screenHeight-heightCount-bounds.Dy()),*images[i])
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
