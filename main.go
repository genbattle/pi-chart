package main

import (
	"github.com/genbattle/openvg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	screenWidth, screenHeight int
	submitPage                []byte
	reqChan chan *http.Request
)

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
		log.Println(current)
		// Extract image from request
		/*file, header, err := current.FormFile("image-file")
		if err != nil {
			log.Println("Error while getting form file from request")
			log.Fatal(err)
		}*/
		log.Println("Drawing image width ", screenWidth, " height ", screenHeight)
		openvg.Start(screenWidth, screenHeight)                               // Start the picture
		openvg.BackgroundColor("black")                           // Black background
		openvg.FillRGB(44, 100, 232, 1)                           // Big blue marble
		openvg.FillColor("white")                                 // White text
		//openvg.ImageGo(0, 0, screenWidth, screenHeight, img)
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
