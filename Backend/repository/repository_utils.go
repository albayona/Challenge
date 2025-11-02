package repository

import (
	"fmt"
	"strings"
)

// weightEntry represents a generic weight entry structure
type weightEntry struct {
	IndicatorName string
	Weight        float64
}

// validateColumnName checks if a column name is in the allowed whitelist
func validateColumnName(colName string, allowedCols []string) bool {
	colName = strings.TrimSpace(strings.ToLower(colName))
	for _, allowed := range allowedCols {
		if strings.TrimSpace(strings.ToLower(allowed)) == colName {
			return true
		}
	}
	return false
}

// escapeSQLString escapes a string for safe SQL usage (PostgreSQL/CockroachDB compatible)
func escapeSQLString(s string) string {
	// Replace single quotes with escaped quotes
	return strings.ReplaceAll(s, "'", "''")
}

// convertNumericalWeights converts NumericalWeightEntry slice to weightEntry slice
func convertNumericalWeights(weights []NumericalWeightEntry) []weightEntry {
	if len(weights) == 0 {
		return nil
	}
	result := make([]weightEntry, len(weights))
	for i := range weights {
		result[i] = weightEntry{
			IndicatorName: weights[i].IndicatorName,
			Weight:        weights[i].Weight,
		}
	}
	return result
}

// convertRatingWeights converts RatingWeightEntry slice to weightEntry slice
func convertRatingWeights(weights []RatingWeightEntry) []weightEntry {
	if len(weights) == 0 {
		return nil
	}
	result := make([]weightEntry, len(weights))
	for i := range weights {
		result[i] = weightEntry{
			IndicatorName: weights[i].IndicatorName,
			Weight:        weights[i].Weight,
		}
	}
	return result
}

// buildWeightedScoreSubquery builds a subquery for calculating weighted scores
// tableName: the table to query (e.g., "numerical_indicators" or "rating_sentiments")
// valueColumn: the column containing the values to weight (e.g., "norm_value" or "norm_rating_score")
// scoreAlias: the alias for the calculated score (e.g., "new_indicator_score" or "new_rating_score")
// tableAlias: the alias for the table in the subquery (e.g., "ni_sub" or "rs_sub")
// weights: slice of weight entries with IndicatorName and Weight
func buildWeightedScoreSubquery(tableName, valueColumn, scoreAlias, tableAlias string, weights []weightEntry) string {
	if len(weights) == 0 {
		return ""
	}

	// Build CASE expression with all weights
	caseExpr := "COALESCE(SUM(CASE"
	for _, weight := range weights {
		escapedName := escapeSQLString(weight.IndicatorName)
		caseExpr += fmt.Sprintf(" WHEN %s.name = '%s' THEN %s.%s * %.6f", tableAlias, escapedName, tableAlias, valueColumn, weight.Weight)
	}
	caseExpr += fmt.Sprintf(" ELSE 0 END), 0) AS %s", scoreAlias)

	// Build and return the subquery
	return fmt.Sprintf(`(
		SELECT %s.stock_data_point_id, %s
		FROM %s %s
		GROUP BY %s.stock_data_point_id
	)`, tableAlias, caseExpr, tableName, tableAlias, tableAlias)
}

// combineWeightedScoreSubqueries combines indicator and rating subqueries into a single combined subquery
// indicatorSubquery: subquery for numerical indicators (can be empty)
// ratingSubquery: subquery for rating sentiments (can be empty)
// Returns the combined subquery that sums both scores, or returns a single subquery if only one exists
func combineWeightedScoreSubqueries(indicatorSubquery, ratingSubquery string) string {
	if indicatorSubquery != "" && ratingSubquery != "" {
		// Both exist: join them and sum
		return fmt.Sprintf(`(
			SELECT 
				COALESCE(i.stock_data_point_id, r.stock_data_point_id) AS stock_data_point_id,
				COALESCE(i.new_indicator_score, 0) + COALESCE(r.new_rating_score, 0) AS weighted_score
			FROM %s i
			FULL OUTER JOIN %s r ON i.stock_data_point_id = r.stock_data_point_id
		)`, indicatorSubquery, ratingSubquery)
	} else if indicatorSubquery != "" {
		// Only indicator exists
		return fmt.Sprintf(`(
			SELECT stock_data_point_id, new_indicator_score AS weighted_score
			FROM %s
		)`, indicatorSubquery)
	} else if ratingSubquery != "" {
		// Only rating exists
		return fmt.Sprintf(`(
			SELECT stock_data_point_id, new_rating_score AS weighted_score
			FROM %s
		)`, ratingSubquery)
	}
	// Neither exists (shouldn't happen if called correctly)
	return ""
}
