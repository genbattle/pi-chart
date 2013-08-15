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
	"encoding/json"
	"math"
)

var (
	screenWidth, screenHeight int
	submitPage                []byte
	reqChan                   chan *http.Request // TODO: Move this to main?
)

type Layout struct {
	Rows, Cols int // Total number of rows/columns to divide the screen into
	Images []ImageLayout
	Graphs []GraphLayout
}

type ImageLayout struct {
	FileName string
	Scale string // scaling mode; stretch, zoom, shrink
	// TODO: z-index to allow overlapping images/graphs
	Left, Right int
	Width, Height int
}

type GraphLayout struct {
	Data []float64
	Left, Right int
	Width, Height int
}

// Stores the calculated properties of the layout grid
type Grid struct {
	Rows, Cols int
	ColWidth, RowHeight int
	TotalWidth, TotalHeight int
}

// Parse the JSON layout/graph data
func parseLayout(layoutJSON string) (*Layout, error) {
	if len(layoutJSON) == 0 {
		// Empty layout field, not an error, just no layout supplied
		log.Println("Warning: No layout supplied")
		return nil, nil
	}
	layout := new(Layout)
	err := json.Unmarshal([]byte(layoutJSON), layout)
	if err != nil {
		log.Println("Error while unmarshaling the JSON layout data")
		return nil, err
	}
	return layout, nil
}

// Draws an image with the top left corner at the top left corner of the (row, col) cell of the grid. Resizes the image to be width cells wide and height cells high.
func gridDrawImage(grid *Grid, img *image.Image, row, col, width, height int) {
	// Calculate screen x,y for image, remembering that OpenVG's origin is bottom left, not top-left
	x := float32(col * (*grid).ColWidth)
	y := float32(grid.TotalHeight - (row * (*grid).RowHeight) - (*img).Bounds().Dy())
	pixwidth := float32(width * (*grid).ColWidth)
	pixheight := float32(height * (*grid).RowHeight)
	// Calculate scale factor from pixel values
	scalewidth := pixwidth / float32((*img).Bounds().Dx())
	scaleheight := pixheight / float32((*img).Bounds().Dy())
	log.Println(scalewidth, scaleheight)
	// Draw the image
	openvg.Scale(scalewidth, scaleheight)
	openvg.ImageGo(x, y, *img)
}

func newGrid(totalWidth, totalHeight, rows, cols int) *Grid {
	grid := new(Grid)
	grid.Rows = rows
	grid.Cols = cols
	grid.ColWidth = totalWidth / cols
	grid.RowHeight = totalWidth / rows
	grid.TotalWidth = totalWidth
	grid.TotalHeight = totalHeight
	return grid
}

// Main internal drawing thread; all graphics calls have to be on the one thread, as exepected by OpenVG.
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
		key := "imagefile"
		for j := range current.MultipartForm.File[key] {
			img, err := extractImage(current.MultipartForm.File[key][j])
			if err != nil {
				log.Println("Error while extracting image ", j, " from POST form")
				log.Println(err)
				continue
			}
			images = append(images, &img)
		}
		// Get all images (urls) from the form
		key = "imageurl"
		for j := range current.MultipartForm.Value[key] {
			if current.MultipartForm.Value[key][j] != "" {
				img, err := downloadImage(current.MultipartForm.Value[key][j])
				if err != nil {
					log.Println("Error while downloading image from url", current.MultipartForm.Value[key][j])
					log.Println(err)
					continue
				}
				images = append(images, &img)
			}
		}
		// Get only the _first_ layout object
		key = "layout"
		if len(current.MultipartForm.Value[key]) != 0 {
			layout, err := parseLayout(current.MultipartForm.Value[key][0])
			if err != nil {
				log.Println("Error while parsing layout")
				log.Println(err)
			}
			log.Println(layout) // TODO: do something with layout
		}

		log.Println("Drawing image width ", screenWidth, " height ", screenHeight)
		openvg.Start(screenWidth, screenHeight) // Start the picture
		openvg.BackgroundColor("black")         // Black background

		// Display images in default grid
		// Can't be any graphs in this case - graphs require layout. TODO: Separate graphs from layout?
		// Calculate grid size based on number of images/graphs
		objcount := len(images) // TODO: Add graph count here
		if objcount == 0 {
			// nothing to display skip to next request
			continue
		}
		dim := int(math.Ceil(math.Sqrt(float64(objcount))))
		grid := newGrid(screenWidth, screenHeight, dim, dim)
		row := 0
		col := 0
		// display images sequentially in the grid
		for i := range(images) {
			gridDrawImage(grid, images[i], row, col, 1, 1)
			col++
			if col > dim {
				row++ // row should never be grater than dim, because of how we calculate dim
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
