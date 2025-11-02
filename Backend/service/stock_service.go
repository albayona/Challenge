package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"dataextractor/config"
	"dataextractor/data_extractor"
	"dataextractor/db_populate"
	"dataextractor/models"
	"dataextractor/repository"
	"dataextractor/utils"
	"dataextractor/validators"
)

// StockService handles business logic for stock operations
type StockService struct {
	repository repository.DataRepositoryInterface
	validator  *validators.StockValidator
}

// NewStockService creates a new StockService instance
func NewStockService(repo repository.DataRepositoryInterface) *StockService {
	return &StockService{
		repository: repo,
		validator:  validators.NewStockValidator(),
	}
}

// Create creates a new stock record with validation
func (s *StockService) Create(request *validators.StockCreateRequest) (*models.StockDataPoint, error) {
	// Validate the request using the service validator
	utils.ErrorPanic(s.validator.ValidateRequest(request), "validation failed")

	// Convert request to Stock model
	stock := request.ToStock()

	// Create the stock record
	createdStock, err := s.repository.Create(stock)
	utils.ErrorPanic(err, "failed to create stock")

	log.Printf("Successfully created stock record for ticker: %s", createdStock.Ticker)
	return createdStock, nil
}

// GetByID retrieves a stock record by its ID
func (s *StockService) GetByID(id uint) (*models.StockDataPoint, error) {
	// Validate the ID using the service validator
	utils.ErrorPanic(s.validator.ValidateID(id), "invalid ID")

	stock, err := s.repository.ReadById(id)
	utils.ErrorPanic(err, fmt.Sprintf("failed to get stock by ID %d", id))

	return stock, nil
}

// GetAll retrieves all stock records
func (s *StockService) GetAll() ([]models.StockDataPoint, error) {
	stocks, err := s.repository.GetAll()
	utils.ErrorPanic(err, "failed to get all stocks")

	return stocks, nil
}

// Update updates an existing stock record with validation
func (s *StockService) Update(request *validators.StockUpdateRequest) (*models.StockDataPoint, error) {
	// Validate the request using the service validator
	utils.ErrorPanic(s.validator.ValidateRequest(request), "validation failed")

	// Convert request to Stock model
	stock := request.ToStock()

	// Update the stock record
	updatedStock, err := s.repository.Update(stock)
	utils.ErrorPanic(err, "failed to update stock")

	log.Printf("Successfully updated stock record for ticker: %s", updatedStock.Ticker)
	return updatedStock, nil
}

// Delete deletes a stock record by ID
func (s *StockService) Delete(id uint) error {
	// Validate the ID using the service validator
	utils.ErrorPanic(s.validator.ValidateID(id), "invalid ID")

	// First, get the stock to ensure it exists
	stock, err := s.repository.ReadById(id)
	utils.ErrorPanic(err, fmt.Sprintf("stock with ID %d not found", id))

	// Delete the stock record
	utils.ErrorPanic(s.repository.Delete(stock), "failed to delete stock")

	log.Printf("Successfully deleted stock record for ticker: %s", stock.Ticker)
	return nil
}

// GetByTicker retrieves the stock record for a specific ticker (unique)
func (s *StockService) GetByTicker(ticker string) (*models.StockDataPoint, error) {
	// Validate the ticker using the service validator
	if err := s.validator.ValidateTicker(ticker); err != nil {
		return nil, fmt.Errorf("invalid ticker: %w", err)
	}

	stock, err := s.repository.GetDataByTicker(ticker)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock by ticker %s: %w", ticker, err)
	}

	return stock, nil
}

// GetByCompany retrieves all stock records for a specific company
func (s *StockService) GetByCompany(company string) ([]models.StockDataPoint, error) {
	// Validate the company using the service validator
	if err := s.validator.ValidateCompany(company); err != nil {
		return nil, fmt.Errorf("invalid company: %w", err)
	}

	stocks, err := s.repository.GetStocksByCompany(company)
	if err != nil {
		return nil, fmt.Errorf("failed to get stocks by company %s: %w", company, err)
	}

	return stocks, nil
}

// GetStocksByCompany is a convenience alias matching new naming
func (s *StockService) GetStocksByCompany(company string) ([]models.StockDataPoint, error) {
	return s.GetByCompany(company)
}

// GetUniqueClusters returns all unique clusters
func (s *StockService) GetUniqueClusters() ([]int, error) {
	clusters, err := s.repository.GetUniqueClusters()
	utils.ErrorPanic(err, "failed to get unique clusters")
	return clusters, nil
}

// GetStocksByCluster returns all stocks for a specific cluster
func (s *StockService) GetStocksByCluster(cluster int) ([]models.StockDataPoint, error) {
	if cluster < 0 {
		return nil, fmt.Errorf("invalid cluster: must be >= 0")
	}
	stocks, err := s.repository.GetStocksByCluster(cluster)
	utils.ErrorPanic(err, fmt.Sprintf("failed to get stocks by cluster %d", cluster))
	return stocks, nil
}

// GetUniqueActions returns all unique actions
func (s *StockService) GetUniqueActions() ([]string, error) {
	actions, err := s.repository.GetUniqueActions()
	utils.ErrorPanic(err, "failed to get unique actions")
	return actions, nil
}

// GetUniqueCompanies returns all unique companies
func (s *StockService) GetUniqueCompanies() ([]string, error) {
	companies, err := s.repository.GetUniqueCompanies()
	utils.ErrorPanic(err, "failed to get unique companies")
	return companies, nil
}

