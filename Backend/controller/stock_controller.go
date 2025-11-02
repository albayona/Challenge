package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"dataextractor/repository"
	"dataextractor/service"
	"dataextractor/utils"
	"dataextractor/validators"

	"github.com/gin-gonic/gin"
)

// StockController handles HTTP requests for stock operations
type StockController struct {
	stockService service.StockServiceInterface
}

// NewStockController creates a new StockController instance
func NewStockController() *StockController {
	// Create repository factory
	repoFactory := repository.NewRepositoryFactory()
	repo := repoFactory.CreateDataRepository()

	// Create stock service
	stockService := service.NewStockService(repo)

	return &StockController{
		stockService: stockService,
	}
}

// CreateStock handles POST /stocks
// @Summary Create a new stock
// @Description Create a new stock record with the provided information
// @Tags stocks
// @Accept json
// @Produce json
// @Param stock body validators.StockCreateRequest true "Stock information"
// @Success 201 {object} map[string]interface{} "Stock created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request format"
// @Failure 500 {object} map[string]interface{} "Failed to create stock"
// @Router /api/v1/stocks [post]
func (sc *StockController) CreateStock(c *gin.Context) {
	var request validators.StockCreateRequest

	// Bind JSON request to StockCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Create stock using service
	stock, err := sc.stockService.Create(&request)
	utils.ErrorPanic(err, "failed to create stock")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Stock created successfully",
		"data":    stock,
	})
}

// GetStockByID handles GET /stocks/:id
// @Summary Get stock by ID
// @Description Retrieve a specific stock record by its ID
// @Tags stocks
// @Produce json
// @Param id path int true "Stock ID"
// @Success 200 {object} map[string]interface{} "Stock found"
// @Failure 400 {object} map[string]interface{} "Invalid stock ID"
// @Failure 404 {object} map[string]interface{} "Stock not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve stock"
// @Router /api/v1/stocks/{id} [get]
func (sc *StockController) GetStockByID(c *gin.Context) {
	// Parse ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a valid number",
		})
		return
	}

	// Get stock by ID
	stock, err := sc.stockService.GetByID(uint(id))
	utils.ErrorPanic(err, "failed to get stock by ID")

	c.JSON(http.StatusOK, gin.H{
		"data": stock,
	})
}

// GetAllStocks handles GET /stocks
// @Summary Get all stocks
// @Description Retrieve all stock records from the database
// @Tags stocks
// @Produce json
// @Success 200 {object} map[string]interface{} "List of stocks"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve stocks"
// @Router /api/v1/stocks [get]
func (sc *StockController) GetAllStocks(c *gin.Context) {
	// Get all stocks
	stocks, err := sc.stockService.GetAll()
	utils.ErrorPanic(err, "failed to get all stocks")

	c.JSON(http.StatusOK, gin.H{
		"data":  stocks,
		"count": len(stocks),
	})
}

// UpdateStock handles PUT /stocks/:id
// @Summary Update stock by ID
// @Description Update an existing stock record with the provided information
// @Tags stocks
// @Accept json
// @Produce json
// @Param id path int true "Stock ID"
// @Param stock body validators.StockUpdateRequest true "Updated stock information"
// @Success 200 {object} map[string]interface{} "Stock updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request format"
// @Failure 404 {object} map[string]interface{} "Stock not found"
// @Failure 500 {object} map[string]interface{} "Failed to update stock"
// @Router /api/v1/stocks/{id} [put]
func (sc *StockController) UpdateStock(c *gin.Context) {
	// Parse ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a valid number",
		})
		return
	}

	var request validators.StockUpdateRequest

	// Bind JSON request to StockUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set the ID from URL parameter
	request.ID = uint(id)

	// Update stock using service
	stock, err := sc.stockService.Update(&request)
	utils.ErrorPanic(err, "failed to update stock")

	c.JSON(http.StatusOK, gin.H{
		"message": "Stock updated successfully",
		"data":    stock,
	})
}

