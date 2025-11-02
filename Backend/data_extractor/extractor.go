package data_extractor

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"dataextractor/repository"
	"dataextractor/utils"
)

// File constants for data storage
const (
	resumeKeyFile       = "processed_pages.txt"
	lastPageFile        = "last_page.txt"
	pageKeysHistoryFile = "page_keys_history.txt"
	csvOutputFile       = "extracted_stock_data.csv"
)

// API endpoint constants
const (
	baseEndpoint = "/swechallenge/list"
)

// Default values
const (
	NoPageLimit = math.MaxInt // Represents no page limit
)

// OldStock represents the legacy data point shape returned by the API
type OldStock struct {
	Ticker     string    `json:"ticker"`
	Company    string    `json:"company"`
	TargetFrom float64   `json:"target_from"`
	TargetTo   float64   `json:"target_to"`
	Action     string    `json:"action"`
	Brokerage  string    `json:"brokerage"`
	RatingFrom string    `json:"rating_from"`
	RatingTo   string    `json:"rating_to"`
	Time       time.Time `json:"time"`
}

// APIResponse represents the response structure from the API
type APIResponse struct {
	Items    []OldStock `json:"items"`
	NextPage string     `json:"next_page"`
}

// DataExtractor handles API data extraction
type DataExtractor struct {
	client     *http.Client
	baseURL    string
	apiKey     string
	repository repository.DataRepositoryInterface
}

// NewDataExtractor creates a new DataExtractor instance
func NewDataExtractor(baseURL, apiKey string, repository repository.DataRepositoryInterface) *DataExtractor {
	return &DataExtractor{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    baseURL,
		apiKey:     apiKey,
		repository: repository,
	}
}

// FetchData retrieves data from the API
func (de *DataExtractor) FetchData(endpoint string) (*APIResponse, error) {
	url := de.baseURL + endpoint

	req, err := createRequest(url, de)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	log.Printf("Fetching data from: %s", url)

	resp, err := de.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	utils.ErrorPanic(err, "failed to read response body")

	// Parse JSON response
	var apiResponse APIResponse
	utils.ErrorPanic(json.Unmarshal(body, &apiResponse), "failed to parse JSON response")

	return &apiResponse, nil
}

func createRequest(url string, de *DataExtractor) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	utils.ErrorPanic(err, "failed to create request")

	// Add authentication header
	if de.apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	req.Header.Set("Authorization", "Bearer "+de.apiKey)

	// Add common headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "DataExtractor/1.0")
	return req, nil
}

// updateResumeKeyFile saves the current page key to the resume file (overwrites previous value)
func updateResumeKeyFile(pageKey string) error {
	file, err := os.OpenFile(lastPageFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	utils.ErrorPanic(err, "failed to open resume file")
	defer file.Close()

	_, err = file.WriteString(pageKey)
	utils.ErrorPanic(err, "failed to write page key to resume file")
	log.Printf("Updated resume file with next page token: %s", pageKey)

	return nil
}

// savePageKeyToHistory saves a page key to the history file in CSV format
func savePageKeyToHistory(pageKey string, pageNumber int, status string) error {
	// Check if file exists to determine if we need to write header
	fileExists := true
	if _, err := os.Stat(pageKeysHistoryFile); os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(pageKeysHistoryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	utils.ErrorPanic(err, "failed to open page keys history file")
	defer file.Close()

	// Write CSV header if file is new
	if !fileExists {
		_, err = file.WriteString("key,page_number,date,status\n")
		utils.ErrorPanic(err, "failed to write CSV header")
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = file.WriteString(fmt.Sprintf("%s,%d,%s,%s\n", pageKey, pageNumber, timestamp, status))
	utils.ErrorPanic(err, "failed to write page key to history file")

	return nil
}

// ExtractAndProcessAllPages processes all pages of data from the API
// maxPages: maximum number of pages to process (0 means no limit, default infinity)
func (de *DataExtractor) ExtractAndProcessAllPages(maxPages int) error {
	// Set default to infinity if maxPages is 0
	if maxPages == 0 {
		maxPages = NoPageLimit
	}

	nextPage := de.getResumePage()

	totalProcessed := 0
	pageCount := 1

	for {

		if pageCount > maxPages {
			log.Printf("Reached maximum page limit of %d pages", maxPages)
			break
		}

		endpoint := de.buildEndpoint(nextPage)

		log.Printf("Processing page %d (key: %s)...", pageCount, nextPage)

		apiResponse, err := de.FetchData(endpoint)

		if err != nil {
			// Save page key to history file with error status
			if saveErr := savePageKeyToHistory(nextPage, pageCount+1, "error"); saveErr != nil {
				log.Printf("Warning: Failed to save error page key to history: %v", saveErr)
			}
			return fmt.Errorf("failed to fetch page %d: %w", pageCount, err)
		}

		log.Printf("Retrieved %d items from page %d", len(apiResponse.Items), pageCount)

		successCount := 0
		for _, item := range apiResponse.Items {
			if err := de.writeToCSV(&item); err != nil {
				log.Printf("Warning: Failed to write data point %s to CSV: %v", item.Ticker, err)
			} else {
				successCount++
				totalProcessed++
			}
		}

		log.Printf("Successfully wrote %d out of %d items from page %d to CSV", successCount, len(apiResponse.Items), pageCount)

		nextPage = apiResponse.NextPage

		if err := updateResumeKeyFile(nextPage); err != nil {
			log.Printf("Warning: Failed to save resume page key %s: %v", nextPage, err)
		}

		// Save page key to history file with success status
		if err := savePageKeyToHistory(nextPage, pageCount+1, "success"); err != nil {
			log.Printf("Warning: Failed to save page key to history: %v", err)
		}

		pageCount++

		if nextPage == "" {
			log.Println("No more pages to process")
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	log.Printf("Data extraction completed! Total items written to CSV: %d across %d pages", totalProcessed, pageCount)
	return nil
}

func (*DataExtractor) getResumePage() string {
	nextPage := ""
	if data, err := os.ReadFile(lastPageFile); err == nil {
		nextPage = strings.TrimSpace(string(data))
		log.Printf("Resuming from last page: %s", nextPage)
	} else {
		log.Println("No previous page found, starting from the beginning")
	}
	return nextPage
}

func (*DataExtractor) buildEndpoint(nextPage string) string {
	var endpoint string

	if nextPage == "" {
		// First page - no parameters
		endpoint = baseEndpoint
	} else {
		// Subsequent pages - with next_page parameter
		endpoint = fmt.Sprintf("%s?next_page=%s", baseEndpoint, nextPage)
	}
	return endpoint
}

// writeToCSV writes a stock item to the CSV file
func (de *DataExtractor) writeToCSV(item *OldStock) error {
	// Check if CSV file exists to determine if we need to write headers
	fileExists := false
	if _, err := os.Stat(csvOutputFile); err == nil {
		fileExists = true
	}

	// Open CSV file for appending
	file, err := os.OpenFile(csvOutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers if file is new
	if !fileExists {
		headers := []string{
			"ticker",
			"company",
			"target_from",
			"target_to",
			"action",
			"brokerage",
			"rating_from",
			"rating_to",
			"time",
		}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write CSV headers: %w", err)
		}
	}

	// Write stock data
	record := []string{
		item.Ticker,
		item.Company,
		fmt.Sprintf("%.2f", item.TargetFrom),
		fmt.Sprintf("%.2f", item.TargetTo),
		item.Action,
		item.Brokerage,
		item.RatingFrom,
		item.RatingTo,
		item.Time.Format("2006-01-02 15:04:05"),
	}

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write CSV record: %w", err)
	}

	return nil
}
