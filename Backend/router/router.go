package router

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"dataextractor/controller"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// SetupRoutes configures all the API routes
func SetupRoutes() *gin.Engine {
	// Create Gin router without default middleware
	router := gin.New()

	// Add logger middleware
	router.Use(gin.Logger())

	// Add custom recovery middleware to handle panics gracefully
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		fmt.Printf("=== RECOVERY MIDDLEWARE TRIGGERED ===\n")
		fmt.Printf("Recovered value: %v (type: %T)\n", recovered, recovered)
		statusCode := http.StatusInternalServerError
		errorType := "Internal server error"
		details := "An unexpected error occurred"

		// Handle different types of recovered values
		switch err := recovered.(type) {
		case error:
			// Check for specific GORM errors first
			if errors.Is(err, gorm.ErrRecordNotFound) {
				statusCode = http.StatusNotFound
				errorType = "Resource not found"
				details = err.Error()
			} else if errors.Is(err, gorm.ErrInvalidData) {
				statusCode = http.StatusBadRequest
				errorType = "Invalid data"
				details = err.Error()
			} else if errors.Is(err, gorm.ErrInvalidTransaction) {
				statusCode = http.StatusBadRequest
				errorType = "Invalid transaction"
				details = err.Error()
			} else {
				// Check error message for common patterns
				errMsg := err.Error()
				if contains(errMsg, "not found") || contains(errMsg, "record not found") {
					statusCode = http.StatusNotFound
					errorType = "Resource not found"
				} else if contains(errMsg, "invalid") || contains(errMsg, "validation") {
					statusCode = http.StatusBadRequest
					errorType = "Invalid request"
				} else if contains(errMsg, "unauthorized") || contains(errMsg, "forbidden") {
					statusCode = http.StatusUnauthorized
					errorType = "Unauthorized"
				}
				details = errMsg
			}
		case string:
			// Handle string errors
			errMsg := err
			if contains(errMsg, "not found") || contains(errMsg, "record not found") {
				statusCode = http.StatusNotFound
				errorType = "Resource not found"
			} else if contains(errMsg, "invalid") || contains(errMsg, "validation") {
				statusCode = http.StatusBadRequest
				errorType = "Invalid request"
			} else if contains(errMsg, "unauthorized") || contains(errMsg, "forbidden") {
				statusCode = http.StatusUnauthorized
				errorType = "Unauthorized"
			}
			details = errMsg
		default:
			// Handle any other type by converting to string
			errMsg := fmt.Sprintf("%v", recovered)
			if contains(errMsg, "not found") || contains(errMsg, "record not found") {
				statusCode = http.StatusNotFound
				errorType = "Resource not found"
			} else if contains(errMsg, "invalid") || contains(errMsg, "validation") {
				statusCode = http.StatusBadRequest
				errorType = "Invalid request"
			} else if contains(errMsg, "unauthorized") || contains(errMsg, "forbidden") {
				statusCode = http.StatusUnauthorized
				errorType = "Unauthorized"
			}
			details = errMsg
		}

		// Log the error for debugging
		fmt.Printf("Recovery middleware caught panic: %v\n", recovered)
		fmt.Printf("Status code: %d, Error type: %s, Details: %s\n", statusCode, errorType, details)

		c.JSON(statusCode, gin.H{
			"error":   errorType,
			"details": details,
		})
		c.Abort()
	}))

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// Create stock controller
	stockController := controller.NewStockController()

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Stock routes
		stocks := v1.Group("/stocks")
		{
			// CRUD operations
			stocks.POST("", stockController.CreateStock)       // POST /api/v1/stocks
			stocks.GET("", stockController.GetAllStocks)       // GET /api/v1/stocks
			
			// Table management operations - must come before /:id routes to avoid conflicts
			stocks.DELETE("/tables", stockController.EmptyAllTables) // DELETE /api/v1/stocks/tables
			
			// CRUD operations with ID - placed after specific routes
			stocks.GET("/:id", stockController.GetStockByID)   // GET /api/v1/stocks/:id
			stocks.PUT("/:id", stockController.UpdateStock)    // PUT /api/v1/stocks/:id
			stocks.DELETE("/:id", stockController.DeleteStock) // DELETE /api/v1/stocks/:id

			// Find operations
			stocks.GET("/ticker/:ticker", stockController.GetStockByTicker)                // GET /api/v1/stocks/ticker/:ticker
			stocks.GET("/company/:company", stockController.GetStocksByCompany)            // GET /api/v1/stocks/company/:company
			stocks.GET("/clusters", stockController.GetUniqueClusters)                     // GET /api/v1/stocks/clusters
			stocks.GET("/cluster/:cluster", stockController.GetStocksByCluster)                  // GET /api/v1/stocks/cluster/:cluster
			stocks.GET("/cluster/:cluster/filter", stockController.FilterByClusterGrouped)       // GET /api/v1/stocks/cluster/:cluster/filter
			stocks.GET("/cluster/:cluster/unique/:column_name", stockController.GetUniqueByGroupSelectColumn) // GET /api/v1/stocks/cluster/:cluster/unique/:column_name
			stocks.GET("/actions", stockController.GetUniqueActions)                             // GET /api/v1/stocks/actions
			stocks.GET("/action/:action", stockController.GetStocksByAction)                     // GET /api/v1/stocks/action/:action

			// Statistics operations
			stocks.GET("/stats/:ticker", stockController.GetStockStats)     // GET /api/v1/stocks/stats/:ticker
			stocks.GET("/database/stats", stockController.GetDatabaseStats) // GET /api/v1/stocks/database/stats

			// Data extraction operations
			stocks.POST("/extract", stockController.ExtractDataFromApi)        // POST /api/v1/stocks/extract
			stocks.POST("/import-enriched", stockController.ImportEnrichedCSV) // POST /api/v1/stocks/import-enriched
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Stock API is running",
		})
	})

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Stock Data Extractor API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"health":  "/health",
				"api":     "/api/v1/stocks",
				"extract": "/api/v1/stocks/extract",
				"swagger": "/swagger/index.html",
			},
		})
	})

	return router
}

// NewRouter creates a new router with the provided controller
func NewRouter(stockController *controller.StockController) http.Handler { return SetupRoutes() }

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
