package main

import (
	"fmt"
	"log"
	"time"

	"dataextractor/repository"
)

func main() {
	// Initialize repository
	repo := repository.NewCockroachDBRepository(nil)
	if err := repo.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Enable SQL logging to see the queries
	// This will be shown in the repository's log output
	fmt.Println("=== SQL Query Logging Enabled ===\n")

	// Check what actions actually exist in the database
	fmt.Println("=== Checking available actions in database ===")
	actions, err := repo.GetUniqueActions()
	if err != nil {
		log.Printf("Warning: Could not get unique actions: %v", err)
	} else {
		fmt.Printf("Available actions (%d):\n", len(actions))
		for i, action := range actions {
			if i < 20 { // Show first 20
				fmt.Printf("  - '%s'\n", action)
			}
		}
		if len(actions) > 20 {
			fmt.Printf("  ... and %d more\n", len(actions)-20)
		}
		fmt.Println()
	}

	// Check what actions exist in cluster 0 specifically
	fmt.Println("=== Checking actions in cluster 0 ===")
	cluster0Stocks, err := repo.GetStocksByCluster(0)
	if err != nil {
		log.Printf("Warning: Could not get stocks for cluster 0: %v", err)
	} else {
		actionMap := make(map[string]int)
		for _, stock := range cluster0Stocks {
			if stock.Action != "" {
				actionMap[stock.Action]++
			}
		}
		fmt.Printf("Actions in cluster 0 (%d stocks):\n", len(cluster0Stocks))
		for action, count := range actionMap {
			fmt.Printf("  - '%s': %d occurrences\n", action, count)
		}
		fmt.Println()
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
		numericalWeights []repository.NumericalWeightEntry
		ratingWeights    []repository.RatingWeightEntry
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
			numericalWeights: []repository.NumericalWeightEntry{
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
			ratingWeights: []repository.RatingWeightEntry{
				{IndicatorName: "action", Weight: 0.7},
			},
		},
		{
			name:           "Test with both weights",
			cluster:        0,
			groupingColumn: "action",
			groupingValue:  "target raised by", // Use an actual action that exists in the database
			sortByColumn:   "date",
			order:          "desc",
			page:           1,
			perPage:        15,
			numericalWeights: []repository.NumericalWeightEntry{
				{IndicatorName: "atr", Weight: 0.4},
				{IndicatorName: "obv", Weight: 0.2},
			},
			ratingWeights: []repository.RatingWeightEntry{
				{IndicatorName: "action", Weight: 0.4},
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		fmt.Printf("\n=== Test: %s ===\n", tc.name)
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
			log.Printf("ERROR: GetStocksByClusterAndGroup failed: %v\n", err)
			continue
		}

		// Verify results
		if stocks == nil {
			log.Printf("ERROR: Expected non-nil stocks slice, got nil\n")
			continue
		}

		// Verify pagination
		if len(stocks) > tc.perPage {
			log.Printf("ERROR: Expected at most %d results, got %d\n", tc.perPage, len(stocks))
			continue
		}

		// Log results
		fmt.Printf("Latency: %v\n", latency)
		fmt.Printf("Results returned: %d\n", len(stocks))
		fmt.Printf("Parameters: cluster=%d, groupingColumn=%s, groupingValue=%s, sortBy=%s, order=%s, page=%d, perPage=%d\n",
			tc.cluster, tc.groupingColumn, tc.groupingValue, tc.sortByColumn, tc.order, tc.page, tc.perPage)
		if len(tc.numericalWeights) > 0 {
			fmt.Printf("Numerical weights: %v\n", tc.numericalWeights)
		}
		if len(tc.ratingWeights) > 0 {
			fmt.Printf("Rating weights: %v\n", tc.ratingWeights)
		}

		// Show sample results with detailed query information
		if len(stocks) > 0 {
			fmt.Printf("\nSample results (first 3):\n")
			maxSamples := 3
			if len(stocks) < maxSamples {
				maxSamples = len(stocks)
			}
			for i := 0; i < maxSamples; i++ {
				sample := stocks[i]
				fmt.Printf("\n  [Sample %d]:\n", i+1)
				fmt.Printf("    ID: %d\n", sample.ID)
				fmt.Printf("    Ticker: %s\n", sample.Ticker)
				fmt.Printf("    Company: %s\n", sample.Company)
				fmt.Printf("    Action: %s\n", sample.Action)
				fmt.Printf("    Date: %s\n", sample.Date.Format("2006-01-02"))
				fmt.Printf("    Cluster: %d\n", sample.Cluster)
				fmt.Printf("    FinalScore: %.6f\n", sample.FinalScore)
				fmt.Printf("    TargetTo: %.6f\n", sample.TargetTo)
				fmt.Printf("    TargetFrom: %.6f\n", sample.TargetFrom)
				fmt.Printf("    RatingTo: %s\n", sample.RatingTo)
				fmt.Printf("    RatingFrom: %s\n", sample.RatingFrom)
				if sample.WeightedScore != nil {
					fmt.Printf("    WeightedScore: %.6f\n", *sample.WeightedScore)
				} else {
					fmt.Printf("    WeightedScore: nil (not calculated)\n")
				}
				fmt.Printf("    RatingSentiments count: %d\n", len(sample.RatingSentiments))
				if len(sample.RatingSentiments) > 0 {
					fmt.Printf("    First RatingSentiment: Name=%s, Rating=%s, NormRatingScore=%.4f\n",
						sample.RatingSentiments[0].Name, sample.RatingSentiments[0].Rating, sample.RatingSentiments[0].NormRatingScore)
				}
				fmt.Printf("    NumericalIndicators count: %d\n", len(sample.NumericalIndicators))
				if len(sample.NumericalIndicators) > 0 {
					fmt.Printf("    First NumericalIndicator: Name=%s, NormValue=%.6f\n",
						sample.NumericalIndicators[0].Name, sample.NumericalIndicators[0].NormValue)
				}
			}
		}

		fmt.Printf("=== End Test: %s ===\n", tc.name)
	}
}
