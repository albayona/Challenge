package validators

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// RatingSentimentRequest captures a sentiment entry
type RatingSentimentRequest struct {
	Name            string  `json:"name" validate:"required,min=1,max=100"`
	Rating          string  `json:"rating" validate:"required,min=1,max=50"`
	RatingScore     float64 `json:"rating_score" validate:"required"`
	NormRatingScore float64 `json:"norm_rating_score" validate:"required"`
}

// NumericalIndicatorRequest captures a numerical indicator
type NumericalIndicatorRequest struct {
	Name      string  `json:"name" validate:"required,min=1,max=100"`
	Value     float64 `json:"value" validate:"required"`
	NormValue float64 `json:"norm_value" validate:"required"`
}

// StockRequest represents the request structure for stock operations with validation
type StockRequest struct {
	ID                  uint                        `json:"id" validate:"omitempty,min=1"`
	Ticker              string                      `json:"ticker" validate:"required,min=1,max=20,alphanum"`
	Company             string                      `json:"company" validate:"required,min=1,max=100"`
	Action              string                      `json:"action" validate:"omitempty,max=100"`
	Date                time.Time                   `json:"date" validate:"required"`
	Cluster             int                         `json:"cluster" validate:"required"`
	TargetTo            float64                     `json:"target_to" validate:"omitempty"`
	TargetFrom          float64                     `json:"target_from" validate:"omitempty"`
	TargetDelta         float64                     `json:"target_delta" validate:"omitempty"`
	LastClose           float64                     `json:"last_close" validate:"omitempty"`
	RatingTo            string                      `json:"rating_to" validate:"omitempty,max=50"`
	RatingFrom          string                      `json:"rating_from" validate:"omitempty,max=50"`
	RatingSentiments    []RatingSentimentRequest    `json:"rating_sentiments" validate:"dive"`
	NumericalIndicators []NumericalIndicatorRequest `json:"numerical_indicators" validate:"dive"`
}

// StockCreateRequest represents the request structure for creating a new stock
type StockCreateRequest struct {
	Ticker              string                      `json:"ticker" validate:"required,min=1,max=20,alphanum"`
	Company             string                      `json:"company" validate:"required,min=1,max=100"`
	Action              string                      `json:"action" validate:"omitempty,max=100"`
	Date                time.Time                   `json:"date" validate:"required"`
	Cluster             int                         `json:"cluster" validate:"required"`
	TargetTo            float64                     `json:"target_to" validate:"omitempty"`
	TargetFrom          float64                     `json:"target_from" validate:"omitempty"`
	TargetDelta         float64                     `json:"target_delta" validate:"omitempty"`
	LastClose           float64                     `json:"last_close" validate:"omitempty"`
	RatingTo            string                      `json:"rating_to" validate:"omitempty,max=50"`
	RatingFrom          string                      `json:"rating_from" validate:"omitempty,max=50"`
	RatingSentiments    []RatingSentimentRequest    `json:"rating_sentiments" validate:"dive"`
	NumericalIndicators []NumericalIndicatorRequest `json:"numerical_indicators" validate:"dive"`
}

// StockUpdateRequest represents the request structure for updating a stock
type StockUpdateRequest struct {
	ID                  uint                        `json:"id" validate:"required,min=1"`
	Ticker              string                      `json:"ticker" validate:"required,min=1,max=20,alphanum"`
	Company             string                      `json:"company" validate:"required,min=1,max=100"`
	Action              string                      `json:"action" validate:"omitempty,max=100"`
	Date                time.Time                   `json:"date" validate:"required"`
	Cluster             int                         `json:"cluster" validate:"required"`
	TargetTo            float64                     `json:"target_to" validate:"omitempty"`
	TargetFrom          float64                     `json:"target_from" validate:"omitempty"`
	TargetDelta         float64                     `json:"target_delta" validate:"omitempty"`
	LastClose           float64                     `json:"last_close" validate:"omitempty"`
	RatingTo            string                      `json:"rating_to" validate:"omitempty,max=50"`
	RatingFrom          string                      `json:"rating_from" validate:"omitempty,max=50"`
	RatingSentiments    []RatingSentimentRequest    `json:"rating_sentiments" validate:"dive"`
	NumericalIndicators []NumericalIndicatorRequest `json:"numerical_indicators" validate:"dive"`
}

// StockExtractRequest represents the request structure for data extraction
type StockExtractRequest struct {
	MaxPages int `json:"max_pages" validate:"required,min=0"`
}

// StockValidator handles validation for stock-related requests
type StockValidator struct {
	validator *validator.Validate
}

// NewStockValidator creates a new StockValidator instance
func NewStockValidator() *StockValidator {
	return &StockValidator{
		validator: validator.New(),
	}
}

// ValidateRequest validates any request struct using the validator
func (sv *StockValidator) ValidateRequest(request interface{}) error {
	return sv.validator.Struct(request)
}

// ValidateTicker validates a ticker string
func (sv *StockValidator) ValidateTicker(ticker string) error {
	return sv.validator.Var(ticker, "required,min=1,max=20,alphanum")
}

// ValidateCompany validates a company string
func (sv *StockValidator) ValidateCompany(company string) error {
	return sv.validator.Var(company, "required,min=1,max=100")
}

// ValidateID validates an ID
func (sv *StockValidator) ValidateID(id uint) error {
	return sv.validator.Var(id, "required,min=1")
}
