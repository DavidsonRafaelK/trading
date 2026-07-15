package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	Reset  = "\033[0m"
	Green  = "\033[32m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
)

type L2BookReq struct {
	Type string `json:"type"`
	Coin string `json:"coin"`
}

type BookLevel struct {
	Px string `json:"px"`
	Sz string `json:"sz"`
	N  int    `json:"n"`
}

type L2BookResp struct {
	Coin   string        `json:"coin"`
	Time   int64         `json:"time"`
	Levels [][]BookLevel `json:"levels"`
}

type TradeData struct {
	Coin string `json:"coin"`
	Side string `json:"side"`
	Px   string `json:"px"`
	Sz   string `json:"sz"`
	Time int64  `json:"time"`
}

type WSResponse struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

func main() {
	coin := "BTC"
	pollingInterval := 400 * time.Millisecond

	var trades []TradeData
	var mu sync.Mutex

	// Connect to the WebSocket server
	u := url.URL{
		Scheme: "wss",
		Host:   "api.hyperliquid.xyz",
		Path:   "/ws",
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Error dialing WebSocket:", err)
		return
	}
	defer conn.Close()

	// Send a trade subscription message to the server
	subscribePayload := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "trades",
			"coin": coin,
		},
	}

	tradePayload, err := json.Marshal(subscribePayload)
	if err != nil {
		fmt.Println("Error marshalling trade subscription:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, tradePayload)
	if err != nil {
		fmt.Println("Error sending trade subscription:", err)
		return
	}

	// Start a goroutine to read trade messages from the WebSocket
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// Unmarshal the message into a WSResponse struct
			var response WSResponse
			err = json.Unmarshal(message, &response)
			if err != nil {
				continue
			}

			// Check if the channel is "trades"
			if response.Channel == "trades" {
				var newTrades []TradeData
				err = json.Unmarshal(response.Data, &newTrades)

				if err != nil {
					continue
				}

				mu.Lock()

				// Add new trades
				trades = append(trades, newTrades...)

				// Keep only trades from the last 5 seconds
				cutoff := time.Now().Add(-5 * time.Second).UnixMilli()

				for len(trades) > 0 && trades[0].Time < cutoff {
					trades = trades[1:]
				}

				mu.Unlock()
			}
		}
	}()

	// Create an HTTP client for the order book request
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	// Create the order book request payload
	payload := L2BookReq{
		Type: "l2Book",
		Coin: coin,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Wait for an interrupt signal to gracefully shut down
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Clear the terminal
	fmt.Print("\033[2J")

	for {
		select {
		case <-interrupt:
			fmt.Print("\033[0m\033[2J\033[H")
			fmt.Println("[+] Order Book dihentikan.")
			return

		default:
		}

		// Send a request to get the order book
		resp, err := client.Post(
			"https://api.hyperliquid.xyz/info",
			"application/json",
			bytes.NewBuffer(jsonPayload),
		)

		if err != nil {
			time.Sleep(pollingInterval)
			continue
		}

		// Decode the order book response
		var book L2BookResp
		err = json.NewDecoder(resp.Body).Decode(&book)
		resp.Body.Close()

		if err != nil {
			time.Sleep(pollingInterval)
			continue
		}

		// Check if bids and asks are available
		if len(book.Levels) < 2 ||
			len(book.Levels[0]) == 0 ||
			len(book.Levels[1]) == 0 {

			time.Sleep(pollingInterval)
			continue
		}

		// Move cursor to top left
		fmt.Print("\033[H")

		fmt.Printf("\033[1;36m Order Book \033[0m\n")
		fmt.Println("--------------------------------------------------")
		fmt.Printf(
			"%-15s %15s %16s\n",
			"Price",
			"Size ("+book.Coin+")",
			"Total ("+book.Coin+")",
		)
		fmt.Println("--------------------------------------------------")

		maxRows := 10

		// Process asks
		nAsks := len(book.Levels[1])
		if nAsks > maxRows {
			nAsks = maxRows
		}

		askTotals := make([]float64, nAsks)
		currentAskTotal := 0.0

		for i := 0; i < nAsks; i++ {
			size, err := strconv.ParseFloat(book.Levels[1][i].Sz, 64)
			if err != nil {
				continue
			}

			currentAskTotal += size
			askTotals[i] = currentAskTotal
		}

		// Print asks
		for i := nAsks - 1; i >= 0; i-- {
			price := book.Levels[1][i].Px
			size := book.Levels[1][i].Sz

			fmt.Printf(
				"\033[31m%-15s\033[0m %15s %16.2f\033[K\n",
				price,
				size,
				askTotals[i],
			)
		}

		// Calculate spread
		highestBid, _ := strconv.ParseFloat(book.Levels[0][0].Px, 64)
		lowestAsk, _ := strconv.ParseFloat(book.Levels[1][0].Px, 64)

		spread := lowestAsk - highestBid
		spreadPct := (spread / highestBid) * 100

		fmt.Println("--------------------------------------------------")
		fmt.Printf(
			"\033[1;33m%-15s %15.3f %15.3f%%\033[0m\033[K\n",
			"Spread",
			spread,
			spreadPct,
		)
		fmt.Println("--------------------------------------------------")

		// Process bids
		nBids := len(book.Levels[0])
		if nBids > maxRows {
			nBids = maxRows
		}

		currentBidTotal := 0.0

		for i := 0; i < nBids; i++ {
			price := book.Levels[0][i].Px
			size := book.Levels[0][i].Sz

			bidSize, err := strconv.ParseFloat(size, 64)
			if err != nil {
				continue
			}

			currentBidTotal += bidSize

			fmt.Printf(
				"\033[32m%-15s\033[0m %15s %16.2f\033[K\n",
				price,
				size,
				currentBidTotal,
			)
		}

		// Calculate book imbalance
		bookImbalance := 0.0

		if currentBidTotal+currentAskTotal > 0 {
			bookImbalance = ((currentBidTotal - currentAskTotal) /
				(currentBidTotal + currentAskTotal)) * 100
		}

		var bookSignal string
		var bookColor string

		if bookImbalance > 5 {
			bookSignal = "BULLISH"
			bookColor = Green
		} else if bookImbalance < -5 {
			bookSignal = "BEARISH"
			bookColor = Red
		} else {
			bookSignal = "NEUTRAL"
			bookColor = Yellow
		}

		// Calculate trade delta from the last 5 seconds
		buyVolume := 0.0
		sellVolume := 0.0

		mu.Lock()

		// Remove trades older than 5 seconds
		cutoff := time.Now().Add(-5 * time.Second).UnixMilli()

		for len(trades) > 0 && trades[0].Time < cutoff {
			trades = trades[1:]
		}

		// Calculate buy and sell volume
		for _, trade := range trades {
			size, err := strconv.ParseFloat(trade.Sz, 64)
			if err != nil {
				continue
			}

			if trade.Side == "B" {
				buyVolume += size
			} else {
				sellVolume += size
			}
		}

		mu.Unlock()

		tradeDelta := buyVolume - sellVolume

		var deltaSignal string
		var deltaColor string

		if tradeDelta > 0 {
			deltaSignal = "BULLISH"
			deltaColor = Green
		} else if tradeDelta < 0 {
			deltaSignal = "BEARISH"
			deltaColor = Red
		} else {
			deltaSignal = "NEUTRAL"
			deltaColor = Yellow
		}

		// Print market summary
		fmt.Println()
		fmt.Println("MARKET SUMMARY (5s)")
		fmt.Println("--------------------------------------------------")

		fmt.Printf(
			"%-20s %+10.4f BTC   %s%s%s\033[K\n",
			"Trade Delta",
			tradeDelta,
			deltaColor,
			deltaSignal,
			Reset,
		)

		fmt.Printf(
			"%-20s %+10.2f%%      %s%s%s\033[K\n",
			"Book Imbalance",
			bookImbalance,
			bookColor,
			bookSignal,
			Reset,
		)

		fmt.Println("--------------------------------------------------")

		fmt.Printf(
			"\033[90mLast update: %s | Press Ctrl+C to Exit\033[0m\033[K",
			time.UnixMilli(book.Time).Format("15:04:05.000"),
		)

		time.Sleep(pollingInterval)
	}
}