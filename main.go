// package main

// import (
// 	"boilerplate/controller"
// 	"boilerplate/database"
// 	"boilerplate/router"
// 	"context"
// 	"database/sql"
// 	"encoding/json"
// 	"log"
// 	"strings"
// 	"sync"

// 	"github.com/ethereum/go-ethereum/rpc"
// 	"github.com/robfig/cron"
// 	"github.com/streadway/amqp"
// )

// var (
// 	latestFetchedBlockNumber string
// 	mutex                    sync.Mutex
// )

// func main() {
// 	database.ConnectDb()

// 	// client, err := rpc.Dial("https://rpc-alpha-testnet.saitascan.io")
// 	// if err != nil {
// 	// 	log.Fatalf("Failed to connect to the Ethereum client: %v", err)
// 	// }
// 	// defer client.Close()
// 	client, err := rpc.Dial("wss://wss-alpha-testnet.saitascan.io")
// 	if err != nil {
// 		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
// 	}
// 	defer client.Close()

// 	// Initialize MySQL database
// 	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/saita")
// 	if err != nil {
// 		log.Fatalf("Failed to connect to the MySQL database: %v", err)
// 	}
// 	defer db.Close()

// 	// RabbitMQ connection details
// 	rabbitMQURL := "amqp://guest:guest@localhost:5672/"
// 	conn, err := amqp.Dial(rabbitMQURL)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
// 	}
// 	defer conn.Close()

// 	// Create a RabbitMQ channel
// 	ch, err := conn.Channel()
// 	if err != nil {
// 		log.Fatalf("Failed to open a channel: %v", err)
// 	}
// 	defer ch.Close()

// 	// Declare a RabbitMQ queue
// 	queueName := "block_queue"
// 	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
// 	if err != nil {
// 		log.Fatalf("Failed to declare a queue: %v", err)
// 	}

// 	// Declare a RabbitMQ queue for transaction details
// 	transactionQueueName := "transaction_queue"
// 	_, err = ch.QueueDeclare(transactionQueueName, true, false, false, false, nil)
// 	if err != nil {
// 		log.Fatalf("Failed to declare a transaction queue: %v", err)
// 	}

// 	// Create a cron job to fetch the latest block number
// 	cronJob := cron.New()
// 	cronJob.AddFunc("@every 1s", func() {
// 		// Fetch the latest block number
// 		var latestBlockNum string
// 		if err := client.CallContext(context.Background(), &latestBlockNum, "eth_blockNumber"); err != nil {
// 			log.Printf("Failed to retrieve latest block number: %v", err)
// 			return
// 		}

// 		mutex.Lock()
// 		defer mutex.Unlock()

// 		// Check if the block is already stored in the database
// 		if isBlockStored(db, latestBlockNum) || latestBlockNum == latestFetchedBlockNumber {
// 			log.Printf("Block number %s already stored or fetched. Skipping.", latestBlockNum)
// 			return
// 		}

// 		// Set the latest fetched block number
// 		latestFetchedBlockNumber = latestBlockNum

// 		// Retrieve block details for the current block
// 		var block map[string]interface{}
// 		if err := client.CallContext(context.Background(), &block, "eth_getBlockByNumber", latestBlockNum, true); err != nil {
// 			log.Printf("Failed to retrieve block details for block number %s: %v", latestBlockNum, err)
// 			return
// 		}

// 		// log.Println("latest block", block["number"].(string))
// 		// log.Println("transactions", block["transactions"].([]interface{}))
// 		// transactions:=block["transactions"].([]interface{});

// 		blockDetails := &controller.BlockDetails{
// 			Number:       block["number"].(string),
// 			ParentHash:   block["parentHash"].(string),
// 			BlockHash:    block["hash"].(string),
// 			Timestamp:    block["timestamp"].(string),
// 			Transactions: len(block["transactions"].([]interface{})),
// 		}
// 		// Store block details in MySQL database
// 		err := storeBlockDetails(db, blockDetails)
// 		if err != nil {
// 			log.Printf("Failed to insert block details into MySQL: %v", err)
// 			return
// 		}
// 		var tx []string
// 		for _, txHashInterface := range block["transactions"].([]interface{}) {
// 			tx = append(tx, txHashInterface.(string))
// 		}

