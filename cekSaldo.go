package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
    url := "https://api.hyperliquid-testnet.xyz/info"
    userAddress := "0x43eCaCC9684da91Cb8dABEBC3c819055E1B63Dd6"

    payload := map[string]interface{} {
        "type": "spotClearinghouseState",
        "user": userAddress,
    }

    jsonData, _ := json.Marshal(payload)

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Printf("Failed: %v\n", err)
        return
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        fmt.Printf("Failed parsing JSON: %v\n", err)
        return
    }

    if balances, ok := result["balances"].([]interface{}); ok {
        found := false
        for _, b := range balances {
            balanceMap := b.(map[string]interface{})
            if balanceMap["coin"] == "USDC" {
                fmt.Printf("Wallet Address: %s\n", userAddress)
                fmt.Printf("Balance: $%s USDC\n", balanceMap["total"])
                found = true
                break
            }
        }
        if !found {
            fmt.Println("Not Found.")
        }
    } else {
        fmt.Printf("field balances: %v\n", result)
    }
}