package walletsTracker

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wallets-tracker-eth/discordWebhook"
)

const etherscanAPIKey = "YOUR_API_KEY"

type tokenTxs struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Result  []txnResponse `json:"result"`
}

type txnResponse struct {
	Hash            string `json:"hash"`
	From            string `json:"from"`
	ContractAddress string `json:"contractAddress"`
	To              string `json:"to"`
	Value           string `json:"value"`
	TokenSymbol     string `json:"tokenSymbol"`
	TokenDecimal    string `json:"tokenDecimal"`
}

type internalTxs struct {
	Status  string                `json:"status"`
	Message string                `json:"message"`
	Result  []internalTxnResponse `json:"result"`
}

type internalTxnResponse struct {
	Hash  string `json:"hash"`
	Value string `json:"value"`
}

type txByHashResponse struct {
	Result txByHashResponseResult `json:"result"`
}

type txByHashResponseResult struct {
	Value string `json:"value"`
}

// Helper to fetch a transaction by its hash
func getValueByHash(hash string) (string, bool) {
	requestURL := fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", hash, etherscanAPIKey)

	resp, err := http.Get(requestURL)
	if err != nil {
		return "0", false
	}
	defer resp.Body.Close()

	var response txByHashResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Println("Error parsing JSON:", err)
		return "0", false
	}

	// Convert hex to big.Int
	decimalValue, success := new(big.Int).SetString(response.Result.Value[2:], 16)
	if !success {
		log.Println("Error converting hex to decimal")
		return "0", false
	}

	if response.Result.Value == "0x0" {
		return "0", true
	}

	// Convert value to float using big.Rat for precision
	tenToTheEighteen := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	result := new(big.Rat).SetFrac(decimalValue, tenToTheEighteen)
	decimalFloat, _ := result.Float64()

	return fmt.Sprintf("%.3f", decimalFloat), true
}

// Helper to fetch internal transaction value by hash
func getInternalTxValueByHash(wallet, hash string) (string, bool) {
	requestURL := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlistinternal&address=%s&page=1&offset=10&sort=desc&apikey=%s", wallet, etherscanAPIKey)

	resp, err := http.Get(requestURL)
	if err != nil {
		return "0", false
	}
	defer resp.Body.Close()

	var response internalTxs
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Println("Error parsing JSON:", err)
		return "0", false
	}

	if response.Message != "OK" {
		return "0", false
	}

	for _, tx := range response.Result {
		if tx.Hash == hash {
			decimalValue, err := strconv.ParseFloat(tx.Value, 64)
			if err != nil {
				log.Println("Error parsing float:", err)
				return "0", false
			}
			return fmt.Sprintf("%.3f", decimalValue/math.Pow(10, 18)), true
		}
	}

	return "0", true
}

// Helper to get ERC20 transactions by wallet
func getERC20Txs(wallet string) ([]txnResponse, bool) {
	requestURL := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=tokentx&address=%s&page=1&offset=10&sort=desc&apikey=%s", wallet, etherscanAPIKey)
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()

	var response tokenTxs
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Println("Error parsing JSON:", err)
		return nil, false
	}

	if response.Message != "OK" {
		return nil, false
	}

	return response.Result, true
}

// Check if a string is in a slice
func inSlice(str string, slice []string) bool {
	for _, elem := range slice {
		if str == elem {
			return true
		}
	}
	return false
}

// Main function to monitor wallet transactions
func MonitorWallets(wallets []string) {
	walletsTxns := make(map[string][]string)

	for {
		for _, wallet := range wallets {
			txs, ok := getERC20Txs(wallet)
			if !ok {
				log.Printf("Error getting txs for wallet %s\n", wallet)
				time.Sleep(1 * time.Second)
				continue
			}

			if _, exists := walletsTxns[wallet]; !exists {
				for _, tx := range txs {
					walletsTxns[wallet] = append(walletsTxns[wallet], tx.Hash)
				}
				time.Sleep(1 * time.Second)
				continue
			}

			if inSlice(txs[0].Hash, walletsTxns[wallet]) {
				time.Sleep(1 * time.Second)
				continue
			}

			for _, tx := range txs {
				if !inSlice(tx.Hash, walletsTxns[wallet]) {
					decimal, err := strconv.ParseFloat(tx.TokenDecimal, 64)
					if err != nil {
						log.Println("Error parsing TokenDecimal:", err)
						continue
					}
					amount, err := strconv.ParseFloat(tx.Value, 64)
					if err != nil {
						log.Println("Error parsing Value:", err)
						continue
					}

					var swapMessage string
					if strings.ToLower(tx.From) == strings.ToLower(wallet) {
						value, ok := getInternalTxValueByHash(wallet, tx.Hash)
						if !ok || value == "0" {
							continue
						}
						swapMessage = fmt.Sprintf("[%s](https://etherscan.io/address/%s) sold %.3f $%s for %s $ETH", wallet, wallet, amount/math.Pow(10, decimal), tx.TokenSymbol, value)
					} else {
						value, ok := getValueByHash(tx.Hash)
						if !ok || value == "0" {
							continue
						}
						swapMessage = fmt.Sprintf("[%s](https://etherscan.io/address/%s) bought %.3f $%s for %s $ETH", wallet, wallet, amount/math.Pow(10, decimal), tx.TokenSymbol, value)
					}

					tokenLink := fmt.Sprintf("[%s](https://etherscan.io/token/%s)", tx.ContractAddress, tx.ContractAddress)
					if err := discordWebhook.SendWebhook(swapMessage, tx.Hash, tokenLink); err != nil {
						log.Printf("Error sending webhook: %v\n", err)
					}
				} else {
					break
				}
			}

			// Reset transactions for the wallet
			walletsTxns[wallet] = nil
			for _, tx := range txs {
				walletsTxns[wallet] = append(walletsTxns[wallet], tx.Hash)
			}

			time.Sleep(1 * time.Second)
		}
	}
}
