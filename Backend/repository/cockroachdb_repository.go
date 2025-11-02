package repository

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"dataextractor/config"
	"dataextractor/models"
	"dataextractor/utils"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// NumericalWeightEntry represents a weight for a numerical indicator
type NumericalWeightEntry struct {
	IndicatorName string
	Weight        float64
}

// RatingWeightEntry represents a weight for a rating sentiment
type RatingWeightEntry struct {
	IndicatorName string
	Weight        float64
}

// CockroachDBRepository implements DataRepositoryInterface for CockroachDB using GORM
type CockroachDBRepository struct {
	db *gorm.DB
}

// NewCockroachDBRepository creates a new CockroachDBRepository instance
func NewCockroachDBRepository(db *gorm.DB) *CockroachDBRepository {
	return &CockroachDBRepository{db: db}
}

// Connect establishes CockroachDB connection and runs migrations
func (r *CockroachDBRepository) Connect() error {
	// Load configuration from environment variables
	cfg := config.LoadConfig()

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Build CockroachDB connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s/client.root.crt sslkey=%s/client.root.key sslrootcert=%s/ca.crt",
		cfg.CockroachDB.Host, cfg.CockroachDB.Port, cfg.CockroachDB.User, cfg.CockroachDB.Password,
		cfg.CockroachDB.DBName, cfg.CockroachDB.SSLMode, cfg.CockroachDB.CertsDir, cfg.CockroachDB.CertsDir, cfg.CockroachDB.CertsDir)

	log.Printf("Connecting to CockroachDB: %s:%s/%s", cfg.CockroachDB.Host, cfg.CockroachDB.Port, cfg.CockroachDB.DBName)

	// Connect to CockroachDB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "stock_data.",
		},
	})
	utils.ErrorPanic(err, "failed to connect to CockroachDB")

	// Run database migrations
	utils.ErrorPanic(db.AutoMigrate(&models.StockDataPoint{}, &models.RatingSentiment{}, &models.NumericalIndicator{}), "failed to run migrations")

	// Create CockroachDB-specific indexes on schema-qualified table
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sdp_ticker ON stock_data.stock_data_points (ticker)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sdp_date ON stock_data.stock_data_points (date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sdp_company ON stock_data.stock_data_points (company)")

	log.Println("CockroachDB setup completed successfully")

	// Set the database connection
	r.db = db
	return nil
}

// ReadById retrieves a data point by its ID
func (r *CockroachDBRepository) ReadById(id uint) (*models.StockDataPoint, error) {
	var stock models.StockDataPoint
	if err := r.db.Preload("RatingSentiments").Preload("NumericalIndicators").First(&stock, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("stock with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get stock by ID %d: %w", id, err)
	}
	return &stock, nil
}

// GetAll retrieves all stock records
func (r *CockroachDBRepository) GetAll() ([]models.StockDataPoint, error) {
	var stocks []models.StockDataPoint
	if err := r.db.Preload("RatingSentiments").Preload("NumericalIndicators").Find(&stocks).Error; err != nil {
		return nil, fmt.Errorf("failed to get all stocks: %w", err)
	}
	return stocks, nil
}

// Create creates a new data point
func (r *CockroachDBRepository) Create(entity *models.StockDataPoint) (*models.StockDataPoint, error) {
	utils.ErrorPanic(r.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity).Error, "failed to create data point")
	return entity, nil
}

// Update updates an existing data point
func (r *CockroachDBRepository) Update(entity *models.StockDataPoint) (*models.StockDataPoint, error) {
	utils.ErrorPanic(r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(entity).Error, "failed to update data point")
	return entity, nil
}

// Delete deletes a data point
func (r *CockroachDBRepository) Delete(entity *models.StockDataPoint) error {
	utils.ErrorPanic(r.db.Delete(entity).Error, "failed to delete data point")
	return nil
}

