package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// loop over alle posities op het canvas
	for x := 0; x < 40; x++ {
		for y := 0; y < 40; y++ {
			// doe http Get request
			resp, err := http.Get(fmt.Sprintf("http://192.168.1.85:1234/draw?x=%d&y=%d&color=00ff00&layer=tester", x, y))
			if err != nil {
				log.Fatalf("error drawing with http.Get: %s\n", err)
			}
			// controlleer de statusCode
			if resp.StatusCode != http.StatusOK {
				log.Fatalf("error drawing, have http status %s\n", resp.Status)
			}
			// sluit de response body zodat de onderliggende TCP connectie hergebruikt kan worden
			resp.Body.Close()
		}
	}
}