// DeleteStock handles DELETE /stocks/:id
// @Summary Delete stock by ID
// @Description Delete a specific stock record by its ID
// @Tags stocks
// @Produce json
// @Param id path int true "Stock ID"
// @Success 200 {object} map[string]interface{} "Stock deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid stock ID"
// @Failure 404 {object} map[string]interface{} "Stock not found"
// @Failure 500 {object} map[string]interface{} "Failed to delete stock"
// @Router /api/v1/stocks/{id} [delete]
func (sc *StockController) DeleteStock(c *gin.Context) {
	// Parse ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a valid number",
		})
		return
	}

	// Delete stock using service
	err = sc.stockService.Delete(uint(id))
	utils.ErrorPanic(err, "failed to delete stock")

	c.JSON(http.StatusOK, gin.H{
		"message": "Stock deleted successfully",
	})
}

// GetStockByTicker handles GET /stocks/ticker/:ticker
// @Summary Get stock by ticker
// @Description Retrieve a specific stock record by its ticker symbol
// @Tags stocks
// @Produce json
// @Param ticker path string true "Stock ticker symbol"
// @Success 200 {object} map[string]interface{} "Stock found"
// @Failure 400 {object} map[string]interface{} "Invalid ticker format"
// @Failure 404 {object} map[string]interface{} "Stock not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve stock"
// @Router /api/v1/stocks/ticker/{ticker} [get]
func (sc *StockController) GetStockByTicker(c *gin.Context) {
	// Get ticker from URL parameter
	ticker := c.Param("ticker")
	if ticker == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Ticker parameter is required",
			"details": "Ticker cannot be empty",
		})
		return
	}

	// Get stock by ticker
	stock, err := sc.stockService.GetByTicker(ticker)
	utils.ErrorPanic(err, "failed to get stock by ticker")

	c.JSON(http.StatusOK, gin.H{
		"data": stock,
	})
}

// GetStocksByCompany handles GET /stocks/company/:company
// @Summary Get stocks by company
// @Description Retrieve all stock records for a specific company
// @Tags stocks
// @Produce json
// @Param company path string true "Company name"
// @Success 200 {object} map[string]interface{} "List of stocks for company"
// @Failure 400 {object} map[string]interface{} "Invalid company name"
// @Failure 404 {object} map[string]interface{} "No stocks found for company"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve stocks"
// @Router /api/v1/stocks/company/{company} [get]
func (sc *StockController) GetStocksByCompany(c *gin.Context) {
	// Get company from URL parameter
	company := c.Param("company")
	if company == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Company parameter is required",
			"details": "Company cannot be empty",
		})
		return
	}

	// Get stocks by company
	stocks, err := sc.stockService.GetByCompany(company)
	utils.ErrorPanic(err, "failed to get stocks by company")

	c.JSON(http.StatusOK, gin.H{
		"data":  stocks,
		"count": len(stocks),
	})
}

// GetUniqueClusters handles GET /stocks/clusters
// @Summary Get unique clusters
// @Description Retrieve all unique cluster IDs
// @Tags stocks
// @Produce json
// @Success 200 {object} map[string]interface{} "List of unique clusters"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve clusters"
// @Router /api/v1/stocks/clusters [get]
func (sc *StockController) GetUniqueClusters(c *gin.Context) {
	clusters, err := sc.stockService.GetUniqueClusters()
	utils.ErrorPanic(err, "failed to get unique clusters")
	c.JSON(http.StatusOK, gin.H{
		"data":  clusters,
		"count": len(clusters),
	})
}

