package main

import (
	"boilerplate/controller"
	"boilerplate/database"
	"boilerplate/router"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/robfig/cron"
	"github.com/streadway/amqp"
)

var (
	latestFetchedBlockNumber string
	mutex                    sync.Mutex
)

func main() {
	database.ConnectDb()

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

	// RabbitMQ connection details
	rabbitMQURL := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create a RabbitMQ channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare a RabbitMQ queue
	queueName := "block_queue"
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Create a cron job to fetch and store block details
	cronJob := cron.New()
	cronJob.AddFunc("@every 1s", func() {
		// Fetch the latest block number
		var latestBlockNum string
		if err := client.CallContext(context.Background(), &latestBlockNum, "eth_blockNumber"); err != nil {
			log.Printf("Failed to retrieve latest block number: %v", err)
			return
		}

		mutex.Lock()
		defer mutex.Unlock()

		// Check if the block is already stored in the database
		if isBlockStored(db, latestBlockNum) || latestBlockNum == latestFetchedBlockNumber {
			log.Printf("Block number %s already stored or fetched. Skipping.", latestBlockNum)
			return
		}

		// Set the latest fetched block number
		latestFetchedBlockNumber = latestBlockNum

		// Retrieve block details for the current block
		var block map[string]interface{}
		if err := client.CallContext(context.Background(), &block, "eth_getBlockByNumber", latestBlockNum, true); err != nil {
			log.Printf("Failed to retrieve block details for block number %s: %v", latestBlockNum, err)
			return
		}

		blockDetails := &controller.BlockDetails{
			Number:       block["number"].(string),
			ParentHash:   block["parentHash"].(string),
			BlockHash:    block["hash"].(string),
			Timestamp:    block["timestamp"].(string),
			Transactions: len(block["transactions"].([]interface{})),
		}

		// Store block details in MySQL database
		err := storeBlockDetails(db, blockDetails)
		if err != nil {
			log.Printf("Failed to insert block details into MySQL: %v", err)
			return
		}

		// Publish block details to RabbitMQ
		err = ch.Publish("", queueName, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        serializeBlockDetails(blockDetails),
		})
		if err != nil {
			log.Printf("Failed to publish block details to RabbitMQ: %v", err)
			return
		}

		log.Printf("Successfully fetched and stored block details for block number %s", latestBlockNum)
	})
	cronJob.Start()
	router.ClientRoutes()

	// Keep the main program running
	select {}
}

func isBlockStored(db *sql.DB, blockNumber string) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM block_details WHERE block_number = ?", blockNumber).Scan(&count)
	if err != nil {
		log.Printf("Error checking if block is stored: %v", err)
		return true // Assume block is stored to avoid duplicates on error
	}
	return count > 0
}

func storeBlockDetails(db *sql.DB, blockDetails *controller.BlockDetails) error {
	stmt, err := db.Prepare(`INSERT INTO block_details (block_number, parent_hash, block_hash, timestamp, transaction_count) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(blockDetails.Number, blockDetails.ParentHash, blockDetails.BlockHash, blockDetails.Timestamp, blockDetails.Transactions)
	return err
}

func serializeBlockDetails(blockDetails *controller.BlockDetails) []byte {
	data, err := json.Marshal(blockDetails)
	if err != nil {
		log.Printf("Failed to serialize block details: %v", err)
		return nil
	}
	return data
}
