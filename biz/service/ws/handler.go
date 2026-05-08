package ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var (
	wsServer   *http.Server
	wsListener net.Listener
	wsMu       sync.Mutex
	wsStarted  bool
)

func StartWSServer(addr string) error {
	wsMu.Lock()
	defer wsMu.Unlock()
	if wsStarted {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ws/reviews", handleWebSocket)

	wsServer = &http.Server{Handler: mux}

	var err error
	wsListener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	wsStarted = true
	go func() {
		if err := wsServer.Serve(wsListener); err != nil && err != http.ErrServerClosed {
			log.Printf("[WS] Server error: %v", err)
		}
	}()

	log.Printf("[WS] WebSocket server listening on %s", addr)
	return nil
}

func StopWSServer() {
	wsMu.Lock()
	defer wsMu.Unlock()
	if wsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		wsServer.Shutdown(ctx)
		wsStarted = false
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	id := generateClientID()
	client := &Client{
		ID:     id,
		Hub:    DefaultHub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Topics: map[string]bool{"review": true},
	}

	DefaultHub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

func generateClientID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b) + "-" + time.Now().Format("150405")
}