// 		for _, txHash := range tx {
// 			var tx map[string]interface{}
// 			if err := client.CallContext(context.Background(), &tx, "eth_getTransactionByNumber", txHash); err != nil {
// 				log.Printf("Failed to retrieve transaction details for blockNumber %s: %v", txHash, err)
// 				continue
// 			}
// 			txDetails := &controller.TransactionDetails{
// 				BlockHash:            tx["blockHash"].(string),
// 				BlockNumber:          tx["numberHash"].(string),
// 				ChainId:              tx["Id"].(string),
// 				From:                 tx["from"].(string),
// 				Gas:                  tx["gas"].(string),
// 				GasPrice:             tx["gasPrice"].(string),
// 				Hash:                 tx["hash"].(string),
// 				MaxFeePerGas:         tx["maxFeeHash"].(string),
// 				MaxPriorityFeePerGas: tx["priorityHash"].(string),
// 				Nonce:                tx["nonce"].(string),
// 				To:                   tx["to"].(string),
// 				TransactionIndex:     tx["transaction"].(string),
// 				Value:                tx["value"].(string),
// 			}
// 			// log.Println("hello workd============", tx)

// 			// Publish transaction details to RabbitMQ
// 			err := ch.Publish("", transactionQueueName, false, false, amqp.Publishing{
// 				ContentType: "application/json",
// 				Body:        serializeTransactionDetails(txDetails),
// 			})
// 			if err != nil {
// 				log.Printf("Failed to publish transaction details to RabbitMQ: %v", err)
// 				return
// 			}

// 			log.Printf("Successfully fetched and published transaction details for block number %s", latestBlockNum)
// 		}

// 		// Publish block details to RabbitMQ
// 		err = ch.Publish("", queueName, false, false, amqp.Publishing{
// 			ContentType: "application/json",
// 			Body:        serializeBlockDetails(blockDetails),
// 		})
// 		if err != nil {
// 			log.Printf("Failed to publish block details to RabbitMQ: %v", err)
// 			return
// 		}

// 		log.Printf("Successfully fetched and published block details for block number %s", latestBlockNum)
// 	})
// 	cronJob.Start()
// 	// Create a RabbitMQ consumer to store block details in the MySQL database
// 	go consumeBlockDetails(ch, db)
// 	router.ClientRoutes()

// 	// Keep the main program running
// 	select {}
// }

// func consumeBlockDetails(ch *amqp.Channel, db *sql.DB) {
// 	msgs, err := ch.Consume("block_queue", "", true, false, false, false, nil)
// 	if err != nil {
// 		log.Fatalf("Failed to register a consumer: %v", err)
// 	}

// 	for msg := range msgs {
// 		var details interface{}

// 		// Determine if the message contains block or transaction details
// 		if strings.Contains(string(msg.Body), "block") {
// 			var blockDetails controller.BlockDetails
// 			if err := json.Unmarshal(msg.Body, &blockDetails); err != nil {
// 				log.Printf("Failed to deserialize block details: %v", err)
// 				continue
// 			}

// 			// Store block details in MySQL database
// 			err := storeBlockDetails(db, &blockDetails)
// 			if err != nil {
// 				log.Printf("Failed to insert block details into MySQL: %v", err)
// 				continue
// 			}

// 			details = &blockDetails
// 		} else if strings.Contains(string(msg.Body), "transaction") {
// 			var txDetails controller.TransactionDetails
// 			if err := json.Unmarshal(msg.Body, &txDetails); err != nil {
// 				log.Printf("Failed to deserialize transaction details: %v", err)
// 				continue
// 			}

