// @title Stock Data Extractor API
// @version 1.0
// @description A RESTful API for managing stock data with full CRUD operations
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8888
// @BasePath /
// @schemes http https

package main

import (
	"log"
	"net/http"
	"os"

	_ "dataextractor/docs"
	"dataextractor/router"
	"dataextractor/utils"
)

func main() {
	// Create routes
	routes := router.SetupRoutes()

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8887"
	}

	// Create server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: routes,
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("API Documentation available at: http://localhost:%s", port)
	log.Printf("Health check available at: http://localhost:%s/health", port)

	// Start server
	err := server.ListenAndServe()
	utils.ErrorPanic(err, "Failed to start server")
}
