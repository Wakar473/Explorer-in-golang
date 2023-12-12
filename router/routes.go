package router

import (
	// "net/http"
	"github.com/gin-gonic/gin"

	// "boilerplate/constant"
	// "boilerplate/controller"
)

var healthCheckRoutes = Routes{
	Route{
		Name:    "blockNumber",
		Method:  "http.MethodGet",
		Pattern: "constant.HealthCheckRoute",
		HandlerFunc: func(*gin.Context) {
		},
	},
}

// "Health check", http.MethodGet, constant.HealthCheckRoute, controller.HealthCheck,},