// GetStocksByCluster handles GET /stocks/cluster/:cluster
// @Summary Get stocks by cluster
// @Description Retrieve all stock records for a specific cluster
// @Tags stocks
// @Produce json
// @Param cluster path int true "Cluster id"
// @Success 200 {object} map[string]interface{} "List of stocks for cluster"
// @Failure 400 {object} map[string]interface{} "Invalid cluster"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve stocks"
// @Router /api/v1/stocks/cluster/{cluster} [get]
func (sc *StockController) GetStocksByCluster(c *gin.Context) {
	clusterStr := c.Param("cluster")
	cluster, err := strconv.Atoi(clusterStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid cluster parameter",
			"details": "Cluster must be an integer",
		})
		return
	}

	stocks, err := sc.stockService.GetStocksByCluster(cluster)
	utils.ErrorPanic(err, "failed to get stocks by cluster")
	c.JSON(http.StatusOK, gin.H{
		"data":  stocks,
		"count": len(stocks),
	})
}

// GetUniqueCompanies handles GET /stocks/companies
// @Summary Get unique companies
// @Description Retrieve all unique company names
// @Tags stocks
// @Produce json
// @Success 200 {object} map[string]interface{} "List of unique companies"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve companies"
// @Router /api/v1/stocks/companies [get]
func (sc *StockController) GetUniqueCompanies(c *gin.Context) {
	companies, err := sc.stockService.GetUniqueCompanies()
	utils.ErrorPanic(err, "failed to get unique companies")
	c.JSON(http.StatusOK, gin.H{
		"data":  companies,
		"count": len(companies),
	})
}

// GetUniqueActions handles GET /stocks/actions
// @Summary Get unique actions
// @Description Retrieve all unique action values
// @Tags stocks
// @Produce json
// @Success 200 {object} map[string]interface{} "List of unique actions"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve actions"
// @Router /api/v1/stocks/actions [get]
func (sc *StockController) GetUniqueActions(c *gin.Context) {
	actions, err := sc.stockService.GetUniqueActions()
	utils.ErrorPanic(err, "failed to get unique actions")
	c.JSON(http.StatusOK, gin.H{
		"data":  actions,
		"count": len(actions),
	})
}

// GetStocksByAction handles GET /stocks/action/:action
// @Summary Get stocks by action
// @Description Retrieve all stock records for a specific action
// @Tags stocks
// @Produce json
// @Param action path string true "Action value"
// @Success 200 {object} map[string]interface{} "List of stocks for action"
// @Failure 400 {object} map[string]interface{} "Invalid action"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve stocks"
// @Router /api/v1/stocks/action/{action} [get]
func (sc *StockController) GetStocksByAction(c *gin.Context) {
	action := c.Param("action")
	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Action parameter is required",
			"details": "Action cannot be empty",
		})
		return
	}

	stocks, err := sc.stockService.GetStocksByAction(action)
	utils.ErrorPanic(err, "failed to get stocks by action")
	c.JSON(http.StatusOK, gin.H{
		"data":  stocks,
		"count": len(stocks),
	})
}

