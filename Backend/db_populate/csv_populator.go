package db_populate

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"dataextractor/models"
	"dataextractor/repository"
	"dataextractor/utils"
)

// GetColIndexByName reads the CSV header and returns a header->index map
func GetColIndexByName(csvr *csv.Reader) map[string]int {
	headers, err := csvr.Read()
	utils.ErrorPanic(err, "failed to read CSV header")

	idx := map[string]int{}
	for i, h := range headers {
		idx[h] = i
	}
	return idx
}

// GetRatingColsValues builds a map of rating column values from a row
func GetRatingColsValues(ratingColsNames []string, row []string, idx map[string]int) map[string]string {
	values := map[string]string{}
	for _, col := range ratingColsNames {
		values[col] = utils.GetCSVValue(row, idx, col)
	}
	return values
}

// GetNumericalColsValues builds a map of numerical column values from a row
func GetNumericalColsValues(numericalColsNames []string, row []string, idx map[string]int) map[string]string {
	values := map[string]string{}
	for _, col := range numericalColsNames {
		values[col] = utils.GetCSVValue(row, idx, col)
	}
	return values
}

// GetRatingScoresAndNormScores returns rating scores and normalized rating scores maps
func GetRatingScoresAndNormScores(ratingColsNames []string, row []string, idx map[string]int) (map[string]string, map[string]string) {
	ratingScores := map[string]string{}
	normRatingScores := map[string]string{}
	for _, name := range ratingColsNames {
		var scoreKey, normScoreKey string
		switch name {
		case "rating_from":
			scoreKey = "rating_from_score"
			normScoreKey = "norm_rating_from_score"
		case "rating_to":
			scoreKey = "rating_to_score"
			normScoreKey = "norm_rating_to_score"
		case "action":
			scoreKey = "rating_delta"
			normScoreKey = "norm_rating_delta"
		default:
			scoreKey = name
			normScoreKey = "norm_" + name
		}
		ratingScores[name] = utils.GetCSVValue(row, idx, scoreKey)
		normRatingScores[name] = utils.GetCSVValue(row, idx, normScoreKey)
	}
	return ratingScores, normRatingScores
}

// GetNormNumericalValues builds a map of normalized numerical values (using norm_ prefix)
func GetNormNumericalValues(numericalColsNames []string, row []string, idx map[string]int) map[string]string {
	values := map[string]string{}
	for _, col := range numericalColsNames {
		values[col] = utils.GetCSVValue(row, idx, "norm_"+col)
	}
	return values
}

// CreateDataPoint builds a StockDataPoint base struct from the row
func CreateDataPoint(row []string, idx map[string]int, ratingColsValues map[string]string) *models.StockDataPoint {
	return &models.StockDataPoint{
		Ticker:     utils.GetCSVValue(row, idx, "ticker"),
		Company:    utils.GetCSVValue(row, idx, "company"),
		Action:     utils.GetCSVValue(row, idx, "action"),
		Cluster:    utils.ParseInt(utils.GetCSVValue(row, idx, "cluster")),
		Date:       utils.ParseTime(utils.GetCSVValue(row, idx, "date"), utils.GetCSVValue(row, idx, "time")),
		TargetTo:   utils.ParseFloat(utils.GetCSVValue(row, idx, "target_to")),
		TargetFrom: utils.ParseFloat(utils.GetCSVValue(row, idx, "target_from")),
		TargetDelta: utils.ParseFloat(utils.GetCSVValue(row, idx, "target_delta")),
		LastClose: utils.ParseFloat(utils.GetCSVValue(row, idx, "last_close")),
		RatingTo:   ratingColsValues["rating_to"],
		RatingFrom: ratingColsValues["rating_from"],
		FinalScore: utils.ParseFloat(utils.GetCSVValue(row, idx, "final_score")),
	}
}

// CreateSentimentsArray builds RatingSentiment slice from rating maps
func CreateSentimentsArray(ratingColsNames []string, ratingScores map[string]string, normRatingScores map[string]string, ratingColsValues map[string]string) []models.RatingSentiment {
	sentiments := make([]models.RatingSentiment, 0, len(ratingColsNames))
	for _, name := range ratingColsNames {
		rs := strings.TrimSpace(ratingScores[name])
		nrs := strings.TrimSpace(normRatingScores[name])
		if rs == "" && nrs == "" && strings.TrimSpace(ratingColsValues[name]) == "" {
			continue
		}
		sentiments = append(sentiments, models.RatingSentiment{
			Name:            name,
			Rating:          ratingColsValues[name],
			RatingScore:     utils.ParseFloat(rs),
			NormRatingScore: utils.ParseFloat(nrs),
		})
	}
	return sentiments
}

// CreateIndicatorsArray builds NumericalIndicator slice from numeric maps
func CreateIndicatorsArray(numericalColsNames []string, numericalColsValues map[string]string, normNumericalColsValues map[string]string) []models.NumericalIndicator {
	indicators := make([]models.NumericalIndicator, 0, len(numericalColsNames))
	for _, name := range numericalColsNames {
		v := strings.TrimSpace(numericalColsValues[name])
		nv := strings.TrimSpace(normNumericalColsValues[name])
		if v == "" && nv == "" {
			continue
		}
		indicators = append(indicators, models.NumericalIndicator{
			Name:      name,
			Value:     utils.ParseFloat(v),
			NormValue: utils.ParseFloat(nv),
		})
	}
	return indicators
}

// ImportFromCSV reads a CSV and builds StockDataPoint entries (no persistence yet)
func ImportFromCSV(reader io.Reader, repo repository.DataRepositoryInterface) (int, error) {
	csvr := csv.NewReader(reader)
	csvr.TrimLeadingSpace = true
	csvr.ReuseRecord = false

	idx := GetColIndexByName(csvr)

	ratingColsNames := []string{
		"rating_from",
		"rating_to",
		"action",
	}

	numericalColsNames := []string{
		"target_from", "target_to", "target_delta", "target_growth", "relative_growth",
		"last_close",
		"atr", "std_dev", "ulcer_index", "price_distance", "obv", "ad_line", "pvt", "force_index",
		"hlc3", "typical_price", "vwap",
	}

	count := 0
	for {
		row, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf("failed to read CSV row: %w", err)
		}

		ratingColsValues := GetRatingColsValues(ratingColsNames, row, idx)
		numericalColsValues := GetNumericalColsValues(numericalColsNames, row, idx)

		ratingScores, normRatingScores := GetRatingScoresAndNormScores(ratingColsNames, row, idx)
		normNumericalColsValues := GetNormNumericalValues(numericalColsNames, row, idx)
		sdp := CreateDataPoint(row, idx, ratingColsValues)

		sentiments := CreateSentimentsArray(ratingColsNames, ratingScores, normRatingScores, ratingColsValues)
		sdp.RatingSentiments = sentiments

		indicators := CreateIndicatorsArray(numericalColsNames, numericalColsValues, normNumericalColsValues)
		sdp.NumericalIndicators = indicators

		if _, err := repo.UpdateOrCreate(sdp); err != nil {
			return count, fmt.Errorf("failed to persist row for ticker %s: %w", sdp.Ticker, err)
		}

		count++
	}

	return count, nil
}
