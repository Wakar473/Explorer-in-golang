package router

import (
	"net/http"

	"boilerplate/constant"
	"boilerplate/controller"
)

// health check service
var healthCheckRoutes = Routes{
	Route{"Health check", http.MethodGet, constant.HealthCheckRoute, controller.HealthCheck},
}

var getBlocksRoutes = Routes{
	Route{"Get Blocks", http.MethodGet, constant.GetBlocksRoute, controller.GetBlocks},
}