// GetStockStats handles GET /stocks/stats/:ticker
// @Summary Get stock statistics by ticker
// @Description Retrieve statistical information for a specific stock ticker
// @Tags stocks
// @Produce json
// @Param ticker path string true "Stock ticker symbol"
// @Success 200 {object} map[string]interface{} "Stock statistics"
// @Failure 400 {object} map[string]interface{} "Invalid ticker format"
// @Failure 404 {object} map[string]interface{} "Stock not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve statistics"
// @Router /api/v1/stocks/stats/{ticker} [get]
func (sc *StockController) GetStockStats(c *gin.Context) {
	// Get ticker from URL parameter
	ticker := c.Param("ticker")
	if ticker == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Ticker parameter is required",
			"details": "Ticker cannot be empty",
		})
		return
	}

	// Get stock statistics
	stats, err := sc.stockService.GetStats(ticker)
	utils.ErrorPanic(err, "failed to get stock statistics")

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// GetDatabaseStats handles GET /stocks/database/stats
// @Summary Get database statistics
// @Description Retrieve statistical information about the database
// @Tags stocks
// @Produce json
// @Success 200 {object} map[string]interface{} "Database statistics"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve database statistics"
// @Router /api/v1/stocks/database/stats [get]
func (sc *StockController) GetDatabaseStats(c *gin.Context) {
	// Get database statistics
	stats, err := sc.stockService.GetDatabaseStats()
	utils.ErrorPanic(err, "failed to get database statistics")

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// ExtractDataFromApi handles POST /stocks/extract
// @Summary Extract data from API
// @Description Trigger data extraction from external API with specified max pages
// @Tags stocks
// @Accept json
// @Produce json
// @Param request body validators.StockExtractRequest true "Extraction request"
// @Success 200 {object} map[string]interface{} "Data extraction completed"
// @Failure 400 {object} map[string]interface{} "Invalid request format"
// @Failure 500 {object} map[string]interface{} "Failed to extract data from API"
// @Router /api/v1/stocks/extract [post]
func (sc *StockController) ExtractDataFromApi(c *gin.Context) {
	var request validators.StockExtractRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Extract data from API using service
	err := sc.stockService.StoreDataFromApi(request.MaxPages)
	utils.ErrorPanic(err, "failed to extract data from API")

	c.JSON(http.StatusOK, gin.H{
		"message":   "Data extraction completed successfully",
		"max_pages": request.MaxPages,
		"status":    "completed",
	})
}

// ImportEnrichedCSV handles POST /stocks/import-enriched
// @Summary Import enriched stock data from default CSV
// @Description Import rows from ./stock_data_enriched.csv into the database
// @Tags stocks
// @Produce json
// @Success 200 {object} map[string]interface{} "CSV imported"
// @Failure 500 {object} map[string]interface{} "Failed to import CSV"
// @Router /api/v1/stocks/import-enriched [post]
func (sc *StockController) ImportEnrichedCSV(c *gin.Context) {
	count, err := sc.stockService.ImportFromEnrichedCSV()
	utils.ErrorPanic(err, "failed to import enriched CSV")
	c.JSON(http.StatusOK, gin.H{
		"message":       "Enriched CSV imported successfully",
		"rows_ingested": count,
	})
}

// FilterByClusterGrouped handles GET /stocks/cluster/:cluster/filter
// @Summary Filter stocks by cluster with grouping, pagination, sorting, and weighted scoring
// @Description Filter stocks by cluster with optional grouping, pagination, sorting, and weighted scoring. Supports numerical and rating weights via query parameters. Note: grouping_column can only be action, rating_to, or rating_from (company and date are excluded due to too many distinct values).
// @Tags stocks
// @Produce json
// @Param cluster path int true "Cluster id"
// @Param grouping_column query string false "Grouping column: action | rating_to | rating_from | None (default: None). Note: company and date are excluded."
// @Param grouping_value query string false "Grouping value to filter by (required if grouping_column is not None)"
// @Param sort_by query string false "Sort by column: ticker | action | date | company | target_to | target_from | rating_to | rating_from | final_score (default: date)"
// @Param order query string false "Sort order: asc | desc (default: desc)"
// @Param page query int false "Page number (default: 1)"
// @Param per_page query int false "Items per page (default: 20)"
// @Param numerical_weights query string false "JSON array of numerical weights: [{\"indicator_name\":\"atr\",\"weight\":0.5}]"
// @Param rating_weights query string false "JSON array of rating weights: [{\"indicator_name\":\"action\",\"weight\":0.7}]"
// @Success 200 {object} map[string]interface{} "Paged grouped results"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Failed to filter"
// @Router /api/v1/stocks/cluster/{cluster}/filter [get]
func (sc *StockController) FilterByClusterGrouped(c *gin.Context) {
	// Parse cluster from path
	clusterStr := c.Param("cluster")
	cluster, err := strconv.Atoi(clusterStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid cluster parameter",
			"details": "Cluster must be an integer",
		})
		return
	}

	// Parse query parameters with defaults
	groupingColumn := c.DefaultQuery("grouping_column", "None")
	groupingValue := c.Query("grouping_value")
	sortByColumn := c.DefaultQuery("sort_by", "date")
	order := strings.ToLower(c.DefaultQuery("order", "desc"))

	// Parse pagination with defaults
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	perPage := 20
	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}

	// Parse numerical weights from query parameter (URL-encoded JSON array)
	var numericalWeights []repository.NumericalWeightEntry
	if numericalWeightsStr := c.Query("numerical_weights"); numericalWeightsStr != "" {
		var weights []struct {
			IndicatorName string  `json:"indicator_name"`
			Weight        float64 `json:"weight"`
		}
		if err := json.Unmarshal([]byte(numericalWeightsStr), &weights); err == nil {
			numericalWeights = make([]repository.NumericalWeightEntry, len(weights))
			for i, w := range weights {
				numericalWeights[i] = repository.NumericalWeightEntry{
					IndicatorName: w.IndicatorName,
					Weight:        w.Weight,
				}
			}
		}
	}

	// Parse rating weights from query parameter (URL-encoded JSON array)
	var ratingWeights []repository.RatingWeightEntry
	if ratingWeightsStr := c.Query("rating_weights"); ratingWeightsStr != "" {
		var weights []struct {
			IndicatorName string  `json:"indicator_name"`
			Weight        float64 `json:"weight"`
		}
		if err := json.Unmarshal([]byte(ratingWeightsStr), &weights); err == nil {
			ratingWeights = make([]repository.RatingWeightEntry, len(weights))
			for i, w := range weights {
				ratingWeights[i] = repository.RatingWeightEntry{
					IndicatorName: w.IndicatorName,
					Weight:        w.Weight,
				}
			}
		}
	}

	// Call service
	result, err := sc.stockService.FilterByClusterGrouped(cluster, groupingColumn, groupingValue, sortByColumn, order, page, perPage, numericalWeights, ratingWeights)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to filter stocks",
			"details": err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"data":            result.Items,
		"total_count":     result.TotalCount,
		"page":            result.Page,
		"per_page":        result.PerPage,
		"grouping_column": groupingColumn,
		"grouping_value":  groupingValue,
		"sort_by":         sortByColumn,
		"order":           order,
	})
}