// 			// Store transaction details in MySQL database asynchronously
// 			go storeTransactionDetailsAsync(db, &txDetails)

// 			details = &txDetails
// 		}

// 		log.Printf("Successfully stored details: %v", details)
// 	}
// }

// // Function to store transaction details asynchronously
// func storeTransactionDetailsAsync(db *sql.DB, txDetails *controller.TransactionDetails) {
// 	if err := storeTransactionDetails(db, txDetails); err != nil {
// 		log.Printf("Failed to store transaction details asynchronously: %v", err)
// 	}
// }

// func isBlockStored(db *sql.DB, blockNumber string) bool {
// 	var count int
// 	err := db.QueryRow("SELECT COUNT(*) FROM block_details WHERE block_number = ?", blockNumber).Scan(&count)
// 	if err != nil {
// 		log.Printf("Error checking if block is stored: %v", err)
// 		return true // Assume block is stored to avoid duplicates on error
// 	}
// 	return count > 0
// }

// func storeBlockDetails(db *sql.DB, blockDetails *controller.BlockDetails) error {
// 	stmt, err := db.Prepare(`INSERT INTO block_details (block_number, parent_hash, block_hash, timestamp, transaction_count) VALUES (?, ?, ?, ?, ?)`)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	_, err = stmt.Exec(blockDetails.Number, blockDetails.ParentHash, blockDetails.BlockHash, blockDetails.Timestamp, blockDetails.Transactions)
// 	return err
// }
// func storeTransactionDetails(db *sql.DB, txDetails *controller.TransactionDetails) error {
// 	stmt, err := db.Prepare(`INSERT INTO transaction_details (BlockNumber,ChainId,From,gas,gasPrice,hash,maxFeePerGas,maxPriorityFeePerGas,nonce,to,transactionIndex,value) VALUES (?, ?, ?, ?, ?,?, ?, ?, ?, ?, ?, ?, ?)`)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()
// 	stmt.Exec(txDetails.BlockHash, txDetails.BlockNumber, txDetails.ChainId, txDetails.From, txDetails.Gas, txDetails.GasPrice, txDetails.Hash, txDetails.MaxFeePerGas, txDetails.MaxPriorityFeePerGas, txDetails.Nonce, txDetails.To, txDetails.TransactionIndex, txDetails.Value)
// 	return err

// }

// func serializeBlockDetails(blockDetails *controller.BlockDetails) []byte {
// 	data, err := json.Marshal(blockDetails)
// 	if err != nil {
// 		log.Printf("Failed to serialize block details: %v", err)
// 		return nil
// 	}
// 	return data

// }

// func serializeTransactionDetails(txDetails *controller.TransactionDetails) []byte {
// 	data, err := json.Marshal(txDetails)
// 	if err != nil {
// 		log.Printf("Failed to serialize transaction details: %v", err)
// 		return nil
// 	}
// 	return data
// }

// package main

// import (
// 	"boilerplate/database"
// 	"boilerplate/router"
// 	"boilerplate/utils"
// 	"database/sql"
// 	"encoding/json"
// 	"log"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/websocket"
// 	"github.com/streadway/amqp"
// )

// var (
// 	wsConn          *websocket.Conn
// 	rabbitURL       = "amqp://guest:guest@localhost:5672/"
// 	queueName       = "block_queue"
// 	db              *sql.DB
// 	rabbitConn      *amqp.Connection
// 	mutex           sync.Mutex
// 	dbMutex         sync.Mutex
// 	processedBlocks = make(map[string]bool)
// )

// type BlockDetails struct {
// 	Number       string `json:"number"`
// 	ParentHash   string `json:"parentHash"`
// 	BlockHash    string `json:"hash"`
// 	Timestamp    string `json:"timestamp"`
// 	Transactions interface{}    `json:"transactions"`
// }

