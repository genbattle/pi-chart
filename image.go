package main

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"mime/multipart"
	"net/http"
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