// GetUniqueByGroupSelectColumn handles GET /stocks/cluster/:cluster/unique/:column_name
// @Summary Get unique values for a specified column filtered by cluster
// @Description Get unique values for a column from StockDataPoint filtered by cluster. Allowed columns: action, rating_to, rating_from. Note: company and date are excluded due to having too many distinct values.
// @Tags stocks
// @Produce json
// @Param cluster path int true "Cluster id"
// @Param column_name path string true "Column name: action | rating_to | rating_from"
// @Success 200 {object} map[string]interface{} "Unique values"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Failed to get unique values"
// @Router /api/v1/stocks/cluster/{cluster}/unique/{column_name} [get]
func (sc *StockController) GetUniqueByGroupSelectColumn(c *gin.Context) {
	// Parse cluster from path parameter
	clusterStr := c.Param("cluster")
	cluster, err := strconv.Atoi(clusterStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid cluster parameter",
			"details": "Cluster must be an integer",
		})
		return
	}

	// Parse column name from path parameter
	columnName := c.Param("column_name")
	if columnName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid column name parameter",
			"details": "Column name is required",
		})
		return
	}

	// Call service
	values, err := sc.stockService.GetUniqueByGroupSelectColumn(cluster, columnName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to get unique values",
			"details": err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"cluster":     cluster,
		"column_name": columnName,
		"values":      values,
		"count":       len(values),
	})
}

// EmptyAllTables handles DELETE /stocks/tables
// @Summary Empty all tables
// @Description Deletes all records from all tables (rating_sentiments, numerical_indicators, stock_data_points)
// @Tags stocks
// @Produce json
// @Success 200 {object} map[string]interface{} "Tables emptied successfully"
// @Failure 500 {object} map[string]interface{} "Failed to empty tables"
// @Router /api/v1/stocks/tables [delete]
func (sc *StockController) EmptyAllTables(c *gin.Context) {
	if err := sc.stockService.EmptyAllTables(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to empty tables",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All tables emptied successfully",
	})
}
