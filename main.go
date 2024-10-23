package main

import (
	"log"
	"os"
	"strings"
	"wallets-tracker-eth/walletsTracker"
)

func main() {
	content, err := os.ReadFile("wallets.txt")
	if err != nil {
		log.Fatal("Can't read wallets.txt")
	}
	wallets := strings.Split(string(content), "\n")

	walletsTracker.MonitorWallets(wallets)
}
