package main

import (
	"fmt"
	"image/gif"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	file, err := os.Open("pacman.gif")
	if err != nil {
		log.Fatalf("Error opening file: %s\n", err)
	}
	g, err := gif.DecodeAll(file)
	if err != nil {
		log.Fatalf("Error decoding gif: %s\n", err)
	}
	for {
		for _, img := range g.Image {
			for x := 0; x < 40; x++ {
				for y := 0; y < 40; y++ {
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
			time.Sleep(300 * time.Millisecond)
		}
	}
}
