package main

import (
	"fmt"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	file, err := os.Open("logo.jpg")
	if err != nil {
		log.Fatalf("error opening file: %s\n", err)
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatalf("error decoding file: %s\n", err)
	}
	for x := 0; x < 40; x++ {
		for y := 0; y < 15; y++ {
			color := img.At(x, y)
			r, g, b, _ := color.RGBA()
			colorString := fmt.Sprintf("%02x%02x%02x", uint8(r), uint8(g), uint8(b))
			fmt.Println(colorString)
			resp, err := http.Get(fmt.Sprintf("http://localhost:1234/draw?x=%d&y=%d&color=%s&layer=foize", x, y, colorString))
			if err != nil {
				log.Fatalf("error drawing with http.Get: %s\n", err)
			}
			if resp.StatusCode != http.StatusOK {
				log.Fatalf("error drawing, have http status %s\n", resp.Status)
			}
			resp.Body.Close()
		}
	}
}
