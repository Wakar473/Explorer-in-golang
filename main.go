//storing transaction details in db through websocket
 package main

import (
	"boilerplate/database"
	"boilerplate/router"
	"boilerplate/utils"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

var (
	wsConn          *websocket.Conn
	rabbitURL       = "amqp://guest:guest@localhost:5672/"
	queueName       = "block_queue"
	db              *sql.DB
	rabbitConn      *amqp.Connection
	mutex           sync.Mutex
	dbMutex         sync.Mutex
	processedBlocks = make(map[string]bool)
)

type BlockDetails struct {
	Number             string               `json:"number"`
	ParentHash         string               `json:"parentHash"`
	BlockHash          string               `json:"hash"`
	Timestamp          string               `json:"timestamp"`
	Transactions       string               `json:"transactions"`
	TransactionDetails []TransactionDetails `json:"transactions"`
}

type TransactionDetails struct {
	BlockHash            string `json:"blockHash"`
	BlockNumber          string `json:"numberHash"`
	ChainId              string `json:"Id"`
	From_address                string `json:"from"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	Hash                 string `json:"hash"`
	MaxFeePerGas         string `json:"maxFeeHash"`
	MaxPriorityFeePerGas string `json:"priorityHash"`
	Nonce                string `json:"nonce"`
	To_address                   string `json:"to"`
	TransactionIndex     string `json:"transaction"`
	Value                string `json:"value"`
}

func main() {
	database.ConnectDb()
	utils.GetClientPort()
	go router.ClientRoutes()
	// Initialize RabbitMQ connection
	var err error
	rabbitConn, err = amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	// Initialize MySQL connection
	db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/saita")
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// Establish a WebSocket connection
	wsURL := "wss://wss-testnet-nodes.nexablockscan.io"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()
	wsConn = conn

	// Start the worker to consume messages from RabbitMQ and store in MySQL
	go startWorker()

	// Start the worker to consume transaction details from RabbitMQ and store in MySQL
	go startTransactionWorker()

	// Fetch block details periodically
	for {
		if err := fetchBlockDetails("latest"); err != nil {
			log.Printf("Error fetching block details: %v", err)
		}

		// Wait for some time before fetching the next block (adjust as needed)
		time.Sleep(1 * time.Second)
	}
}
func startTransactionWorker() {
	// Initialize RabbitMQ channel
	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}
	defer ch.Close()

	// Declare RabbitMQ queue for transactions
	_, err = ch.QueueDeclare("transaction_details_queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Consume messages from "transaction_details_queue"
	msgs, err := ch.Consume("transaction_details_queue", "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume messages from RabbitMQ: %v", err)
	}

	// Process incoming transaction messages
	for msg := range msgs {
		var transactionDetails TransactionDetails
		err := json.Unmarshal(msg.Body, &transactionDetails)
		if err != nil {
			log.Printf("Error unmarshalling RabbitMQ message: %v", err)
			// Nack the message to avoid it being removed from the queue
			msg.Nack(false, false)
			continue
		}

		err = storeTransactionInMySQL(&transactionDetails, db)
		if err != nil {
			log.Printf("Failed to store transaction in MySQL: %v", err)
			// Nack the message to avoid it being removed from the queue
			msg.Nack(false, false)
			continue
		}

		// Acknowledge the message
		msg.Ack(false)
	}
}

func startWorker() {
	// Initialize RabbitMQ channel
	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}
	defer ch.Close()

	// Declare RabbitMQ queue
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Consume messages from RabbitMQ
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume messages from RabbitMQ: %v", err)
	}

	// Process incoming messages
	for msg := range msgs {
		var blockDetails BlockDetails
		err := json.Unmarshal(msg.Body, &blockDetails)
		if err != nil {
			log.Printf("Error unmarshalling RabbitMQ message: %v", err)
			// Nack the message to avoid it being removed from the queue
			msg.Nack(false, false)
			continue
		}

		// Check if the block has already been processed
		if processedBlocks[blockDetails.Number] {
			log.Printf("Block %s already processed. Skipping.", blockDetails.Number)
			// Acknowledge the message
			msg.Ack(false)
			continue
		}

		// Store in MySQL
		err = storeInMySQL(&blockDetails, db)
		if err != nil {
			log.Printf("Failed to store in MySQL: %v", err)
			// Nack the message to avoid it being removed from the queue
			msg.Nack(false, false)
			continue
		}

		// Mark the block as processed
		processedBlocks[blockDetails.Number] = true

		// Acknowledge the message
		msg.Ack(false)
	}
}

func fetchBlockDetails(blockNumber string) error {
	mutex.Lock()
	defer mutex.Unlock()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{blockNumber, true},
		"id":      1,
	}

	err := wsConn.WriteJSON(request)
	if err != nil {
		return err
	}

	var response map[string]interface{}
	err = wsConn.ReadJSON(&response)
	if err != nil {
		return err
	}

	if response["error"] != nil {
		// Handle error if needed
		return nil
	}

	blockDetails := response["result"].(map[string]interface{})
	log.Printf("Fetched block details for block number %s: %+v", blockNumber, blockDetails)

	// Publish block details to RabbitMQ
	err = publishToRabbitMQ("block_details_queue", blockDetails)
	if err != nil {
		log.Printf("Failed to publish block details to RabbitMQ: %v", err)
		return err
	}

	// Publish transaction details to RabbitMQ
	transactions := blockDetails["transactions"].([]interface{})
	for _, transaction := range transactions {
		transactionDetails := transaction.(map[string]interface{})
		transactionDetails["blockNumber"] = blockNumber
		err = publishToRabbitMQ("transaction_details_queue", transactionDetails)
		if err != nil {
			log.Printf("Failed to publish transaction details to RabbitMQ: %v", err)
			// Handle the error as needed
		}
	}

	return nil
}

func publishToRabbitMQ(queueName string, details map[string]interface{}) error {
	// Initialize RabbitMQ channel
	ch, err := rabbitConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare RabbitMQ queue
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Convert details map to JSON
	message, err := json.Marshal(details)
	if err != nil {
		return err
	}
	return ch.Publish("", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        message,
	})
}

func storeInMySQL(details *BlockDetails, db *sql.DB) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Your MySQL insertion logic
	stmt, err := db.Prepare(`INSERT INTO block_details (block_number, parent_hash, block_hash, timestamp, transaction_count) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(details.Number, details.ParentHash, details.BlockHash, details.Timestamp, details.Transactions)
	if err != nil {
		return err
	}

	log.Printf("Successfully inserted block details into MySQL for block number %s", details.Number)

	return nil
}

func storeTransactionInMySQL(transaction *TransactionDetails, db *sql.DB) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Your MySQL insertion logic for transaction details
	stmtTransaction, err := db.Prepare(`
	INSERT INTO transaction_details (block_hash, block_number, chain_id, from_address, gas, gas_price, hash, maxFee_PerGas, maxPriorityfee_PerGas, nonce, to_address, transaction_index, value)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmtTransaction.Close()

	_, err = stmtTransaction.Exec(
		transaction.BlockHash,
		transaction.BlockNumber,
		transaction.ChainId,
		transaction.From_address,
		transaction.Gas,
		transaction.GasPrice,
		transaction.Hash,
		transaction.MaxFeePerGas,
		transaction.MaxPriorityFeePerGas,
		transaction.Nonce,
		transaction.To_address,
		transaction.TransactionIndex,
		transaction.Value,
	)
	if err != nil {
		return err
	}

	log.Printf("Successfully inserted transaction details into MySQL for transaction hash %s", transaction.Hash)
	return nil
}