// GetStocksByAction returns all stocks for a specific action
func (s *StockService) GetStocksByAction(action string) ([]models.StockDataPoint, error) {
	if action == "" {
		return nil, fmt.Errorf("invalid action: required")
	}
	stocks, err := s.repository.GetStocksByAction(action)
	utils.ErrorPanic(err, fmt.Sprintf("failed to get stocks by action %s", action))
	return stocks, nil
}

// (moved) ImportFromCSV now lives in package db_populate

// GetStats retrieves statistics for a specific ticker
func (s *StockService) GetStats(ticker string) (map[string]interface{}, error) {
	// Validate the ticker using the service validator
	utils.ErrorPanic(s.validator.ValidateTicker(ticker), "invalid ticker")

	stats, err := s.repository.GetTickerStats(ticker)
	utils.ErrorPanic(err, fmt.Sprintf("failed to get stats for ticker %s", ticker))

	return stats, nil
}

// GetDatabaseStats retrieves overall database statistics
func (s *StockService) GetDatabaseStats() (map[string]interface{}, error) {
	stats, err := s.repository.GetDatabaseStats()
	utils.ErrorPanic(err, "failed to get database stats")

	return stats, nil
}

// StoreDataFromApi handles the complete data extraction process from API
func (s *StockService) StoreDataFromApi(maxPages int) error {
	// Load configuration for API
	cfg := config.LoadConfig()

	// Create data extractor and run it
	extractor := data_extractor.NewDataExtractor(cfg.APIBaseURL, cfg.APIKey, s.repository)

	log.Printf("Starting data extraction with maxPages: %d", maxPages)
	if err := extractor.ExtractAndProcessAllPages(maxPages); err != nil {
		return fmt.Errorf("error during data extraction: %w", err)
	}

	log.Println("Data extraction completed successfully! Data written to CSV file.")
	return nil
}

// ImportFromCSV delegates CSV import to db_populate, persisting with the repository
func (s *StockService) ImportFromCSV(reader io.Reader) (int, error) {
	return db_populate.ImportFromCSV(reader, s.repository)
}

// ImportFromEnrichedCSV opens the default CSV file and imports it
func (s *StockService) ImportFromEnrichedCSV() (int, error) {
	const defaultCSV = "./stock_data_enriched.csv"
	f, err := os.Open(defaultCSV)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file %s: %w", defaultCSV, err)
	}
	defer f.Close()
	return db_populate.ImportFromCSV(f, s.repository)
}

// RankByWeightedScore computes weighted scores for all data points in a cluster and returns them sorted desc
func (s *StockService) RankByWeightedScore(cluster int, weights []WeightEntry) ([]RankedResult, error) {
	if cluster < 0 {
		return nil, fmt.Errorf("invalid cluster: must be >= 0")
	}

	// Fetch data points for the cluster with preloaded associations
	dataPoints, err := s.repository.GetStocksByCluster(cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get stocks by cluster %d: %w", cluster, err)
	}

	// Build weight map (case-insensitive on indicator/sentiment name)
	weightByName := make(map[string]float64, len(weights))
	for _, w := range weights {
		key := strings.TrimSpace(strings.ToLower(w.IndicatorName))
		if key == "" {
			continue
		}
		weightByName[key] = w.Weight
	}

	results := make([]RankedResult, 0, len(dataPoints))
	for _, sdp := range dataPoints {
		var score float64

		// Rating sentiments -> use NormRatingScore
		for _, rs := range sdp.RatingSentiments {
			key := strings.TrimSpace(strings.ToLower(rs.Name))
			if w, ok := weightByName[key]; ok {
				score += w * rs.NormRatingScore
			}
		}

		// Numerical indicators -> use NormValue
		for _, ni := range sdp.NumericalIndicators {
			key := strings.TrimSpace(strings.ToLower(ni.Name))
			if w, ok := weightByName[key]; ok {
				score += w * ni.NormValue
			}
		}

		results = append(results, RankedResult{Stock: sdp, Score: score})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	return results, nil
}

// FilterByClusterGrouped filters by cluster with grouping, pagination, sorting, and optional weighted scoring
func (s *StockService) FilterByClusterGrouped(cluster int, groupingColumn string, groupingValue string, sortByColumn string, order string, page, perPage int, numericalWeights []repository.NumericalWeightEntry, ratingWeights []repository.RatingWeightEntry) (PagedGroupedResults, error) {

	// Get stocks from repository (returns stocks and total count)
	stocks, totalCount, err := s.repository.GetStocksByClusterAndGroup(cluster, groupingColumn, groupingValue, sortByColumn, order, page, perPage, numericalWeights, ratingWeights)
	if err != nil {
		return PagedGroupedResults{}, fmt.Errorf("failed to filter stocks: %w", err)
	}

	return PagedGroupedResults{
		Items:      stocks,
		TotalCount: totalCount,
		Page:       page,
		PerPage:    perPage,
	}, nil
}

// GetUniqueByGroupSelectColumn returns unique values for a specified column filtered by cluster
func (s *StockService) GetUniqueByGroupSelectColumn(cluster int, columnName string) ([]string, error) {
	if columnName == "" {
		return nil, fmt.Errorf("column name is required")
	}

	values, err := s.repository.GetUniqueByGroupSelectColumn(cluster, columnName)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique values for column %s in cluster %d: %w", columnName, cluster, err)
	}

	return values, nil
}

// EmptyAllTables empties all tables by deleting all records
func (s *StockService) EmptyAllTables() error {
	if err := s.repository.EmptyAllTables(); err != nil {
		return fmt.Errorf("failed to empty all tables: %w", err)
	}
	return nil
}
