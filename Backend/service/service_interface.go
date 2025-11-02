package service

import (
	"dataextractor/models"
	"dataextractor/repository"
	"dataextractor/validators"
	"io"
)

// StockServiceInterface defines the contract for stock service operations
type StockServiceInterface interface {
	// CRUD Operations
	Create(request *validators.StockCreateRequest) (*models.StockDataPoint, error)
	GetByID(id uint) (*models.StockDataPoint, error)
	GetAll() ([]models.StockDataPoint, error)
	Update(request *validators.StockUpdateRequest) (*models.StockDataPoint, error)
	Delete(id uint) error

	// Find Operations
	GetByTicker(ticker string) (*models.StockDataPoint, error)
	GetByCompany(company string) ([]models.StockDataPoint, error)
	GetStocksByCompany(company string) ([]models.StockDataPoint, error)
	GetUniqueCompanies() ([]string, error)

	// Statistics Operations
	GetStats(ticker string) (map[string]interface{}, error)
	GetDatabaseStats() (map[string]interface{}, error)

	// Data Extraction Operations
	StoreDataFromApi(maxPages int) error

	// Cluster Operations
	GetUniqueClusters() ([]int, error)
	GetStocksByCluster(cluster int) ([]models.StockDataPoint, error)

	// Action Operations
	GetUniqueActions() ([]string, error)
	GetStocksByAction(action string) ([]models.StockDataPoint, error)

	// CSV Import
	ImportFromCSV(reader io.Reader) (int, error)
	ImportFromEnrichedCSV() (int, error)

	// Scoring Operations
	RankByWeightedScore(cluster int, weights []WeightEntry) ([]RankedResult, error)

	// Grouped, paginated, sortable filter by cluster
	FilterByClusterGrouped(cluster int, groupingColumn string, groupingValue string, sortByColumn string, order string, page, perPage int, numericalWeights []repository.NumericalWeightEntry, ratingWeights []repository.RatingWeightEntry) (PagedGroupedResults, error)

	// Group select column operations
	GetUniqueByGroupSelectColumn(cluster int, columnName string) ([]string, error)

	// Table management operations
	EmptyAllTables() error
}

// WeightEntry represents a weight for a given indicator/sentiment name
type WeightEntry struct {
	IndicatorName string  `json:"indicator_name"`
	Weight        float64 `json:"weight"`
}

// WeightEntry represents a weight for a given indicator/sentiment name
type NumericalWeightEntry struct {
	IndicatorName string  `json:"indicator_name"`
	Weight        float64 `json:"weight"`
}

// WeightEntry represents a weight for a given indicator/sentiment name
type RatingWeightEntry struct {
	IndicatorName string  `json:"indicator_name"`
	Weight        float64 `json:"weight"`
}

// RankedResult represents a data point with its computed weighted score
type RankedResult struct {
	Stock models.StockDataPoint
	Score float64
}

// GroupedStock wraps a stock with its grouping value
type GroupedStock struct {
	Group string                `json:"group"`
	Stock models.StockDataPoint `json:"stock"`
}

// PagedGroupedResults carries page data and total for pagination
type PagedGroupedResults struct {
	Items      []models.StockDataPoint `json:"items"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	PerPage    int                     `json:"per_page"`
}
