package checker

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"enhanced-gateway-scraper/pkg/types"
)

// Checker manages the account checking process
type Checker struct {
	Config *types.CheckerConfig
}

// NewChecker returns a new Checker
func NewChecker(config *types.CheckerConfig) *Checker {
	return &Checker{
		Config: config,
	}
}

// CheckAccounts performs account checks based on provided configurations
func (c *Checker) CheckAccounts(ctx context.Context, combos []types.Combo, configs []types.Config) []types.CheckResult {
	results := make([]types.CheckResult, 0)
	var wg sync.WaitGroup
	comboChan := make(chan types.Combo, len(combos))
	resultChan := make(chan types.CheckResult, len(combos))

	// Start workers
	for i := 0; i < c.Config.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.worker(ctx, comboChan, resultChan, configs)
		}()
	}

	// Send combos to the channel
	go func() {
		for _, combo := range combos {
			comboChan <- combo
		}
		close(comboChan)
	}()

	// Wait for the workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func (c *Checker) worker(ctx context.Context, comboChan <-chan types.Combo, resultChan chan<- types.CheckResult, configs []types.Config) {
	for combo := range comboChan {
		for _, config := range configs {
			result := c.checkCombo(ctx, combo, config)
			resultChan <- result
		}
	}
}

func (c *Checker) checkCombo(ctx context.Context, combo types.Combo, config types.Config) types.CheckResult {
	// Create HTTP client
	client := &http.Client{
		Timeout: time.Duration(c.Config.RequestTimeout) * time.Millisecond,
	}

	// Create request
	url := c.buildURL(combo, config)
	req, err := http.NewRequest(config.Method, url, nil)
	if err != nil {
		return types.CheckResult{
			Combo:  combo,
			Config: config.Name,
			Status: "error",
			Error:  err.Error(),
		}
	}

	// Set headers and form data
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	// Execute the request
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return types.CheckResult{
			Combo:    combo,
			Config:   config.Name,
			Status:   "error",
			Error:    err.Error(),
			Latency:  int(latency),
			Response: "",
		}
	}
	defer resp.Body.Close()

	// Analyze response
	response, _ := ioutil.ReadAll(resp.Body)
	status := c.analyzeResponse(string(response), resp.StatusCode, config)

	return types.CheckResult{
		Combo:    combo,
		Config:   config.Name,
		Status:   status,
		Response: string(response),
		Latency:  int(latency),
	}
}

// buildURL constructs the URL with the combo values
func (c *Checker) buildURL(combo types.Combo, config types.Config) string {
	url, _ := url.Parse(config.URL)
	params := url.Query()
	for k, v := range config.Data {
		value := fmt.Sprintf("%v", v)
		value = strings.ReplaceAll(value, "<USER>", combo.Username)
		value = strings.ReplaceAll(value, "<PASS>", combo.Password)
		params.Set(k, value)
	}
	url.RawQuery = params.Encode()
	return url.String()
}

// analyzeResponse determines the check result based on response content
func (c *Checker) analyzeResponse(body string, statusCode int, config types.Config) string {
	// Check for success strings
	for _, successStr := range config.SuccessStrings {
		if strings.Contains(body, successStr) {
			return "valid"
		}
	}

	// Check for failure strings
	for _, failureStr := range config.FailureStrings {
		if strings.Contains(body, failureStr) {
			return "invalid"
		}
	}

	// Default to invalid if no specific conditions match
	return "invalid"
}

