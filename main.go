package main

import (
	"boilerplate/database"
	"boilerplate/router"
	"context"
	"database/sql"
	"log"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/robfig/cron"
)

func main() {
	database.ConnectDb()
	router.Fetched()
	// Initialize Ethereum client
	client, err := rpc.Dial("https://rpc-alpha-testnet.saitascan.io")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	// Initialize MySQL database
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/saita")
	if err != nil {
		log.Fatalf("Failed to connect to the MySQL database: %v", err)
	}
	defer db.Close()

	// Create a cron job to fetch and store block details
	cronJob := cron.New()
	cronJob.AddFunc("@every 3s", func() {
		// Fetch the latest block number
		var latestBlockNum string
		if err := client.CallContext(context.Background(), &latestBlockNum, "eth_blockNumber"); err != nil {
			log.Printf("Failed to retrieve latest block number: %v", err)
			return
		}

		// Retrieve block details for the current block
		var block map[string]interface{}
		if err := client.CallContext(context.Background(), &block, "eth_getBlockByNumber", latestBlockNum, true); err != nil {
			log.Printf("Failed to retrieve block details for block number %s: %v", latestBlockNum, err)
			return
		}

		blockDetails := &router.BlockDetails{
			Number:       block["number"].(string),
			ParentHash:   block["parentHash"].(string),
			BlockHash:    block["hash"].(string),
			Timestamp:    block["timestamp"].(string),
			Transactions: len(block["transactions"].([]interface{})),
		}

		// Store block details in MySQL database
		stmt, err := db.Prepare(`INSERT INTO block_details (block_number, parent_hash, block_hash, timestamp, transaction_count) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			log.Printf("Failed to prepare SQL statement: %v", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(blockDetails.Number, blockDetails.ParentHash, blockDetails.BlockHash, blockDetails.Timestamp, blockDetails.Transactions)
		if err != nil {
			log.Printf("Failed to insert block details into MySQL: %v", err)
			return
		}

		log.Printf("Successfully fetched and stored block details for block number %s", latestBlockNum)
	})
	cronJob.Start()
	router.ClientRoutes()
	// defer cronJob.Stop()

	// Keep the main program running
	select {}
}