// type TransactionDetails struct {
// 	BlockHash              string `json:"blockHash"`
// 	BlockNumber            string `json:"numberHash"`
// 	ChainId                string `json:"Id"`
// 	From                   string `json:"from"`
// 	Gas                    string `json:"gas"`
// 	GasPrice               string `json:"gasPrice"`
// 	Hash                   string `json:"hash"`
// 	MaxFeePerGas           string `json:"maxFeeHash"`
// 	MaxPriorityFeePerGas   string `json:"priorityHash"`
// 	Nonce                  string `json:"nonce"`
// 	To                     string `json:"to"`
// 	TransactionIndex       string `json:"transaction"`
// 	Value                  string `json:"value"`
// }

// func main() {
// 	database.ConnectDb()
// 	utils.GetClientPort()
// 	go router.ClientRoutes()
// 	// Initialize RabbitMQ connection
// 	var err error
// 	rabbitConn, err = amqp.Dial(rabbitURL)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
// 	}
// 	defer rabbitConn.Close()

// 	// Initialize MySQL connection
// 	db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/saita")
// 	if err != nil {
// 		log.Fatalf("Failed to connect to MySQL: %v", err)
// 	}
// 	defer db.Close()

// 	// Establish a WebSocket connection
// 	wsURL := "wss://wss-alpha-testnet.saitascan.io"
// 	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to WebSocket: %v", err)
// 	}
// 	defer conn.Close()
// 	wsConn = conn

// 	// Start the worker to consume messages from RabbitMQ and store in MySQL
// 	go startWorker()

// 	// Fetch block details periodically
// 	for {
// 		if err := fetchBlockDetails("latest"); err != nil {
// 			log.Printf("Error fetching block details: %v", err)
// 		}

// 		// Wait for some time before fetching the next block (adjust as needed)
// 		time.Sleep(1 * time.Second)
// 	}
// }

// func startWorker() {
// 	// Initialize RabbitMQ channel
// 	ch, err := rabbitConn.Channel()
// 	if err != nil {
// 		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
// 	}
// 	defer ch.Close()

// 	// Declare RabbitMQ queue
// 	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
// 	if err != nil {
// 		log.Fatalf("Failed to declare a queue: %v", err)
// 	}

// 	// Consume messages from RabbitMQ
// 	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
// 	if err != nil {
// 		log.Fatalf("Failed to consume messages from RabbitMQ: %v", err)
// 	}

// 	for msg := range msgs {
// 		var blockDetails BlockDetails
// 		err := json.Unmarshal(msg.Body, &blockDetails)
// 		if err != nil {
// 			log.Printf("Error unmarshalling RabbitMQ message: %v", err)
// 			// Nack the message to avoid it being removed from the queue
// 			msg.Nack(false, false)
// 			continue
// 		}

// 		// Check if the block has already been processed
// 		if processedBlocks[blockDetails.Number] {
// 			log.Printf("Block %s already processed. Skipping.", blockDetails.Number)
// 			// Acknowledge the message
// 			msg.Ack(false)
// 			continue
// 		}

// 		// Convert Transactions to int if it's an array
// 		if transactionsArray, ok := blockDetails.Transactions.([]interface{}); ok {
// 			blockDetails.Transactions = len(transactionsArray)
// 		}

// 		// Store block details in MySQL
// 		err = storeBlockDetailsInMySQL(&blockDetails, db)
// 		if err != nil {
// 			log.Printf("Failed to store block details in MySQL: %v", err)
// 			// Nack the message to avoid it being removed from the queue
// 			msg.Nack(false, false)
// 			continue
// 		}

// 		// Mark the block as processed
// 		processedBlocks[blockDetails.Number] = true

// 		// Acknowledge the message
// 		msg.Ack(false)
// 	}
// }

// func fetchBlockDetails(blockNumber string) error {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	request := map[string]interface{}{
// 		"jsonrpc": "2.0",
// 		"method":  "eth_getBlockByNumber",
// 		"params":  []interface{}{blockNumber, true},
// 		"id":      1,
// 	}

