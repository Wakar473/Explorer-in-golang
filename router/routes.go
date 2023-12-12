package router

import (
	"net/http"

	"boilerplate/constant"
	"boilerplate/controller"
)

// var healthCheckRoutes = Routes{
// 	Route{"Health check", http.MethodGet, constant.HealthCheckRoute, controller.HealthCheck},
// }
// var healthCheck = Routes{
// 	Route{
// 		Name:    "blockNumber",
// 		Method:  "http.MethodGet",
// 		Pattern: "constant.HealthCheckRoute",
// 		HandlerFunc: func(*gin.Context) {
// 		},
// 	},
// }

var GetBlockDetailsFromDb = Routes{
	Route{"Get Block Details", http.MethodGet, constant.GetLatestBlock, controller.FetchBlocks},
}

// "Health check", http.MethodGet, constant.HealthCheckRoute, controller.HealthCheck,},
