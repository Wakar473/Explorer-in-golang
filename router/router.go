/*
 *	Author: Puneet
 *	Use to start the client
 */
package router

import (
	"log"
	"net/http"

	"boilerplate/utils"
	"github.com/gin-gonic/gin"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc func(*gin.Context)
}
type routes struct {
	router *gin.Engine
}

type Routes []Route

/*
 *	Function for grouping esg health routes
 */
// func (r routes) ESGHealthGrouping(rg *gin.RouterGroup) {
// 	orderRouteGrouping := rg.Group("/esg")
// 	orderRouteGrouping.Use(CORSMiddleware())
// 	for _, route := range healthCheckRoutes {
// 		switch route.Method {
// 		case "GET":
// 			orderRouteGrouping.GET(route.Pattern, route.HandlerFunc)
// 		case "POST":
// 			orderRouteGrouping.POST(route.Pattern, route.HandlerFunc)
// 		case "OPTIONS":
// 			orderRouteGrouping.OPTIONS(route.Pattern, route.HandlerFunc)
// 		case "PUT":
// 			orderRouteGrouping.PUT(route.Pattern, route.HandlerFunc)
// 		case "DELETE":
// 			orderRouteGrouping.DELETE(route.Pattern, route.HandlerFunc)
// 		default:
// 			orderRouteGrouping.GET(route.Pattern, func(c *gin.Context) {
// 				c.JSON(200, gin.H{
// 					"result": "Specify a valid http method with this route.",
// 				})
// 			})
// 		}
	// }
// }


func (r routes) GetBlocksGrouping(rg *gin.RouterGroup) {
	orderRouteGrouping := rg.Group("/current")
	orderRouteGrouping.Use(CORSMiddleware())
	for _, route := range getBlocksRoutes {
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
	r.GetBlocksGrouping(v1)

	// currentBlock := r.router.GET("/current-block",)

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
