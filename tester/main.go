package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	for x := 0; x < 40; x++ {
		for y := 0; y < 40; y++ {
			resp, err := http.Get(fmt.Sprintf("http://localhost:1234/draw?x=%d&y=%d&color=0f7733&layer=leon", x, y))
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
