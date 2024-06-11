package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

var clients = make(map[*websocket.Conn]bool)
var mu sync.Mutex

func main() {
	http.HandleFunc("/media", mediaHandler)
	http.HandleFunc("/playback", playbackHandler)
	fmt.Println("Server started at port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func mediaHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Could not establish WebSocket connection:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	fmt.Println("WebSocket connection established")

	var totalLength int
	var mediaData []byte

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message from client:", err)
			break
		}

		if string(message) == "StreamingStopped" {
			fmt.Println("Streaming stopped by client")
			break
		}

		length := len(message)
		totalLength += length
		mediaData = append(mediaData, message...)
	}

	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
	fmt.Println("Total length of media data received:", totalLength)

	// Save media data to file
	err = os.WriteFile("media_data.raw", mediaData, 0644)
	if err != nil {
		fmt.Println("Error saving media data to file:", err)
	}
}

func playbackHandler(w http.ResponseWriter, r *http.Request) {
	mediaData, err := os.ReadFile("media_data.raw")
	if err != nil {
		http.Error(w, "Could not read media file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/webm") // Assuming the media data is in webm format
	w.Write(mediaData)
}
