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

type TradeData struct {
	Coin string `json:"coin"`
	Side string `json:"side"`
	Px string `json:"px"`
	Sz string `json:"sz"`
	Time int64 `json:"time"`
}

type WSResponse struct {
	Channel string `json:"channel"`
	Data json.RawMessage `json:"data"`
}

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

	fmt.Printf("%-20s | %-10s | %-10s | %-10s | %-20s\n", "Coin", "Side", "Price", "Size (USDC)", "Time")
	// Start a goroutine to read messages from the WebSocket
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failed to read message:", err)
				return
			}
			
			// Unmarshal the message into a WSResponse struct
			var response WSResponse
			err = json.Unmarshal(message, &response)
			if err != nil {
				log.Println("Failed to unmarshal message:", err)
				continue
			}

			// Check if the channel is "trades" and unmarshal the data into a slice of TradeData
			if response.Channel == "trades" {
				var trades []TradeData
				err = json.Unmarshal(response.Data, &trades)

				if err != nil {
					log.Println("Failed to unmarshal trade data:", err)
					continue
				}

				for _, trade := range trades {
					t := time.UnixMilli(trade.Time).Format("15:04:05.000")
					fmt.Printf("%-20s | %-10s | %-10s | %-10s | %-20s\n", trade.Coin, trade.Side, trade.Px, trade.Sz, t)
				}
			}
		}
	}()

	// Send a subscription message to the server
	subscribePayload := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "trades",
			"coin": "BTC", // TODO: make this dynamic based on user input or configuration (FUTURE WORK)
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
}