// 	err := wsConn.WriteJSON(request)
// 	if err != nil {
// 		return err
// 	}

// 	var response map[string]interface{}
// 	err = wsConn.ReadJSON(&response)
// 	if err != nil {
// 		return err
// 	}

// 	if response["error"] != nil {
// 		// Handle error if needed
// 		return nil
// 	}

// 	blockDetails := response["result"].(map[string]interface{})
// 	log.Printf("Fetched block details for block number %s: %+v", blockNumber, blockDetails)

// 	// Publish to RabbitMQ
// 	err = publishBlockToRabbitMQ(blockDetails)
// 	if err != nil {
// 		log.Printf("Failed to publish block to RabbitMQ: %v", err)
// 		return err
// 	}

// 	// Fetch transaction details
// 	transactions, ok := blockDetails["transactions"].([]interface{})
// 	if ok {
// 		for _, tx := range transactions {
// 			txHash, ok := tx.(string)
// 			if !ok {
// 				log.Printf("Invalid transaction hash format: %+v", tx)
// 				continue
// 			}

// 			// Fetch and process individual transaction details
// 			err := fetchAndStoreTransactionDetails(txHash)
// 			if err != nil {
// 				log.Printf("Error fetching and storing transaction details: %v", err)
// 			}
// 		}
// 	}

// 	return nil
// }

// func publishBlockToRabbitMQ(details map[string]interface{}) error {
// 	// Initialize RabbitMQ channel
// 	ch, err := rabbitConn.Channel()
// 	if err != nil {
// 		return err
// 	}
// 	defer ch.Close()

// 	// Declare RabbitMQ queue
// 	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
// 	if err != nil {
// 		return err
// 	}

// 	// Convert details map to JSON
// 	message, err := json.Marshal(details)
// 	if err != nil {
// 		return err
// 	}

// 	// Publish the JSON message to RabbitMQ
// 	return ch.Publish("", queueName, false, false, amqp.Publishing{
// 		ContentType: "application/json",
// 		Body:        message,
// 	})
// }

// func fetchAndStoreTransactionDetails(txHash string) error {
// 	// Implement the logic to fetch transaction details using the txHash
// 	// For example, you can make an Ethereum RPC call to get the details

// 	// Dummy transaction details (replace this with actual logic)
// 	txDetails := TransactionDetails{
// 		BlockHash:            "dummyBlockHash",
// 		BlockNumber:          "1797232",
// 		ChainId:              "dummyChainId",
// 		From:                 "dummyFromAddress",
// 		Gas:                  "100000",
// 		GasPrice:             "1000000000",
// 		Hash:                 txHash,
// 		MaxFeePerGas:         "5000000000",
// 		MaxPriorityFeePerGas: "1000000000",
// 		Nonce:                "1",
// 		To:                   "dummyToAddress",
// 		TransactionIndex:     "0",
// 		Value:                "123456789",
// 	}

// 	// Store transaction details in MySQL
// 	err := storeTransactionDetailsInMySQL(&txDetails, db)
// 	if err != nil {
// 		log.Printf("Failed to store transaction details in MySQL: %v", err)
// 		return err
// 	}

// 	return nil
// }

// func storeTransactionDetailsInMySQL(details *TransactionDetails, db *sql.DB) error {
// 	dbMutex.Lock()
// 	defer dbMutex.Unlock()

// 	// Your MySQL insertion logic for transaction details
// 	stmt, err := db.Prepare(`
// 		INSERT INTO transaction_details (
// 			block_hash, block_number, chain_id, from_address, gas, gas_price,
// 			hash, max_fee_per_gas, max_priority_fee_per_gas, nonce, to_address,
// 			transaction_index, value
// 		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
// 	`)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	_, err = stmt.Exec(
// 		details.BlockHash, details.BlockNumber, details.ChainId, details.From, details.Gas,
// 		details.GasPrice, details.Hash, details.MaxFeePerGas, details.MaxPriorityFeePerGas,
// 		details.Nonce, details.To, details.TransactionIndex, details.Value,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	log.Printf("Successfully inserted transaction details into MySQL for transaction hash %s", details.Hash)
// 	return nil
// }

