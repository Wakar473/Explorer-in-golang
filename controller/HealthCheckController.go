package controller

import (
	"boilerplate/database"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// var db *sql.DB

type BlockDetails struct {
	Number       string `json:"number"`
	ParentHash   string `json:"parentHash"`
	BlockHash    string `json:"hash"`
	Timestamp    string `json:"timestamp"`
	Transactions int    `json:"transactions"`
	// HandlerFunc  func(*gin.Context)
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func FetchBlocks(c *gin.Context) {
	var blocks []BlockDetails
	database.ConnectDb()

	db, _ := sql.Open("mysql", "root:@tcp(localhost:3306)/saita")
	rows, err := db.Query("SELECT block_number, parent_hash, block_hash, timestamp, transaction_count FROM block_details ORDER BY block_number DESC LIMIT 10")
	if err != nil {
		log.Printf("Failed to query the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query the database"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var block BlockDetails
		var blockNumberHex string // Assuming block_number is stored in hexadecimal format

		err := rows.Scan(&blockNumberHex, &block.ParentHash, &block.BlockHash, &block.Timestamp, &block.Transactions)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve block details"})
			return
		}

		// Convert hexadecimal block number to uint64
		blockNumber, err := strconv.ParseUint(strings.TrimPrefix(blockNumberHex, "0x"), 16, 64)
		if err != nil {
			log.Printf("Failed to convert block number: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert block number"})
			return
		}

		// Assign the converted block number to the BlockDetails struct
		block.Number = strconv.FormatUint(blockNumber, 10) // Store as string if needed
		blocks = append(blocks, block)
	}

	c.JSON(http.StatusOK, blocks)
}