// UpdateOrCreate attempts to create; on unique-constraint conflict updates the existing row
func (r *CockroachDBRepository) UpdateOrCreate(entity *models.StockDataPoint) (*models.StockDataPoint, error) {
	// Try create first
	if err := r.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(entity).Error; err != nil {
		msg := err.Error()
		lower := strings.ToLower(msg)
		if strings.Contains(lower, "duplicate key") || strings.Contains(msg, "SQLSTATE 23505") {
			// Fetch existing by unique key (ticker) and update
			var existing models.StockDataPoint
			if e := r.db.Where("ticker = ?", entity.Ticker).First(&existing).Error; e != nil {
				return nil, fmt.Errorf("failed to fetch existing for upsert: %w", e)
			}
			entity.ID = existing.ID
			if e := r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(entity).Error; e != nil {
				return nil, fmt.Errorf("failed to update existing record: %w", e)
			}
			return entity, nil
		}
		return nil, err
	}
	return entity, nil
}

// GetTotalCount returns the total number of records in the database
func (r *CockroachDBRepository) GetTotalCount() (int64, error) {
	var count int64
	if err := r.db.Model(&models.StockDataPoint{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}
	return count, nil
}

// GetUniqueTickers returns a list of unique tickers in the database
func (r *CockroachDBRepository) GetUniqueTickers() ([]string, error) {
	var tickers []string
	if err := r.db.Model(&models.StockDataPoint{}).Distinct("ticker").Pluck("ticker", &tickers).Error; err != nil {
		return nil, fmt.Errorf("failed to get unique tickers: %w", err)
	}
	return tickers, nil
}

// GetUniqueCompanies returns a list of unique companies in the database
func (r *CockroachDBRepository) GetUniqueCompanies() ([]string, error) {
	var companies []string
	if err := r.db.Model(&models.StockDataPoint{}).Distinct("company").Pluck("company", &companies).Error; err != nil {
		return nil, fmt.Errorf("failed to get unique companies: %w", err)
	}
	return companies, nil
}

// GetDataByTicker returns the data point for a specific ticker (unique)
func (r *CockroachDBRepository) GetDataByTicker(ticker string) (*models.StockDataPoint, error) {
	var stock models.StockDataPoint
	if err := r.db.Preload("RatingSentiments").Preload("NumericalIndicators").Where("ticker = ?", ticker).First(&stock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("stock with ticker %s not found", ticker)
		}
		return nil, fmt.Errorf("failed to get data by ticker %s: %w", ticker, err)
	}
	return &stock, nil
}

// GetDataByCompany returns all data points for a specific company
func (r *CockroachDBRepository) GetDataByCompany(company string) ([]models.StockDataPoint, error) {
	var stocks []models.StockDataPoint
	if err := r.db.Preload("RatingSentiments").Preload("NumericalIndicators").Where("company = ?", company).Find(&stocks).Error; err != nil {
		return nil, fmt.Errorf("failed to get data by company %s: %w", company, err)
	}
	return stocks, nil
}

// GetStocksByCompany is an alias to GetDataByCompany matching service naming
func (r *CockroachDBRepository) GetStocksByCompany(company string) ([]models.StockDataPoint, error) {
	return r.GetDataByCompany(company)
}

// GetLatestData returns the most recent data points (limit specifies how many)
func (r *CockroachDBRepository) GetLatestData(limit int) ([]models.StockDataPoint, error) {
	var stocks []models.StockDataPoint
	if err := r.db.Preload("RatingSentiments").Preload("NumericalIndicators").Order("date DESC").Limit(limit).Find(&stocks).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest data: %w", err)
	}
	return stocks, nil
}

// GetDataByTimeRange returns data points within a specific time range
func (r *CockroachDBRepository) GetDataByTimeRange(startTime, endTime string) ([]models.StockDataPoint, error) {
	var stocks []models.StockDataPoint
	if err := r.db.Preload("RatingSentiments").Preload("NumericalIndicators").Where("date >= ? AND date <= ?", startTime, endTime).Find(&stocks).Error; err != nil {
		return nil, fmt.Errorf("failed to get data by time range: %w", err)
	}
	return stocks, nil
}

