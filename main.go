package main

import (
	"encoding/json"
	"fmt"
	"github.com/GeertJohan/go.incremental"
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var regexpColor = regexp.MustCompile(`^([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

type canvas struct {
	width  uint64
	height uint64

	layers     map[string]*layer
	layersLock sync.RWMutex

	lock           sync.RWMutex // need writelock for consistent snapshot and subscriber map update
	subscribers    map[uint]chan *canvasUpdate
	subscribersInc incremental.Uint
}

type snapshot struct {
	Width           uint64
	Height          uint64
	SinglelayerMode bool
	Layers          map[string]*layer
}

type canvasUpdate struct {
	LayerName     string
	PixelPosition uint64
	PixelColor    string
}

type layer struct {
	Name       string
	Pixels     pixelMap
	pixelsLock sync.Mutex
}

type pixelMap map[uint64]*pixel

func (pm pixelMap) MarshalJSON() ([]byte, error) {
	type smashedPixel struct {
		Position  uint64
		Timestamp int64
		Color     string
	}
	var smashedPixels = make([]smashedPixel, 0, len(pm))
	for pos, pix := range pm {
		smashedPixels = append(smashedPixels, smashedPixel{
			Position:  pos,
			Timestamp: pix.timestamp,
			Color:     pix.color,
		})
	}
	return json.Marshal(smashedPixels)
}

type pixel struct {
	timestamp int64
	color     string
}

func (c *canvas) getLayer(name string) *layer {
	// get readlock and find layer
	c.layersLock.RLock()
	l, exists := c.layers[name]
	c.layersLock.RUnlock()

	// if layer doesn't exist, we might need to make it
	if !exists {
		// layer does not exist, get write lock
		c.layersLock.Lock()
		// check again if layer doesn't exist (avoid race)
		l, exists = c.layers[name]
		if !exists {
			// still doesn't exist with write-lock, now make layer
			l = &layer{
				Name:   name,
				Pixels: make(map[uint64]*pixel),
			}
			c.layers[name] = l
		}
		// write unlock
		c.layersLock.Unlock()
	}

	// return layer
	return l
}

func (c *canvas) socketHandler(w http.ResponseWriter, r *http.Request) {
	layerName := r.FormValue("layer")

	conn, err := websocket.Upgrade(w, r, nil, 10240, 10240)
	if err != nil {
		log.Printf("Error upgrading websocket: %s\n", err)
	}

	c.lock.Lock()
	snapshot := &snapshot{
		Width:  c.width,
		Height: c.height,
	}
	if len(layerName) > 0 {
		snapshot.SinglelayerMode = true
		snapshot.Layers = make(map[string]*layer)
		l, exists := c.layers[layerName]
		if exists {
			snapshot.Layers[layerName] = l
		}
	} else {
		snapshot.Layers = c.layers
	}
	err = conn.WriteJSON(snapshot)
	if err != nil {
		log.Printf("Error sending snapshot: %s\n", err)
		c.lock.Unlock()
		return
	}
	subscriberID := c.subscribersInc.Next()
	subscriberCh := make(chan *canvasUpdate, 100) // buffered, no need for sync, just be asap
	c.subscribers[subscriberID] = subscriberCh
	c.lock.Unlock()

	// cleanup subscription
	defer func() {
		c.lock.Lock()
		delete(c.subscribers, subscriberID)
		c.lock.Unlock()
	}()

	for {
		update := <-subscriberCh
		if len(layerName) > 0 && update.LayerName != layerName {
			continue // skip layers that are not the requested layer
		}
		err = conn.WriteJSON(update)
		if err != nil {
			log.Printf("Error sending udpate: %s\n", err)
			return
		}
	}
}

func (c *canvas) drawHandler(w http.ResponseWriter, r *http.Request) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// get layer name
	layerName := r.FormValue("layer")
	if len(layerName) == 0 {
		http.Error(w, "field 'layer' is missing", http.StatusBadRequest)
		return
	}

	xStr := r.FormValue("x")
	if len(xStr) == 0 {
		http.Error(w, "field 'x' is missing", http.StatusBadRequest)
		return
	}
	x, err := strconv.ParseUint(xStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("'x' is invalid: %s", err), http.StatusBadRequest)
		return
	}
	if x >= c.width {
		http.Error(w, "'x' is out of range", http.StatusBadRequest)
		return
	}

	yStr := r.FormValue("y")
	if len(yStr) == 0 {
		http.Error(w, "field 'y' is missing", http.StatusBadRequest)
		return
	}
	y, err := strconv.ParseUint(yStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("'y' is invalid: %s", err), http.StatusBadRequest)
		return
	}
	if y >= c.height {
		http.Error(w, "'y' is out of range", http.StatusBadRequest)
		return
	}

	// calculate position
	position := y*c.width + x

	// get layer and lock pixmap
	layer := c.getLayer(layerName)

	// prep canvasUpdate
	update := &canvasUpdate{
		LayerName: layerName,
	}

	// get and verify color
	color := r.FormValue("color")
	if !regexpColor.MatchString(color) {
		http.Error(w, "'color' is invalid, must be HEX color code without hash, e.g. `FF0000`", http.StatusBadRequest)
		return
	}
	layer.Pixels[position] = &pixel{
		timestamp: time.Now().Unix(),
		color:     color,
	}

	// set fields on update
	update.PixelPosition = position
	update.PixelColor = color

	// send update to subs
	c.sendUpdate(update)

	// all done
	io.WriteString(w, "done!")
}

func (c *canvas) sendUpdate(update *canvasUpdate) {
	for subID, sub := range c.subscribers {
		select {
		case sub <- update:
			// successfull sent or buffer
		default:
			log.Printf("dropped update for subscriber %d\n", subID)
		}
	}
}

func main() {

	// create canvas
	canvas := &canvas{
		width:       40,
		height:      40,
		layers:      make(map[string]*layer),
		subscribers: make(map[uint]chan *canvasUpdate),
	}

	// register http handlers
	http.HandleFunc("/socket", canvas.socketHandler)
	http.HandleFunc("/draw", canvas.drawHandler)
	http.Handle("/", http.FileServer(rice.MustFindBox("http-files").HTTPBox()))

	server := &http.Server{
		Addr:           ":1234",
		Handler:        http.DefaultServeMux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error listenAndServe: %s\n", err)
	}
}
