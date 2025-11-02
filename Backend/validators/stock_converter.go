package validators

import (
	"dataextractor/models"
	"time"
)

// ToStockRequest converts a Stock model to StockRequest
func (sr *StockRequest) ToStockRequest(stock *models.StockDataPoint) *StockRequest {
	return &StockRequest{
		ID:                  stock.ID,
		Ticker:              stock.Ticker,
		Company:             stock.Company,
		Action:              stock.Action,
		Date:                stock.Date,
		Cluster:             stock.Cluster,
		TargetTo:            stock.TargetTo,
		TargetFrom:          stock.TargetFrom,
		TargetDelta:         stock.TargetDelta,
		LastClose:           stock.LastClose,
		RatingTo:            stock.RatingTo,
		RatingFrom:          stock.RatingFrom,
		RatingSentiments:    toRatingSentimentRequests(stock.RatingSentiments),
		NumericalIndicators: toNumericalIndicatorRequests(stock.NumericalIndicators),
	}
}

// ToStock converts a StockRequest to Stock model
func (sr *StockRequest) ToStock() *models.StockDataPoint {
	return &models.StockDataPoint{
		ID:                  sr.ID,
		Ticker:              sr.Ticker,
		Company:             sr.Company,
		Action:              sr.Action,
		Date:                sr.Date,
		Cluster:             sr.Cluster,
		TargetTo:            sr.TargetTo,
		TargetFrom:          sr.TargetFrom,
		TargetDelta:         sr.TargetDelta,
		LastClose:           sr.LastClose,
		RatingTo:            sr.RatingTo,
		RatingFrom:          sr.RatingFrom,
		RatingSentiments:    toRatingSentiments(sr.RatingSentiments),
		NumericalIndicators: toNumericalIndicators(sr.NumericalIndicators),
	}
}

// ToStockCreateRequest converts a Stock model to StockCreateRequest
func (scr *StockCreateRequest) ToStockCreateRequest(stock *models.StockDataPoint) *StockCreateRequest {
	return &StockCreateRequest{
		Ticker:              stock.Ticker,
		Company:             stock.Company,
		Action:              stock.Action,
		Date:                stock.Date,
		Cluster:             stock.Cluster,
		TargetTo:            stock.TargetTo,
		TargetFrom:          stock.TargetFrom,
		TargetDelta:         stock.TargetDelta,
		LastClose:           stock.LastClose,
		RatingTo:            stock.RatingTo,
		RatingFrom:          stock.RatingFrom,
		RatingSentiments:    toRatingSentimentRequests(stock.RatingSentiments),
		NumericalIndicators: toNumericalIndicatorRequests(stock.NumericalIndicators),
	}
}

// ToStock converts a StockCreateRequest to Stock model
func (scr *StockCreateRequest) ToStock() *models.StockDataPoint {
	return &models.StockDataPoint{
		Ticker:              scr.Ticker,
		Company:             scr.Company,
		Action:              scr.Action,
		Date:                scr.Date,
		Cluster:             scr.Cluster,
		TargetTo:            scr.TargetTo,
		TargetFrom:          scr.TargetFrom,
		TargetDelta:         scr.TargetDelta,
		LastClose:           scr.LastClose,
		RatingTo:            scr.RatingTo,
		RatingFrom:          scr.RatingFrom,
		RatingSentiments:    toRatingSentiments(scr.RatingSentiments),
		NumericalIndicators: toNumericalIndicators(scr.NumericalIndicators),
	}
}

// ToStockUpdateRequest converts a Stock model to StockUpdateRequest
func (sur *StockUpdateRequest) ToStockUpdateRequest(stock *models.StockDataPoint) *StockUpdateRequest {
	return &StockUpdateRequest{
		ID:                  stock.ID,
		Ticker:              stock.Ticker,
		Company:             stock.Company,
		Action:              stock.Action,
		Date:                stock.Date,
		Cluster:             stock.Cluster,
		TargetTo:            stock.TargetTo,
		TargetFrom:          stock.TargetFrom,
		TargetDelta:         stock.TargetDelta,
		LastClose:           stock.LastClose,
		RatingTo:            stock.RatingTo,
		RatingFrom:          stock.RatingFrom,
		RatingSentiments:    toRatingSentimentRequests(stock.RatingSentiments),
		NumericalIndicators: toNumericalIndicatorRequests(stock.NumericalIndicators),
	}
}

// ToStock converts a StockUpdateRequest to Stock model
func (sur *StockUpdateRequest) ToStock() *models.StockDataPoint {
	return &models.StockDataPoint{
		ID:                  sur.ID,
		Ticker:              sur.Ticker,
		Company:             sur.Company,
		Action:              sur.Action,
		Date:                sur.Date,
		Cluster:             sur.Cluster,
		TargetTo:            sur.TargetTo,
		TargetFrom:          sur.TargetFrom,
		TargetDelta:         sur.TargetDelta,
		LastClose:           sur.LastClose,
		RatingTo:            sur.RatingTo,
		RatingFrom:          sur.RatingFrom,
		RatingSentiments:    toRatingSentiments(sur.RatingSentiments),
		NumericalIndicators: toNumericalIndicators(sur.NumericalIndicators),
	}
}

// NewStockCreateRequest creates a new StockCreateRequest with default values
func NewStockCreateRequest(ticker, company string) *StockCreateRequest {
	return &StockCreateRequest{
		Ticker:  ticker,
		Company: company,
		Date:    time.Now(),
	}
}

// NewStockUpdateRequest creates a new StockUpdateRequest
func NewStockUpdateRequest(id uint, ticker, company string) *StockUpdateRequest {
	return &StockUpdateRequest{
		ID:      id,
		Ticker:  ticker,
		Company: company,
		Date:    time.Now(),
	}
}

// Helpers: map between request slices and model slices
func toRatingSentiments(reqs []RatingSentimentRequest) []models.RatingSentiment {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]models.RatingSentiment, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, models.RatingSentiment{
			Name:            r.Name,
			Rating:          r.Rating,
			RatingScore:     r.RatingScore,
			NormRatingScore: r.NormRatingScore,
		})
	}
	return out
}

func toNumericalIndicators(reqs []NumericalIndicatorRequest) []models.NumericalIndicator {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]models.NumericalIndicator, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, models.NumericalIndicator{
			Name:      r.Name,
			Value:     r.Value,
			NormValue: r.NormValue,
		})
	}
	return out
}

func toRatingSentimentRequests(modelsIn []models.RatingSentiment) []RatingSentimentRequest {
	if len(modelsIn) == 0 {
		return nil
	}
	out := make([]RatingSentimentRequest, 0, len(modelsIn))
	for _, m := range modelsIn {
		out = append(out, RatingSentimentRequest{
			Name:            m.Name,
			Rating:          m.Rating,
			RatingScore:     m.RatingScore,
			NormRatingScore: m.NormRatingScore,
		})
	}
	return out
}

func toNumericalIndicatorRequests(modelsIn []models.NumericalIndicator) []NumericalIndicatorRequest {
	if len(modelsIn) == 0 {
		return nil
	}
	out := make([]NumericalIndicatorRequest, 0, len(modelsIn))
	for _, m := range modelsIn {
		out = append(out, NumericalIndicatorRequest{
			Name:      m.Name,
			Value:     m.Value,
			NormValue: m.NormValue,
		})
	}
	return out
}
