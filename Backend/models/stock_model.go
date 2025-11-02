package models

import (
	"time"
)

// StockDataPoint represents a stock data point with related sentiments and indicators
type StockDataPoint struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Ticker      string    `json:"ticker" gorm:"size:20;not null;uniqueIndex"`
	Action      string    `json:"action" gorm:"size:100"`
	Date        time.Time `json:"date" gorm:"not null;index"`
	Company     string    `json:"company" gorm:"size:100;not null;index"`
	Cluster     int       `json:"cluster" gorm:"not null"`
	TargetTo    float64   `json:"target_to" gorm:"type:decimal(18,6)"`
	TargetFrom  float64   `json:"target_from" gorm:"type:decimal(18,6)"`
	TargetDelta float64   `json:"target_delta" gorm:"type:decimal(18,6)"`
	LastClose   float64   `json:"last_close" gorm:"type:decimal(18,6)"`
	RatingTo    string    `json:"rating_to" gorm:"size:50"`
	RatingFrom  string    `json:"rating_from" gorm:"size:50"`
	FinalScore  float64   `json:"final_score" gorm:"type:decimal(18,6);not null;default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	RatingSentiments    []RatingSentiment    `json:"rating_sentiments" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	NumericalIndicators []NumericalIndicator `json:"numerical_indicators" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Computed field from queries (not persisted)
	// No gorm tag - GORM will map weighted_score column (snake_case) to WeightedScore field (PascalCase) automatically
	// This field is never written to the database, only populated from SELECT queries
	WeightedScore *float64 `json:"weighted_score,omitempty"`
}

// TableName returns the table name for StockDataPoint
func (StockDataPoint) TableName() string {
	return "stock_data_points"
}

// RatingSentiment represents a qualitative rating with scores
type RatingSentiment struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	StockDataPointID uint      `json:"stock_data_point_id" gorm:"not null;uniqueIndex:idx_stock_rating_unique"`
	Name             string    `json:"name" gorm:"size:100;not null;uniqueIndex:idx_stock_rating_unique"`
	Rating           string    `json:"rating" gorm:"size:50;not null"`
	RatingScore      float64   `json:"rating_score" gorm:"type:decimal(10,4);not null"`
	NormRatingScore  float64   `json:"norm_rating_score" gorm:"type:decimal(10,4);not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for RatingSentiment
func (RatingSentiment) TableName() string {
	return "rating_sentiments"
}

// NumericalIndicator represents a quantitative indicator value
type NumericalIndicator struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	StockDataPointID uint      `json:"stock_data_point_id" gorm:"not null;uniqueIndex:idx_stock_indicator_unique"`
	Name             string    `json:"name" gorm:"size:100;not null;uniqueIndex:idx_stock_indicator_unique"`
	Value            float64   `json:"value" gorm:"type:decimal(18,6);not null"`
	NormValue        float64   `json:"norm_value" gorm:"type:decimal(18,6);not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for NumericalIndicator
func (NumericalIndicator) TableName() string {
	return "numerical_indicators"
}

// (alias removed; use StockDataPoint directly)
