package repository

import (
	"log"
	"testing"
	"time"

	"dataextractor/models"
)

// TestGetStocksByClusterAndGroup tests the GetStocksByClusterAndGroup method
// This is a temporary test file for performance and functionality testing
func TestGetStocksByClusterAndGroup(t *testing.T) {
	// Initialize repository
	repo := NewCockroachDBRepository(nil)
	if err := repo.Connect(); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Test cases
	testCases := []struct {
		name             string
		cluster          int
		groupingColumn   string
		groupingValue    string
		sortByColumn     string
		order            string
		page             int
		perPage          int
		numericalWeights []NumericalWeightEntry
		ratingWeights    []RatingWeightEntry
	}{
		{
			name:           "Basic test - cluster only",
			cluster:        0,
			groupingColumn: "None",
			groupingValue:  "",
			sortByColumn:   "date",
			order:          "desc",
			page:           1,
			perPage:        20,
		},
		{
			name:           "Test with grouping by company",
			cluster:        0,
			groupingColumn: "company",
			groupingValue:  "Apple Inc.",
			sortByColumn:   "final_score",
			order:          "desc",
			page:           1,
			perPage:        10,
		},
		{
			name:           "Test with numerical weights",
			cluster:        0,
			groupingColumn: "None",
			groupingValue:  "",
			sortByColumn:   "date",
			order:          "asc",
			page:           1,
			perPage:        20,
			numericalWeights: []NumericalWeightEntry{
				{IndicatorName: "atr", Weight: 0.5},
				{IndicatorName: "obv", Weight: 0.3},
			},
		},
		{
			name:           "Test with rating weights",
			cluster:        0,
			groupingColumn: "None",
			groupingValue:  "",
			sortByColumn:   "date",
			order:          "desc",
			page:           1,
			perPage:        20,
			ratingWeights: []RatingWeightEntry{
				{IndicatorName: "action", Weight: 0.7},
			},
		},
		{
			name:           "Test with both weights",
			cluster:        0,
			groupingColumn: "action",
			groupingValue:  "Buy",
			sortByColumn:   "weighted_score",
			order:          "desc",
			page:           1,
			perPage:        15,
			numericalWeights: []NumericalWeightEntry{
				{IndicatorName: "atr", Weight: 0.4},
				{IndicatorName: "obv", Weight: 0.2},
			},
			ratingWeights: []RatingWeightEntry{
				{IndicatorName: "action", Weight: 0.4},
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			startTime := time.Now()

			// Execute the method
					stocks, _, err := repo.GetStocksByClusterAndGroup(
				tc.cluster,
				tc.groupingColumn,
				tc.groupingValue,
				tc.sortByColumn,
				tc.order,
				tc.page,
				tc.perPage,
				tc.numericalWeights,
				tc.ratingWeights,
			)

			latency := time.Since(startTime)

			// Check for errors
			if err != nil {
				t.Errorf("GetStocksByClusterAndGroup failed: %v", err)
				return
			}

			// Verify results
			if stocks == nil {
				t.Error("Expected non-nil stocks slice, got nil")
				return
			}

			// Verify pagination
			if len(stocks) > tc.perPage {
				t.Errorf("Expected at most %d results, got %d", tc.perPage, len(stocks))
			}

			// Log results
			log.Printf("\n=== Test: %s ===", tc.name)
			log.Printf("Latency: %v", latency)
			log.Printf("Results returned: %d", len(stocks))
			log.Printf("Parameters: cluster=%d, groupingColumn=%s, groupingValue=%s, sortBy=%s, order=%s, page=%d, perPage=%d",
				tc.cluster, tc.groupingColumn, tc.groupingValue, tc.sortByColumn, tc.order, tc.page, tc.perPage)
			if len(tc.numericalWeights) > 0 {
				log.Printf("Numerical weights: %v", tc.numericalWeights)
			}
			if len(tc.ratingWeights) > 0 {
				log.Printf("Rating weights: %v", tc.ratingWeights)
			}

			// Show sample results
			if len(stocks) > 0 {
				log.Printf("Sample result:")
				sample := stocks[0]
				log.Printf("  Ticker: %s, Company: %s, Date: %s, FinalScore: %.4f",
					sample.Ticker, sample.Company, sample.Date.Format("2006-01-02"), sample.FinalScore)
				if sample.WeightedScore != nil {
					log.Printf("  WeightedScore: %.4f", *sample.WeightedScore)
				}
				log.Printf("  RatingSentiments count: %d", len(sample.RatingSentiments))
				log.Printf("  NumericalIndicators count: %d", len(sample.NumericalIndicators))
			}

			// Verify sorting (if applicable)
			if tc.sortByColumn != "" && len(stocks) > 1 {
				verifySorting(t, stocks, tc.sortByColumn, tc.order)
			}

			log.Printf("=== End Test: %s ===\n", tc.name)
		})
	}
}

// verifySorting checks if the results are sorted correctly
func verifySorting(t *testing.T, stocks []models.StockDataPoint, sortByColumn string, order string) {
	isDesc := order == "desc" || order == "DESC"

	for i := 0; i < len(stocks)-1; i++ {
		current := stocks[i]
		next := stocks[i+1]
		var isCorrectOrder bool

		switch sortByColumn {
		case "date":
			if isDesc {
				isCorrectOrder = current.Date.After(next.Date) || current.Date.Equal(next.Date)
			} else {
				isCorrectOrder = current.Date.Before(next.Date) || current.Date.Equal(next.Date)
			}
		case "final_score":
			if isDesc {
				isCorrectOrder = current.FinalScore >= next.FinalScore
			} else {
				isCorrectOrder = current.FinalScore <= next.FinalScore
			}
		case "ticker":
			if isDesc {
				isCorrectOrder = current.Ticker >= next.Ticker
			} else {
				isCorrectOrder = current.Ticker <= next.Ticker
			}
		case "company":
			if isDesc {
				isCorrectOrder = current.Company >= next.Company
			} else {
				isCorrectOrder = current.Company <= next.Company
			}
		case "weighted_score":
			if current.WeightedScore != nil && next.WeightedScore != nil {
				if isDesc {
					isCorrectOrder = *current.WeightedScore >= *next.WeightedScore
				} else {
					isCorrectOrder = *current.WeightedScore <= *next.WeightedScore
				}
			}
		default:
			// Skip verification for unsupported columns
			return
		}

		if !isCorrectOrder {
			t.Errorf("Sorting verification failed at index %d: %s order not maintained for column %s",
				i, order, sortByColumn)
		}
	}
}

// BenchmarkGetStocksByClusterAndGroup benchmarks the method performance
func BenchmarkGetStocksByClusterAndGroup(b *testing.B) {
	repo := NewCockroachDBRepository(nil)
	if err := repo.Connect(); err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}

	numericalWeights := []NumericalWeightEntry{
		{IndicatorName: "atr", Weight: 0.4},
		{IndicatorName: "obv", Weight: 0.2},
	}
	ratingWeights := []RatingWeightEntry{
		{IndicatorName: "action", Weight: 0.4},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.GetStocksByClusterAndGroup(
			0,      // cluster
			"None", // groupingColumn
			"",     // groupingValue
			"date", // sortByColumn
			"desc", // order
			1,      // page
			20,     // perPage
			numericalWeights,
			ratingWeights,
		)
		if err != nil {
			b.Errorf("Benchmark failed: %v", err)
		}
	}
}
