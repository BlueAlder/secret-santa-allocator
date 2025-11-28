package main

import (
	"log"
)

func main() {

	server, err := createServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	server.logger.Info("server starting", "host", server.host, "port", server.port, "storage_type", server.storageType)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
