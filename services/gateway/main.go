package main

import (
	"context"
	"log"

	"social-network/services/common/httpserver"
	"social-network/services/common/redisstore"
	"social-network/services/gateway/edge"
	"social-network/services/gateway/proxy"
)

func main() {
	targets, err := proxy.TargetsFromEnvironment()
	if err != nil {
		log.Fatalf("Invalid gateway configuration: %v", err)
	}
	edgeConfig, err := edge.ConfigFromEnvironment()
	if err != nil {
		log.Fatalf("Invalid gateway edge-control configuration: %v", err)
	}
	redisConfig, err := redisstore.FromEnvironment()
	if err != nil {
		log.Fatalf("Invalid Redis configuration: %v", err)
	}
	redisState, err := redisstore.Open(context.Background(), redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisState.Close()
	controls := edge.NewDistributed(edgeConfig, edge.NewRedisLimiter(redisState, edgeConfig))

	address := httpserver.Address("8080")
	log.Printf("Gateway starting on %s", address)
	if err := httpserver.Run(httpserver.New(address, proxy.NewRouter(targets, controls))); err != nil {
		log.Fatalf("Gateway stopped with error: %v", err)
	}
}