// func storeBlockDetailsInMySQL(details *BlockDetails, db *sql.DB) error {
// 	dbMutex.Lock()
// 	defer dbMutex.Unlock()

// 	// Your MySQL insertion logic for block details
// 	stmt, err := db.Prepare(`
// 		INSERT INTO block_details (
// 			block_number, parent_hash, block_hash, timestamp, transaction_count
// 		) VALUES (?, ?, ?, ?, ?)
// 	`)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()

// 	_, err = stmt.Exec(
// 		details.Number, details.ParentHash, details.BlockHash, details.Timestamp, details.Transactions,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	log.Printf("Successfully inserted block details into MySQL for block number %s", details.Number)
// 	return nil
// }

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
	Number        string `json:"number"`
	ParentHash    string `json:"parentHash"`
	BlockHash     string `json:"hash"`
	Timestamp     string `json:"timestamp"`
	Transactions  string    `json:"transactions"`
	TransactionDetails []TransactionDetails `json:"transactions"`
}

type TransactionDetails struct {
	BlockHash              string `json:"blockHash"`
	BlockNumber            string `json:"numberHash"`
	ChainId                string `json:"Id"`
	From                   string `json:"from"`
	Gas                    string `json:"gas"`
	GasPrice               string `json:"gasPrice"`
	Hash                   string `json:"hash"`
	MaxFeePerGas           string `json:"maxFeeHash"`
	MaxPriorityFeePerGas   string `json:"priorityHash"`
	Nonce                  string `json:"nonce"`
	To                     string `json:"to"`
	TransactionIndex       string `json:"transaction"`
	Value                  string `json:"value"`
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
	wsURL := "wss://wss-alpha-testnet.saitascan.io"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()
	wsConn = conn

	// Start the worker to consume messages from RabbitMQ and store in MySQL
	go startWorker()

	// Fetch block details periodically
	for {
		if err := fetchBlockDetails("latest"); err != nil {
			log.Printf("Error fetching block details: %v", err)
		}

		// Wait for some time before fetching the next block (adjust as needed)
		time.Sleep(1 * time.Second)
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

	// Publish to RabbitMQ
	err = publishToRabbitMQ(blockDetails)
	if err != nil {
		log.Printf("Failed to publish to RabbitMQ: %v", err)
		return err
	}

	return nil
}

func publishToRabbitMQ(details map[string]interface{}) error {
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

	// Publish the JSON message to RabbitMQ
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

	// Store transaction details
	for _, transaction := range details.TransactionDetails {
		err = storeTransactionInMySQL(transaction, db)
		if err != nil {
			log.Printf("Failed to store transaction in MySQL: %v", err)
			return err
		}
	}

	return nil
}

func storeTransactionInMySQL(transaction TransactionDetails, db *sql.DB) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Your MySQL insertion logic for transaction details
	stmtTransaction, err := db.Prepare(`
		INSERT INTO transaction_details (block_hash, block_number, chain_id, from_address, gas, gas_price, hash, max_fee_per_gas, max_priority_fee_per_gas, nonce, to_address, transaction_index, value)
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
		transaction.From,
		transaction.Gas,
		transaction.GasPrice,
		transaction.Hash,
		transaction.MaxFeePerGas,
		transaction.MaxPriorityFeePerGas,
		transaction.Nonce,
		transaction.To,
		transaction.TransactionIndex,
		transaction.Value,
	)
	if err != nil {
		return err
	}

	log.Printf("Successfully inserted transaction details into MySQL for transaction hash %s", transaction.Hash)
	return nil
}
