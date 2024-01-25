/*
 *	Author: Puneet
 *	Use to start the client
 */
package router

import (
	// "database/sql"
	// "encoding/json"
	"log"
	"net/http"

	"boilerplate/utils"

	"github.com/gin-gonic/gin"
)

// var db *sql.DB

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc func(*gin.Context)
}

//	type BlockDetails struct {
//		Number       string `json:"number"`
//		ParentHash   string `json:"parentHash"`
//		BlockHash    string `json:"hash"`
//		Timestamp    string `json:"timestamp"`
//		Transactions int    `json:"transactions"`
//		HandlerFunc  func(*gin.Context)
//	}
type routes struct {
	router *gin.Engine
}

// type: This keyword indicates that you are defining a new type.
// Routes: This is the name of the new type.
// []: This denotes that it is a slice. A slice is a data structure that can hold a variable-length sequence of elements of the same type.
// Route: This specifies the type of elements that the slice can hold. In this case, each element in the Routes slice must be of type Route.
// Therefore, a variable of type Routes can be declared and used to store a collection of Route objects.
type Routes []Route

/*
 *	Function for grouping esg health routes
 */
// func (r routes) ESGHealthGrouping(rg *gin.RouterGroup) {

/*
This code defines a method named ESGHealthGrouping for a type named routes. The method takes a pointer to a gin.RouterGroup as input and uses it to create a new router group specific to ESG Health. Let's break down the code line by line:
1. Method definition:
func (r routes) ESGHealthGrouping(rg *gin.RouterGroup) { ... }: This line defines a method named ESGHealthGrouping for a type named routes. The method takes a pointer to a gin.RouterGroup as input.
2. Group creation:
orderRouteGrouping := rg.Group("/esg"): This line creates a new router group within the existing rg group. The new group is named "/esg" and will be responsible for handling all API endpoints related to ESG Health.
3. Middleware usage:
orderRouteGrouping.Use(CORSMiddleware()): This line applies the CORSMiddleware function to the newly created /esg group. This middleware is likely responsible for handling Cross-Origin Resource Sharing (CORS) requests and ensuring that the API endpoints are accessible from different origins.
*/
func Fetched() {
	// r := gin.Default()

	// r.GET("/blockDetails", func(c *gin.Context) {
	// 	var blocks []BlockDetails

	// 	rows, err := db.Query("SELECT block_number, parent_hash, block_hash, timestamp, transaction_count FROM block_details ORDER BY block_number DESC LIMIT 10")
	// 	if err != nil {
	// 		log.Printf("Failed to query the database: %v", err)
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query the database"})
	// 		return
	// 	}
	// 	defer rows.Close()

	// 	for rows.Next() {
	// 		var block BlockDetails
	// 		err := rows.Scan(&block.Number, &block.ParentHash, &block.BlockHash, &block.Timestamp, &block.Transactions)
	// 		if err != nil {
	// 			log.Printf("Failed to scan row: %v", err)
	// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve block details"})
	// 			return
	// 		}
	// 		blocks = append(blocks, block)
	// 	}

	// 	jsonData, err := json.Marshal(blocks)
	// 	if err != nil {
	// 		log.Printf("Failed to marshal block details to JSON: %v", err)
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal block details to JSON"})
	// 		return
	// 	}

	// 	c.JSON(http.StatusOK, string(jsonData))
	// })
}
func (r routes) ESGHealthGrouping(rg *gin.RouterGroup) {

	orderRouteGrouping := rg.Group("/esg")
	orderRouteGrouping.Use(CORSMiddleware())
	// for _, route := range healthCheckRoutes {
	// 	switch route.Method {
	// 	case "GET":
	// 		orderRouteGrouping.GET(route.Pattern,route.HandlerFunc)
	// 	case "POST":
	// 		orderRouteGrouping.POST(route.Pattern,route.HandlerFunc )
	// 	case "OPTIONS":
	// 		orderRouteGrouping.OPTIONS(route.Pattern,route.HandlerFunc )
	// 	case "PUT":
	// 		orderRouteGrouping.PUT(route.Pattern,route.HandlerFunc)
	// 	case "DELETE":
	// 		orderRouteGrouping.DELETE(route.Pattern,route.HandlerFunc)
	// 	default:
	// 		orderRouteGrouping.GET(route.Pattern, func(c *gin.Context) {
	// 			c.JSON(200, gin.H{
	// 				"result": "Specify a valid http method with this route.",
	// 			})
	// 		})
	// 	}
	// }

	for _, route := range GetBlockDetailsFromDb {
		switch route.Method {
		case "GET":
			orderRouteGrouping.GET(route.Pattern, route.HandlerFunc)
		case "POST":
			orderRouteGrouping.POST(route.Pattern, route.HandlerFunc)
		case "OPTIONS":
			orderRouteGrouping.OPTIONS(route.Pattern, route.HandlerFunc)
		case "PUT":
			orderRouteGrouping.PUT(route.Pattern, route.HandlerFunc)
		case "DELETE":
			orderRouteGrouping.DELETE(route.Pattern, route.HandlerFunc)
		default:
			orderRouteGrouping.GET(route.Pattern, func(c *gin.Context) {
				c.JSON(200, gin.H{
					"result": "Specify a valid http method with this route.",
				})
			})
		}
	}



	for _, route := range GetTransactionDetailsFromDb {
		switch route.Method {
		case "GET":
			orderRouteGrouping.GET(route.Pattern, route.HandlerFunc)
		case "POST":
			orderRouteGrouping.POST(route.Pattern, route.HandlerFunc)
		case "OPTIONS":
			orderRouteGrouping.OPTIONS(route.Pattern, route.HandlerFunc)
		case "PUT":
			orderRouteGrouping.PUT(route.Pattern, route.HandlerFunc)
		case "DELETE":
			orderRouteGrouping.DELETE(route.Pattern, route.HandlerFunc)
		default:
			orderRouteGrouping.GET(route.Pattern, func(c *gin.Context) {
				c.JSON(200, gin.H{
					"result": "Specify a valid http method with this route.",
				})
			})
		}
	}

}

// append routes with versions
func ClientRoutes() {
	r := routes{
		router: gin.Default(),
	}
	v1 := r.router.Group(utils.GetAPIVersion())
	r.ESGHealthGrouping(v1)

	if err := r.router.Run(":" + utils.GetClientPort()); err != nil {
		log.Printf("Failed to run server: %v", err)
	}

}

// Middlewares
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusOK)
		}
	}
}
