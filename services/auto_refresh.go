package services

import (
	"log"
	"time"
	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

// AutoRefreshBalances periodically updates balances for tracked addresses
func AutoRefreshBalances(interval time.Duration) {
	go func() {
		for {
			addresses := db.GetTrackedAddresses() // Implement this to return addresses to track
			for _, addr := range addresses {
				balance, err := utils.GetEthBalance(addr)
				if err != nil {
					log.Printf("Error fetching balance for %s: %v", addr, err)
					continue
				}
				db.UpdateAddressBalance(addr, balance) // Implement this to update DB
			}
			time.Sleep(interval)
		}
	}()
}