// GetTickerStats returns statistics for a specific ticker
func (r *CockroachDBRepository) GetTickerStats(ticker string) (map[string]interface{}, error) {
	var count int64
	var earliestTime, latestTime time.Time

	// Get count
	if err := r.db.Model(&models.StockDataPoint{}).Where("ticker = ?", ticker).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to get ticker count: %w", err)
	}

	// Get time statistics
	if err := r.db.Model(&models.StockDataPoint{}).Where("ticker = ?", ticker).Select("MIN(date), MAX(date)").Row().Scan(&earliestTime, &latestTime); err != nil {
		return nil, fmt.Errorf("failed to get ticker time stats: %w", err)
	}

	return map[string]interface{}{
		"ticker":        ticker,
		"count":         count,
		"earliest_time": earliestTime,
		"latest_time":   latestTime,
	}, nil
}

// GetTopTickersByCount returns the top N tickers by record count
func (r *CockroachDBRepository) GetTopTickersByCount(limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	if err := r.db.Model(&models.StockDataPoint{}).
		Select("ticker, COUNT(*) as count").
		Group("ticker").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get top tickers by count: %w", err)
	}

	return results, nil
}

// GetDatabaseStats returns overall database statistics
func (r *CockroachDBRepository) GetDatabaseStats() (map[string]interface{}, error) {
	var totalCount int64
	var uniqueTickers, uniqueCompanies int64

	// Get total count
	if err := r.db.Model(&models.StockDataPoint{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get unique tickers count
	if err := r.db.Model(&models.StockDataPoint{}).Distinct("ticker").Count(&uniqueTickers).Error; err != nil {
		return nil, fmt.Errorf("failed to get unique tickers count: %w", err)
	}

	// Get unique companies count
	if err := r.db.Model(&models.StockDataPoint{}).Distinct("company").Count(&uniqueCompanies).Error; err != nil {
		return nil, fmt.Errorf("failed to get unique companies count: %w", err)
	}

	return map[string]interface{}{
		"total_records":    totalCount,
		"unique_tickers":   uniqueTickers,
		"unique_companies": uniqueCompanies,
	}, nil
}

// GetUniqueClusters returns a list of unique cluster IDs
func (r *CockroachDBRepository) GetUniqueClusters() ([]int, error) {
	var clusters []int
	if err := r.db.Model(&models.StockDataPoint{}).Distinct("cluster").Pluck("cluster", &clusters).Error; err != nil {
		return nil, fmt.Errorf("failed to get unique clusters: %w", err)
	}
	sort.Ints(clusters)
	return clusters, nil
}

// GetStocksByCluster returns all data points for a specific cluster
func (r *CockroachDBRepository) GetStocksByCluster(cluster int) ([]models.StockDataPoint, error) {
	var stocks []models.StockDataPoint
	if err := r.db.Preload("RatingSentiments").Preload("NumericalIndicators").Where("cluster = ?", cluster).Find(&stocks).Error; err != nil {
		return nil, fmt.Errorf("failed to get data by cluster %d: %w", cluster, err)
	}
	return stocks, nil
}

// GetUniqueActions returns a list of unique actions
func (r *CockroachDBRepository) GetUniqueActions() ([]string, error) {
	var actions []string
	if err := r.db.Model(&models.StockDataPoint{}).Distinct("action").Pluck("action", &actions).Error; err != nil {
		return nil, fmt.Errorf("failed to get unique actions: %w", err)
	}
	sort.Strings(actions)
	return actions, nil
}

// GetStocksByAction returns all data points for a specific action
func (r *CockroachDBRepository) GetStocksByAction(action string) ([]models.StockDataPoint, error) {
	var stocks []models.StockDataPoint
	if err := r.db.Preload("RatingSentiments").Preload("NumericalIndicators").Where("action = ?", action).Find(&stocks).Error; err != nil {
		return nil, fmt.Errorf("failed to get data by action %s: %w", action, err)
	}
	return stocks, nil
}

// GetStocksByClusterAndGroup filters by cluster and optionally by groupingColumn using GORM
// Returns stocks, total count, and error
func (r *CockroachDBRepository) GetStocksByClusterAndGroup(cluster int, groupingColumn string, groupingValue string, sortByColumn string, order string, page, perPage int, numericalWeights []NumericalWeightEntry, ratingWeights []RatingWeightEntry) ([]models.StockDataPoint, int64, error) {
	// Whitelist of allowed column names for sorting/filtering (full list)
	allowedColumns := []string{
		"ticker", "action", "date", "company", "cluster",
		"target_to", "target_from", "target_delta", "last_close", "rating_to", "rating_from", "final_score", "weighted_score",
	}

	// Whitelist of allowed grouping columns (excluding company and date due to too many distinct values)
	allowedGroupingColumns := []string{
		"action", "rating_to", "rating_from",
	}

	// Validate sortByColumn early
	if sortByColumn != "" {
		if !validateColumnName(sortByColumn, allowedColumns) {
			return nil, 0, fmt.Errorf("invalid sort column: %s", sortByColumn)
		}
	}

	// Check if both weight arrays are provided (required for weighted_score sorting)
	hasBothWeights := len(numericalWeights) > 0 && len(ratingWeights) > 0
	hasAnyWeights := len(numericalWeights) > 0 || len(ratingWeights) > 0

	// Determine if we should sort by weighted_score (only if both arrays are provided)
	sortByWeightedScore := sortByColumn == "weighted_score" && hasBothWeights

	// Build base query for filtering and counting (before weighted scores join)
	baseQuery := r.db.Model(&models.StockDataPoint{}).
		Where("cluster = ?", cluster)

	// Filter by groupingColumn if not "None" - validate against grouping-specific whitelist
	if groupingColumn != "None" && groupingValue != "" {
		if !validateColumnName(groupingColumn, allowedGroupingColumns) {
			return nil, 0, fmt.Errorf("invalid grouping column: %s. Allowed grouping columns: %v", groupingColumn, allowedGroupingColumns)
		}
		baseQuery = baseQuery.Where(fmt.Sprintf("%s = ?", groupingColumn), groupingValue)
	}

	// Calculate total count efficiently before weighted score joins
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count stocks: %w", err)
	}

	// Build query for fetching stocks (same filters as count query)
	query := baseQuery

	var sortOrder string

	// If not sorting by weighted_score (or if weighted_score sort is not applicable), sort before the join
	// Note: If sortByColumn is "weighted_score" but both weights aren't provided, skip sorting entirely
	if sortByColumn != "" && !sortByWeightedScore {
		// Only sort if it's not a weighted_score request without both weights
		if !(sortByColumn == "weighted_score" && !hasBothWeights) {
			if strings.ToLower(order) == "desc" {
				sortOrder = "DESC"
			} else {
				sortOrder = "ASC"
			}
			query = query.Order(fmt.Sprintf("%s %s", sortByColumn, sortOrder))
		}
	}

	// Prepare sort order for weighted_score (always DESC when sorting by weighted_score)
	if sortByWeightedScore {
		sortOrder = "DESC"
	}

	// Calculate combined weighted scores: join indicator and rating subqueries, sum their scores
	if hasAnyWeights {
		// Get table names
		niTableName := (&models.NumericalIndicator{}).TableName()
		rsTableName := (&models.RatingSentiment{}).TableName()

		// Convert weight slices to generic format using helper methods
		indicatorWeights := convertNumericalWeights(numericalWeights)
		ratingWeightEntries := convertRatingWeights(ratingWeights)

		// Build subqueries using helper method
		indicatorSubquery := buildWeightedScoreSubquery(niTableName, "norm_value", "new_indicator_score", "ni_sub", indicatorWeights)
		ratingSubquery := buildWeightedScoreSubquery(rsTableName, "norm_rating_score", "new_rating_score", "rs_sub", ratingWeightEntries)

		// Combine indicator and rating subqueries into a single combined subquery
		combinedSubquery := combineWeightedScoreSubqueries(indicatorSubquery, ratingSubquery)

		// Simple INNER JOIN with stock_data_points
		// Select weighted_score with explicit alias to ensure GORM maps it to WeightedScore field
		// GORM maps snake_case column names (weighted_score) to PascalCase fields (WeightedScore)
		query = query.
			Select("stock_data_points.*, combined_scores.weighted_score AS weighted_score").
			Joins(fmt.Sprintf("INNER JOIN %s combined_scores ON combined_scores.stock_data_point_id = stock_data_points.id", combinedSubquery))

		// Sort by weighted_score after the join
		if sortByWeightedScore {
			query = query.Order(fmt.Sprintf("combined_scores.weighted_score %s", sortOrder))
		}
	}

	// Apply pagination
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	// Preload relations: RatingSentiments and NumericalIndicators
	query = query.Preload("RatingSentiments").Preload("NumericalIndicators")

	// Define struct that embeds StockDataPoint and includes weighted_score
	type StockDataPointWithWeightedScore struct {
		models.StockDataPoint
		WeightedScore float64 `gorm:"column:weighted_score"`
	}

	var stocksWithScore []StockDataPointWithWeightedScore

	// Use Find() with Preload - GORM will automatically populate weighted_score from the JOIN
	if len(numericalWeights) > 0 || len(ratingWeights) > 0 {
		// Find() with Preload handles both the weighted_score mapping and relation preloading
		if err := query.Find(&stocksWithScore).Error; err != nil {
			return nil, 0, fmt.Errorf("failed to get stocks with weighted score: %w", err)
		}

		// Convert back to StockDataPoint and set WeightedScore
		stocks := make([]models.StockDataPoint, len(stocksWithScore))
		for i, sws := range stocksWithScore {
			stocks[i] = sws.StockDataPoint
			// Map the weighted_score column value to WeightedScore pointer field
			stocks[i].WeightedScore = &sws.WeightedScore
		}

		return stocks, totalCount, nil
	}

	// No weighted scores, use normal Find() which handles Preload automatically
	var stocks []models.StockDataPoint
	if err := query.Find(&stocks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get stocks by cluster and group: %w", err)
	}

	return stocks, totalCount, nil
}

// GetUniqueByGroupSelectColumn returns unique values for a specified column filtered by cluster
// columnName must be one of: 'action', 'rating_to', 'rating_from'
// Note: 'company' and 'date' are excluded due to having too many distinct values
func (r *CockroachDBRepository) GetUniqueByGroupSelectColumn(cluster int, columnName string) ([]string, error) {
	// Whitelist of allowed column names (excluding company and date due to too many distinct values)
	allowedColumns := []string{"action", "rating_to", "rating_from"}

	// Validate column name
	if !validateColumnName(columnName, allowedColumns) {
		return nil, fmt.Errorf("invalid column name: %s. Allowed values: %v", columnName, allowedColumns)
	}

	// Filter by cluster first, then get distinct values for the specified column
	var values []string
	if err := r.db.Model(&models.StockDataPoint{}).
		Where("cluster = ?", cluster).
		Distinct(columnName).
		Pluck(columnName, &values).Error; err != nil {
		return nil, fmt.Errorf("failed to get unique values for column %s in cluster %d: %w", columnName, cluster, err)
	}

	// Sort the results
	sort.Strings(values)

	return values, nil
}

// EmptyAllTables deletes all records from all tables in the correct order
// Deletes child tables first (rating_sentiments, numerical_indicators), then parent table (stock_data_points)
// If tables don't exist, GORM will handle the error gracefully
func (r *CockroachDBRepository) EmptyAllTables() error {
	log.Println("Emptying all tables...")

	// Delete from child tables first (due to foreign key constraints)
	// Using GORM's Model and Delete - will return error if table doesn't exist, which is acceptable
	if err := r.db.Model(&models.RatingSentiment{}).Where("1 = 1").Delete(&models.RatingSentiment{}).Error; err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			log.Println("rating_sentiments table does not exist, skipping")
		} else {
			return fmt.Errorf("failed to empty rating_sentiments table: %w", err)
		}
	} else {
		log.Println("Emptied rating_sentiments table")
	}

	if err := r.db.Model(&models.NumericalIndicator{}).Where("1 = 1").Delete(&models.NumericalIndicator{}).Error; err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			log.Println("numerical_indicators table does not exist, skipping")
		} else {
			return fmt.Errorf("failed to empty numerical_indicators table: %w", err)
		}
	} else {
		log.Println("Emptied numerical_indicators table")
	}

	// Delete from parent table last
	if err := r.db.Model(&models.StockDataPoint{}).Where("1 = 1").Delete(&models.StockDataPoint{}).Error; err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			log.Println("stock_data_points table does not exist, skipping")
		} else {
			return fmt.Errorf("failed to empty stock_data_points table: %w", err)
		}
	} else {
		log.Println("Emptied stock_data_points table")
	}

	log.Println("All tables emptied successfully")
	return nil
}
