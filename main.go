package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	u := url.URL{
		Scheme: "wss",
		Host: "api.hyperliquid.xyz",
		Path: "/ws",
	}

	fmt.Println(u.String())

	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Error dialing WebSocket:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to WebSocket server")

	// Start a goroutine to read messages from the WebSocket
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failed to read message:", err)
				return
			}
			fmt.Printf("Data masuk: %s\n\n", string(message)) // output akan berupa JSON string (TODO: bisa diubah ke struct)
		}
	}()

	// Send a subscription message to the server
	subscribePayload := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "trades",
			"coin": "BTC",
		},
	}

	// Marshal the subscription payload to JSON
	jsonPayload, err := json.Marshal(subscribePayload)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, jsonPayload)
	if err != nil {
		fmt.Println("Error sending subscription message:", err)
		return
	}
	fmt.Println("Subscription message sent")

	// Wait for an interrupt signal to gracefully shut down
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println("Waiting for messages...")

	// Set a read deadline to avoid blocking indefinitely
	err = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	if err != nil {
		fmt.Println("Error setting read deadline:", err)
		return
	}
	time.Sleep(1 * time.Second)
}