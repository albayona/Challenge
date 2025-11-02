package utils

import (
	"strconv"
	"time"
)

// GetCSVValue returns the value for a given header key from a CSV row using the provided header index map.
func GetCSVValue(row []string, headerIndex map[string]int, key string) string {
	if pos, ok := headerIndex[key]; ok && pos < len(row) {
		return row[pos]
	}
	return ""
}

// ParseFloat converts string to float64; returns 0 for empty or invalid input.
func ParseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ParseInt converts string to int; returns 0 for empty or invalid input.
func ParseInt(s string) int {
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}

// ParseTime attempts to parse time; tries RFC3339 from timeStr, then date-only (YYYY-MM-DD), else returns time.Now().
func ParseTime(dateStr, timeStr string) time.Time {
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t
	}
	if dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			return d
		}
	}
	return time.Now()
}




