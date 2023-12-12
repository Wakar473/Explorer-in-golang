package controller

import (
	"boilerplate/database"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var db *sql.DB

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
	println("here1")
	database.ConnectDb()
	println("here2", db)

	db, _ := sql.Open("mysql", "root:@tcp(localhost:3306)/saita")
	println("here2", db)
	rows, err := db.Query("SELECT block_number, parent_hash, block_hash, timestamp, transaction_count FROM block_details limit 10")
	if err != nil {
		println("here2")

		log.Printf("Failed to query the database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query the database"})
		return
	}
	// println("rows===============> ", rows)

	defer rows.Close()
	println("here4")

	for rows.Next() {
		var block BlockDetails
		err := rows.Scan(&block.Number, &block.ParentHash, &block.BlockHash, &block.Timestamp, &block.Transactions)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve block details"})
			return
		}
		// log.Printf("block ==============> ", block)
		blocks = append(blocks, block)
	}
	println("blocks ===========>", blocks)

	jsonData, err := json.Marshal(blocks)
	if err != nil {
		log.Printf("Failed to marshal block details to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal block details to JSON"})
		return
	}
	println("here6")

	c.JSON(http.StatusOK, string(jsonData))
}
