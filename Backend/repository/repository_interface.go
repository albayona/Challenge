package repository

import "dataextractor/models"

// DataRepositoryInterface defines the contract for data repository operations
type DataRepositoryInterface interface {
	// Connection management
	Connect() error

	// Basic CRUD operations
	ReadById(id uint) (*models.StockDataPoint, error)
	GetAll() ([]models.StockDataPoint, error)
	Create(entity *models.StockDataPoint) (*models.StockDataPoint, error)
	Update(entity *models.StockDataPoint) (*models.StockDataPoint, error)
	Delete(entity *models.StockDataPoint) error
	UpdateOrCreate(entity *models.StockDataPoint) (*models.StockDataPoint, error)

	// Database exploration methods
	GetTotalCount() (int64, error)
	GetUniqueTickers() ([]string, error)
	GetUniqueCompanies() ([]string, error)
	GetStocksByCompany(company string) ([]models.StockDataPoint, error)
	GetDataByTicker(ticker string) (*models.StockDataPoint, error)
	GetLatestData(limit int) ([]models.StockDataPoint, error)
	GetDataByTimeRange(startTime, endTime string) ([]models.StockDataPoint, error)
	GetTickerStats(ticker string) (map[string]interface{}, error)
	GetTopTickersByCount(limit int) ([]map[string]interface{}, error)
	GetDatabaseStats() (map[string]interface{}, error)

	// Cluster queries
	GetUniqueClusters() ([]int, error)
	GetStocksByCluster(cluster int) ([]models.StockDataPoint, error)
	GetStocksByClusterAndGroup(cluster int, groupingColumn string, groupingValue string, sortByColumn string, order string,
		page, perPage int, numericalWeights []NumericalWeightEntry, ratingWeights []RatingWeightEntry) ([]models.StockDataPoint, int64, error)

	// Action queries
	GetUniqueActions() ([]string, error)
	GetStocksByAction(action string) ([]models.StockDataPoint, error)

	// Group select column queries
	GetUniqueByGroupSelectColumn(cluster int, columnName string) ([]string, error)

	// Table management
	EmptyAllTables() error
}
